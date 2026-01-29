// Code generated manually for tests; gomock-style mock for service.InvoiceStore.
package mocks

import (
	"reflect"

	"saas-subscription-platform/services/billing-service/internal/model"
	"saas-subscription-platform/services/billing-service/internal/repository"

	"github.com/golang/mock/gomock"
)

type MockInvoiceStore struct {
	ctrl     *gomock.Controller
	recorder *MockInvoiceStoreMockRecorder
}

type MockInvoiceStoreMockRecorder struct {
	mock *MockInvoiceStore
}

func NewMockInvoiceStore(ctrl *gomock.Controller) *MockInvoiceStore {
	mock := &MockInvoiceStore{ctrl: ctrl}
	mock.recorder = &MockInvoiceStoreMockRecorder{mock}
	return mock
}

func (m *MockInvoiceStore) EXPECT() *MockInvoiceStoreMockRecorder { return m.recorder }

func (m *MockInvoiceStore) CreateInvoice(invoice *model.Invoice) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "CreateInvoice", invoice)
	ret0, _ := ret[0].(error)
	return ret0
}

func (mr *MockInvoiceStoreMockRecorder) CreateInvoice(invoice interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CreateInvoice", reflect.TypeOf((*MockInvoiceStore)(nil).CreateInvoice), invoice)
}

func (m *MockInvoiceStore) GetInvoiceByID(userID string, id int) (*model.Invoice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInvoiceByID", userID, id)
	ret0, _ := ret[0].(*model.Invoice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInvoiceStoreMockRecorder) GetInvoiceByID(userID, id interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInvoiceByID", reflect.TypeOf((*MockInvoiceStore)(nil).GetInvoiceByID), userID, id)
}

func (m *MockInvoiceStore) GetInvoices(filter repository.InvoiceFilter) ([]*model.Invoice, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetInvoices", filter)
	ret0, _ := ret[0].([]*model.Invoice)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

func (mr *MockInvoiceStoreMockRecorder) GetInvoices(filter interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetInvoices", reflect.TypeOf((*MockInvoiceStore)(nil).GetInvoices), filter)
}
