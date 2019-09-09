// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import proto "github.com/paysuper/paysuper-reporter/pkg/proto"

// ReportFileRepositoryInterface is an autogenerated mock type for the ReportFileRepositoryInterface type
type ReportFileRepositoryInterface struct {
	mock.Mock
}

// GetById provides a mock function with given fields: _a0
func (_m *ReportFileRepositoryInterface) GetById(_a0 string) (*proto.MgoReportFile, error) {
	ret := _m.Called(_a0)

	var r0 *proto.MgoReportFile
	if rf, ok := ret.Get(0).(func(string) *proto.MgoReportFile); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*proto.MgoReportFile)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Insert provides a mock function with given fields: _a0
func (_m *ReportFileRepositoryInterface) Insert(_a0 *proto.MgoReportFile) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*proto.MgoReportFile) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: _a0
func (_m *ReportFileRepositoryInterface) Update(_a0 *proto.MgoReportFile) error {
	ret := _m.Called(_a0)

	var r0 error
	if rf, ok := ret.Get(0).(func(*proto.MgoReportFile) error); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
