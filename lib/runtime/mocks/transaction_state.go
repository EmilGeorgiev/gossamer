// Code generated by mockery v2.9.4. DO NOT EDIT.

package mocks

import (
	common "github.com/ChainSafe/gossamer/lib/common"
	mock "github.com/stretchr/testify/mock"

	transaction "github.com/ChainSafe/gossamer/lib/transaction"
)

// TransactionState is an autogenerated mock type for the TransactionState type
type TransactionState struct {
	mock.Mock
}

// AddToPool provides a mock function with given fields: vt
func (_m *TransactionState) AddToPool(vt *transaction.ValidTransaction) common.Hash {
	ret := _m.Called(vt)

	var r0 common.Hash
	if rf, ok := ret.Get(0).(func(*transaction.ValidTransaction) common.Hash); ok {
		r0 = rf(vt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(common.Hash)
		}
	}

	return r0
}