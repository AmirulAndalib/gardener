// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/gardener/gardener/pkg/component/extensions/worker (interfaces: Interface)
//
// Generated by this command:
//
//	mockgen -package worker -destination=mocks.go github.com/gardener/gardener/pkg/component/extensions/worker Interface
//

// Package worker is a generated GoMock package.
package worker

import (
	context "context"
	reflect "reflect"

	v1beta1 "github.com/gardener/gardener/pkg/apis/core/v1beta1"
	v1alpha1 "github.com/gardener/gardener/pkg/apis/extensions/v1alpha1"
	operatingsystemconfig "github.com/gardener/gardener/pkg/component/extensions/operatingsystemconfig"
	gomock "go.uber.org/mock/gomock"
	runtime "k8s.io/apimachinery/pkg/runtime"
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

// Get mocks base method.
func (m *MockInterface) Get(arg0 context.Context) (*v1alpha1.Worker, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Get", arg0)
	ret0, _ := ret[0].(*v1alpha1.Worker)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Get indicates an expected call of Get.
func (mr *MockInterfaceMockRecorder) Get(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Get", reflect.TypeOf((*MockInterface)(nil).Get), arg0)
}

// MachineDeployments mocks base method.
func (m *MockInterface) MachineDeployments() []v1alpha1.MachineDeployment {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "MachineDeployments")
	ret0, _ := ret[0].([]v1alpha1.MachineDeployment)
	return ret0
}

// MachineDeployments indicates an expected call of MachineDeployments.
func (mr *MockInterfaceMockRecorder) MachineDeployments() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "MachineDeployments", reflect.TypeOf((*MockInterface)(nil).MachineDeployments))
}

// Migrate mocks base method.
func (m *MockInterface) Migrate(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Migrate", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// Migrate indicates an expected call of Migrate.
func (mr *MockInterfaceMockRecorder) Migrate(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Migrate", reflect.TypeOf((*MockInterface)(nil).Migrate), ctx)
}

// Restore mocks base method.
func (m *MockInterface) Restore(ctx context.Context, shootState *v1beta1.ShootState) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Restore", ctx, shootState)
	ret0, _ := ret[0].(error)
	return ret0
}

// Restore indicates an expected call of Restore.
func (mr *MockInterfaceMockRecorder) Restore(ctx, shootState any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Restore", reflect.TypeOf((*MockInterface)(nil).Restore), ctx, shootState)
}

// SetInfrastructureProviderStatus mocks base method.
func (m *MockInterface) SetInfrastructureProviderStatus(arg0 *runtime.RawExtension) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetInfrastructureProviderStatus", arg0)
}

// SetInfrastructureProviderStatus indicates an expected call of SetInfrastructureProviderStatus.
func (mr *MockInterfaceMockRecorder) SetInfrastructureProviderStatus(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetInfrastructureProviderStatus", reflect.TypeOf((*MockInterface)(nil).SetInfrastructureProviderStatus), arg0)
}

// SetSSHPublicKey mocks base method.
func (m *MockInterface) SetSSHPublicKey(arg0 []byte) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetSSHPublicKey", arg0)
}

// SetSSHPublicKey indicates an expected call of SetSSHPublicKey.
func (mr *MockInterfaceMockRecorder) SetSSHPublicKey(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetSSHPublicKey", reflect.TypeOf((*MockInterface)(nil).SetSSHPublicKey), arg0)
}

// SetWorkerPoolNameToOperatingSystemConfigsMap mocks base method.
func (m *MockInterface) SetWorkerPoolNameToOperatingSystemConfigsMap(arg0 map[string]*operatingsystemconfig.OperatingSystemConfigs) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "SetWorkerPoolNameToOperatingSystemConfigsMap", arg0)
}

// SetWorkerPoolNameToOperatingSystemConfigsMap indicates an expected call of SetWorkerPoolNameToOperatingSystemConfigsMap.
func (mr *MockInterfaceMockRecorder) SetWorkerPoolNameToOperatingSystemConfigsMap(arg0 any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SetWorkerPoolNameToOperatingSystemConfigsMap", reflect.TypeOf((*MockInterface)(nil).SetWorkerPoolNameToOperatingSystemConfigsMap), arg0)
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

// WaitMigrate mocks base method.
func (m *MockInterface) WaitMigrate(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitMigrate", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitMigrate indicates an expected call of WaitMigrate.
func (mr *MockInterfaceMockRecorder) WaitMigrate(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitMigrate", reflect.TypeOf((*MockInterface)(nil).WaitMigrate), ctx)
}

// WaitUntilWorkerStatusMachineDeploymentsUpdated mocks base method.
func (m *MockInterface) WaitUntilWorkerStatusMachineDeploymentsUpdated(ctx context.Context) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WaitUntilWorkerStatusMachineDeploymentsUpdated", ctx)
	ret0, _ := ret[0].(error)
	return ret0
}

// WaitUntilWorkerStatusMachineDeploymentsUpdated indicates an expected call of WaitUntilWorkerStatusMachineDeploymentsUpdated.
func (mr *MockInterfaceMockRecorder) WaitUntilWorkerStatusMachineDeploymentsUpdated(ctx any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WaitUntilWorkerStatusMachineDeploymentsUpdated", reflect.TypeOf((*MockInterface)(nil).WaitUntilWorkerStatusMachineDeploymentsUpdated), ctx)
}
