// Code generated manually for tests; based on gomock style.
package mocks

import (
	"context"
	"reflect"
	"saas-subscription-platform/services/auth-service/internal/client"

	"github.com/golang/mock/gomock"
)

// MockUserClient is a mock of service.UserClient.
type MockUserClient struct {
	ctrl     *gomock.Controller
	recorder *MockUserClientMockRecorder
}

// MockUserClientMockRecorder records invocations for MockUserClient.
type MockUserClientMockRecorder struct {
	mock *MockUserClient
}

// NewMockUserClient creates a new mock instance.
func NewMockUserClient(ctrl *gomock.Controller) *MockUserClient {
	mock := &MockUserClient{ctrl: ctrl}
	mock.recorder = &MockUserClientMockRecorder{mock}
	return mock
}

// EXPECT returns the recorder.
func (m *MockUserClient) EXPECT() *MockUserClientMockRecorder { return m.recorder }

// CreateUserWithContext mocks base method.
func (m *MockUserClient) CreateUserWithContext(ctx context.Context, email, password string, headers map[string]string) (client.CreateUserResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateUserWithContext", ctx, email, password, headers)
	ret0, _ := ret[0].(client.CreateUserResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// CreateUserWithContext indicates expected call.
func (mr *MockUserClientMockRecorder) CreateUserWithContext(ctx, email, password, headers interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateUserWithContext", reflect.TypeOf((*MockUserClient)(nil).CreateUserWithContext), ctx, email, password, headers)
}

// GetUserByEmailWithContext mocks base method.
func (m *MockUserClient) GetUserByEmailWithContext(ctx context.Context, email string, headers map[string]string) (client.GetUserByEmailResponse, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetUserByEmailWithContext", ctx, email, headers)
	ret0, _ := ret[0].(client.GetUserByEmailResponse)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// GetUserByEmailWithContext indicates expected call.
func (mr *MockUserClientMockRecorder) GetUserByEmailWithContext(ctx, email, headers interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetUserByEmailWithContext", reflect.TypeOf((*MockUserClient)(nil).GetUserByEmailWithContext), ctx, email, headers)
}
