package mocks

import (
	"context"
	"erp-system/internal/inventory/models"
	"erp-system/internal/inventory/repository"

	"github.com/google/uuid"
	"github.com/stretchr/testify/mock"
)

// WarehouseRepository is an autogenerated mock type for the WarehouseRepository type
type WarehouseRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, warehouse
func (_m *WarehouseRepository) Create(ctx context.Context, warehouse *models.Warehouse) (*models.Warehouse, error) {
	ret := _m.Called(ctx, warehouse)

	var r0 *models.Warehouse
	if rf, ok := ret.Get(0).(func(context.Context, *models.Warehouse) *models.Warehouse); ok {
		r0 = rf(ctx, warehouse)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Warehouse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *models.Warehouse) error); ok {
		r1 = rf(ctx, warehouse)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *WarehouseRepository) Delete(ctx context.Context, id uuid.UUID) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByCode provides a mock function with given fields: ctx, code
func (_m *WarehouseRepository) GetByCode(ctx context.Context, code string) (*models.Warehouse, error) {
	ret := _m.Called(ctx, code)

	var r0 *models.Warehouse
	if rf, ok := ret.Get(0).(func(context.Context, string) *models.Warehouse); ok {
		r0 = rf(ctx, code)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Warehouse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, code)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *WarehouseRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Warehouse, error) {
	ret := _m.Called(ctx, id)

	var r0 *models.Warehouse
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID) *models.Warehouse); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Warehouse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, offset, limit, filters
func (_m *WarehouseRepository) List(ctx context.Context, offset int, limit int, filters map[string]interface{}) ([]*models.Warehouse, int64, error) {
	ret := _m.Called(ctx, offset, limit, filters)

	var r0 []*models.Warehouse
	if rf, ok := ret.Get(0).(func(context.Context, int, int, map[string]interface{}) []*models.Warehouse); ok {
		r0 = rf(ctx, offset, limit, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*models.Warehouse)
		}
	}

	var r1 int64
	if rf, ok := ret.Get(1).(func(context.Context, int, int, map[string]interface{}) int64); ok {
		r1 = rf(ctx, offset, limit, filters)
	} else {
		r1 = ret.Get(1).(int64)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, int, int, map[string]interface{}) error); ok {
		r2 = rf(ctx, offset, limit, filters)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// Update provides a mock function with given fields: ctx, warehouse
func (_m *WarehouseRepository) Update(ctx context.Context, warehouse *models.Warehouse) (*models.Warehouse, error) {
	ret := _m.Called(ctx, warehouse)

	var r0 *models.Warehouse
	if rf, ok := ret.Get(0).(func(context.Context, *models.Warehouse) *models.Warehouse); ok {
		r0 = rf(ctx, warehouse)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*models.Warehouse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *models.Warehouse) error); ok {
		r1 = rf(ctx, warehouse)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}


// NewWarehouseRepository creates a new instance of WarehouseRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWarehouseRepositoryMock(t interface {
	mock.TestingT
	Cleanup(func())
}) *WarehouseRepository {
	mock := &WarehouseRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}

var _ repository.WarehouseRepository = (*WarehouseRepository)(nil)
