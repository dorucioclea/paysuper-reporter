// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import proto "github.com/paysuper/paysuper-reporter/pkg/proto"

// DocumentGeneratorInterface is an autogenerated mock type for the DocumentGeneratorInterface type
type DocumentGeneratorInterface struct {
	mock.Mock
}

// Render provides a mock function with given fields: payload
func (_m *DocumentGeneratorInterface) Render(payload *proto.GeneratorPayload) (*proto.File, error) {
	ret := _m.Called(payload)

	var r0 *proto.File
	if rf, ok := ret.Get(0).(func(*proto.GeneratorPayload) *proto.File); ok {
		r0 = rf(payload)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*proto.File)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*proto.GeneratorPayload) error); ok {
		r1 = rf(payload)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
