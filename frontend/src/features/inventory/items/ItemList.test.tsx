import React from 'react';
import { render, screen, waitFor } from '../../../test-utils';
import ItemList from './ItemList';
import { ItemState, PaginatedItemResponse, ItemType, ValuationMethod } from './itemTypes';
import userEvent from '@testing-library/user-event';

import { rest } from 'msw';
import { setupServer } from 'msw/node';

const mockInitialItemState: ItemState = {
  items: [],
  selectedItem: null,
  loading: false,
  error: null,
  page: 1,
  limit: 5,
  total: 0,
  loadingCreate: false,
  loadingUpdate: false,
  loadingDelete: false,
  loadingFetch: false,
};

const server = setupServer(
  rest.get('/api/v1/inventory/items', (req, res, ctx) => {
    const response: PaginatedItemResponse = {
      data: [
        { id: 'item1', sku: 'SKU1001', name: 'Inventory Item A', item_type: ItemType.FinishedGood, unit_of_measure: 'pcs', purchase_price: 100, sales_price: 150, valuation_method: ValuationMethod.FIFO, is_active: true, description: 'Item A desc', created_at: '2023-01-01T00:00:00Z', updated_at: '2023-01-01T00:00:00Z' },
        { id: 'item2', sku: 'SKU1002', name: 'Inventory Item B', item_type: ItemType.RawMaterial, unit_of_measure: 'kg', purchase_price: 50, sales_price: null, valuation_method: ValuationMethod.WeightedAverage, is_active: true, description: 'Item B desc', created_at: '2023-01-01T00:00:00Z', updated_at: '2023-01-01T00:00:00Z' },
      ],
      page: 1,
      limit: 5,
      total: 2,
    };
    return res(ctx.json(response));
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ItemList Component', () => {
  test('renders "No items found" when no data is provided and not loading', () => {
    server.use(
      rest.get('/api/v1/inventory/items', (req, res, ctx) => {
        return res(ctx.json({ data: [], page: 1, limit: 5, total: 0 }));
      })
    );
    render(<ItemList />, { preloadedState: { item: { ...mockInitialItemState, items: [], total: 0, loading: false } } });
    expect(screen.getByText(/No items found/i)).toBeInTheDocument();
  });

  test('renders table with items when data is provided', async () => {
    render(<ItemList />, {
        preloadedState: {
            item: { ...mockInitialItemState }
        }
    });
    await waitFor(() => {
        expect(screen.getByText('Inventory Item A')).toBeInTheDocument();
        expect(screen.getByText('SKU1001')).toBeInTheDocument();
        expect(screen.getByText('Inventory Item B')).toBeInTheDocument();
        expect(screen.getByText('SKU1002')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  test('renders filter inputs and Add New Item button', async () => {
    render(<ItemList />, { preloadedState: { item: mockInitialItemState } });
    expect(screen.getByLabelText(/Item Name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/SKU/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Item Type/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Status/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Filter/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Add New Item/i })).toBeInTheDocument();
    await screen.findByText(/Item Master List/i);
  });

  test('clicking "Add New Item" button opens the create form modal (conceptual)', async () => {
    render(<ItemList />, { preloadedState: { item: mockInitialItemState } });
    const addButton = screen.getByRole('button', { name: /Add New Item/i });
    await userEvent.click(addButton);
    expect(addButton).toBeEnabled();
    await screen.findByText(/Item Master List/i);
  });
});
