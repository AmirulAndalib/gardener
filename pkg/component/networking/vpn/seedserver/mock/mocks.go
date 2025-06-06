// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener/pkg/component/networking/vpn/seedserver (interfaces: Interface)
//
// Generated by this command:
//
//	mockgen -package mock -destination=mocks.go github.com/gardener/gardener/pkg/component/networking/vpn/seedserver Interface
//

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	net "net"
	reflect "reflect"

	seedserver "github.com/gardener/gardener/pkg/component/networking/vpn/seedserver"
	gomock "go.uber.org/mock/gomock"
)

// MockInterface is a mock of Interface interface.
type MockInterface struct {
	ctrl     *gomock.Controller
	recorder *MockInterfaceMockRecorder
	isgomock struct{}
}

// MockInterfaceMockRecorder is the mock recorder for MockInterface.
type MockInterfaceMockRecorder struct {
	mock *MockInterface
}

// NewMockInterface creates a new mock instance.
func NewMockInterface(ctrl *gomock.Controller) *MockInterface {
	mock := &MockInterface{ctrl: ctrl}
	mock.recorder = &MockInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockInterface) EXPECT() *MockInterfaceMockRecorder {
	return m.recorder
}

// Deploy mocks base method.
func (m *MockInterface) Deploy(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Deploy", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Deploy indicates an expected call of Deploy.
func (mr *MockInterfaceMockRecorder) Deploy(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Deploy", reflect.TypeOf((*MockInterface)(nil).Deploy), ctx)
}

// Destroy mocks base method.
func (m *MockInterface) Destroy(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Destroy", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Destroy indicates an expected call of Destroy.
func (mr *MockInterfaceMockRecorder) Destroy(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Destroy", reflect.TypeOf((*MockInterface)(nil).Destroy), ctx)
}

// GetValues mocks base method.
func (m *MockInterface) GetValues() seedserver.Values {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetValues")
	ret0, _ := ret[0].(seedserver.Values)
	return ret0
}

// GetValues indicates an expected call of GetValues.
func (mr *MockInterfaceMockRecorder) GetValues() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetValues", reflect.TypeOf((*MockInterface)(nil).GetValues))
}

// SetNodeNetworkCIDRs mocks base method.
func (m *MockInterface) SetNodeNetworkCIDRs(nodes []net.IPNet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetNodeNetworkCIDRs", nodes)
}

// SetNodeNetworkCIDRs indicates an expected call of SetNodeNetworkCIDRs.
func (mr *MockInterfaceMockRecorder) SetNodeNetworkCIDRs(nodes any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetNodeNetworkCIDRs", reflect.TypeOf((*MockInterface)(nil).SetNodeNetworkCIDRs), nodes)
}

// SetPodNetworkCIDRs mocks base method.
func (m *MockInterface) SetPodNetworkCIDRs(pods []net.IPNet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetPodNetworkCIDRs", pods)
}

// SetPodNetworkCIDRs indicates an expected call of SetPodNetworkCIDRs.
func (mr *MockInterfaceMockRecorder) SetPodNetworkCIDRs(pods any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetPodNetworkCIDRs", reflect.TypeOf((*MockInterface)(nil).SetPodNetworkCIDRs), pods)
}

// SetServiceNetworkCIDRs mocks base method.
func (m *MockInterface) SetServiceNetworkCIDRs(services []net.IPNet) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetServiceNetworkCIDRs", services)
}

// SetServiceNetworkCIDRs indicates an expected call of SetServiceNetworkCIDRs.
func (mr *MockInterfaceMockRecorder) SetServiceNetworkCIDRs(services any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetServiceNetworkCIDRs", reflect.TypeOf((*MockInterface)(nil).SetServiceNetworkCIDRs), services)
}

// Wait mocks base method.
func (m *MockInterface) Wait(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Wait", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Wait indicates an expected call of Wait.
func (mr *MockInterfaceMockRecorder) Wait(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Wait", reflect.TypeOf((*MockInterface)(nil).Wait), ctx)
}

// WaitCleanup mocks base method.
func (m *MockInterface) WaitCleanup(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitCleanup", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitCleanup indicates an expected call of WaitCleanup.
func (mr *MockInterfaceMockRecorder) WaitCleanup(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitCleanup", reflect.TypeOf((*MockInterface)(nil).WaitCleanup), ctx)
}
