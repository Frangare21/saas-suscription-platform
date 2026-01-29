package service

import (
	"errors"
	"testing"
	"time"

	"saas-subscription-platform/services/billing-service/internal/model"
	"saas-subscription-platform/services/billing-service/internal/repository"
	"saas-subscription-platform/services/billing-service/internal/service/mocks"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/require"
)

func TestBillingService_CreateInvoice(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockInvoiceStore(ctrl)
	svc := NewBillingService(store)

	store.EXPECT().CreateInvoice(gomock.Any()).DoAndReturn(func(inv *model.Invoice) error {
		inv.ID = 42
		return nil
	})

	inv, err := svc.CreateInvoice("user-1", 1500, "USD")
	require.NoError(t, err)
	require.Equal(t, 42, inv.ID)
	require.Equal(t, int64(1500), inv.AmountCents)
	require.Equal(t, "pending", inv.Status)
	require.WithinDuration(t, time.Now(), inv.CreatedAt, time.Second)
}

func TestBillingService_CreateInvoice_Validation(t *testing.T) {
	svc := NewBillingService(nil)

	_, err := svc.CreateInvoice("", 100, "USD")
	require.Error(t, err)

	_, err = svc.CreateInvoice("user-1", 0, "USD")
	require.Error(t, err)
}

func TestBillingService_GetInvoiceByID(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockInvoiceStore(ctrl)
	svc := NewBillingService(store)

	expected := &model.Invoice{ID: 1, UserID: "user-1"}
	store.EXPECT().GetInvoiceByID("user-1", 1).Return(expected, nil)

	inv, err := svc.GetInvoiceByID("user-1", 1)
	require.NoError(t, err)
	require.Equal(t, expected, inv)

	_, err = svc.GetInvoiceByID("", 1)
	require.Error(t, err)
}

func TestBillingService_ListInvoices(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockInvoiceStore(ctrl)
	svc := NewBillingService(store)

	expected := []*model.Invoice{{ID: 1}, {ID: 2}}
	store.EXPECT().GetInvoices(repository.InvoiceFilter{UserID: "user-1", Status: "paid", Limit: 10, Offset: 5}).Return(expected, nil)

	invoices, err := svc.ListInvoices("user-1", "paid", 10, 5)
	require.NoError(t, err)
	require.Equal(t, expected, invoices)

	_, err = svc.ListInvoices("", "paid", 10, 5)
	require.Error(t, err)
}

func TestBillingService_CreateInvoice_RepoError(t *testing.T) {
	ctrl := gomock.NewController(t)
	store := mocks.NewMockInvoiceStore(ctrl)
	svc := NewBillingService(store)

	store.EXPECT().CreateInvoice(gomock.Any()).Return(errors.New("db down"))

	_, err := svc.CreateInvoice("user-1", 100, "USD")
	require.Error(t, err)
}
