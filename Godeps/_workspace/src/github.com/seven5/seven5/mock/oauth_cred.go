// Automatically generated by MockGen. DO NOT EDIT!
// Source: github.com/seven5/seven5 (interfaces: OauthCred)

package mock

import (
	gomock "code.google.com/p/gomock/gomock"
)

// Mock of OauthCred interface
type MockOauthCred struct {
	ctrl     *gomock.Controller
	recorder *_MockOauthCredRecorder
}

// Recorder for MockOauthCred (not exported)
type _MockOauthCredRecorder struct {
	mock *MockOauthCred
}

func NewMockOauthCred(ctrl *gomock.Controller) *MockOauthCred {
	mock := &MockOauthCred{ctrl: ctrl}
	mock.recorder = &_MockOauthCredRecorder{mock}
	return mock
}

func (_m *MockOauthCred) EXPECT() *_MockOauthCredRecorder {
	return _m.recorder
}

func (_m *MockOauthCred) Secret() string {
	ret := _m.ctrl.Call(_m, "Secret")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockOauthCredRecorder) Secret() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Secret")
}

func (_m *MockOauthCred) Token() string {
	ret := _m.ctrl.Call(_m, "Token")
	ret0, _ := ret[0].(string)
	return ret0
}

func (_mr *_MockOauthCredRecorder) Token() *gomock.Call {
	return _mr.mock.ctrl.RecordCall(_mr.mock, "Token")
}