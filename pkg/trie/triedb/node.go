// Copyright 2024 ChainSafe Systems (ON)
// SPDX-License-Identifier: LGPL-3.0-only

package triedb

import (
	"bytes"

	"github.com/ChainSafe/gossamer/lib/common"
	"github.com/ChainSafe/gossamer/pkg/trie/db"
	"github.com/ChainSafe/gossamer/pkg/trie/triedb/codec"
)

type nodeValue interface {
	getHash() common.Hash
	equal(other nodeValue) bool
}

type (
	inline struct {
		Data []byte
	}

	valueRef struct {
		hash common.Hash
	}

	newValueRef struct {
		hash *common.Hash
		Data []byte
	}
)

func (inline) getHash() common.Hash { return common.EmptyHash }
func (n inline) equal(other nodeValue) bool {
	switch otherValue := other.(type) {
	case inline:
		return bytes.Equal(n.Data, otherValue.Data)
	default:
		return false
	}
}
func (vr valueRef) getHash() common.Hash { return vr.hash }
func (vr valueRef) equal(other nodeValue) bool {
	switch otherValue := other.(type) {
	case valueRef:
		return vr.hash == otherValue.hash
	default:
		return false
	}
}
func (vr newValueRef) getHash() common.Hash {
	if vr.hash == nil {
		return common.EmptyHash
	}

	return *vr.hash
}
func (vr newValueRef) equal(other nodeValue) bool {
	switch otherValue := other.(type) {
	case newValueRef:
		if vr.hash != nil && otherValue.hash != nil {
			return *vr.hash == *otherValue.hash
		}
		return vr.hash == otherValue.hash
	default:
		return false
	}
}

func NewValue(data []byte, threshold int) nodeValue {
	if len(data) >= threshold {
		return newValueRef{Data: data}
	}

	return inline{Data: data}
}

func NewFromEncoded(encodedValue codec.NodeValue) nodeValue {
	switch encoded := encodedValue.(type) {
	case codec.InlineValue:
		return inline{Data: encoded.Data}
	case codec.HashedValue:
		return valueRef{hash: common.NewHash(encoded.Data)}
	}

	return nil
}

func inMemoryFetchedValue(value nodeValue, prefix []byte, db db.DBGetter) ([]byte, error) {
	switch v := value.(type) {
	case inline:
		return v.Data, nil
	case newValueRef:
		return v.Data, nil
	case valueRef:
		prefixedKey := bytes.Join([][]byte{prefix, v.hash.ToBytes()}, nil)
		value, err := db.Get(prefixedKey)
		if err != nil {
			return nil, err
		}
		if value != nil {
			return value, nil
		}
		return value, ErrIncompleteDB
	default:
		panic("unreachable")
	}
}

type Node interface {
	getPartialKey() []byte
}

type (
	Empty struct{}
	Leaf  struct {
		partialKey []byte
		value      nodeValue
	}
	Branch struct {
		partialKey []byte
		children   [codec.ChildrenCapacity]NodeHandle
		value      nodeValue
	}
)

func (Empty) getPartialKey() []byte    { return nil }
func (n Leaf) getPartialKey() []byte   { return n.partialKey }
func (n Branch) getPartialKey() []byte { return n.partialKey }

// Create a new node from the encoded data, decoding this data into a codec.Node
// and mapping that with this node type
func newNodeFromEncoded(nodeHash common.Hash, data []byte, storage NodeStorage) (Node, error) {
	reader := bytes.NewReader(data)
	encodedNode, err := codec.Decode(reader)
	if err != nil {
		return nil, err
	}

	switch encoded := encodedNode.(type) {
	case codec.Empty:
		return Empty{}, nil
	case codec.Leaf:
		return Leaf{partialKey: encoded.PartialKey, value: NewFromEncoded(encoded.Value)}, nil
	case codec.Branch:
		key := encoded.PartialKey
		encodedChildren := encoded.Children
		value := encoded.Value

		child := func(i int) (NodeHandle, error) {
			if encodedChildren[i] != nil {
				newChild, err := newFromEncodedMerkleValue(nodeHash, encodedChildren[i], storage)
				if err != nil {
					return nil, err
				}
				return newChild, nil
			}
			return nil, nil //nolint:nilnil
		}

		children := [codec.ChildrenCapacity]NodeHandle{}
		for i := 0; i < len(children); i++ {
			child, err := child(i)
			if err != nil {
				return nil, err
			}
			children[i] = child
		}

		return Branch{partialKey: key, children: children, value: NewFromEncoded(value)}, nil
	default:
		panic("unreachable")
	}
}
