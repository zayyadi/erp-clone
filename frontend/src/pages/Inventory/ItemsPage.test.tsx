import React from 'react';
import { render, screen, waitFor } from '../../test-utils'; // Use custom render
import ItemsPage from './ItemsPage';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { PaginatedItemResponse, ItemType, ValuationMethod } from '../../features/inventory/items/itemTypes';

// Mock API calls for ItemList
const server = setupServer(
  rest.get('/api/v1/inventory/items', (req, res, ctx) => {
    const response: PaginatedItemResponse = {
      data: [
        { id: 'item1', sku: 'SKU001', name: 'Test Item 1', item_type: ItemType.FinishedGood, unit_of_measure: 'pcs', purchase_price: 10, sales_price: 20, valuation_method: ValuationMethod.FIFO, is_active: true, created_at: '', updated_at: '' },
        { id: 'item2', sku: 'SKU002', name: 'Test Item 2', item_type: ItemType.RawMaterial, unit_of_measure: 'kg', purchase_price: 5, sales_price: null, valuation_method: ValuationMethod.WeightedAverage, is_active: true, created_at: '', updated_at: '' },
      ],
      page: 1,
      limit: 20,
      total: 2,
    };
    return res(ctx.json(response));
  })
);

beforeAll(() => server.listen());
afterEach(() => server.resetHandlers());
afterAll(() => server.close());

describe('ItemsPage', () => {
  test('renders page title and ItemList component (which shows loading/data)', async () => {
    render(<ItemsPage />);

    expect(screen.getByRole('heading', { name: /Inventory Items/i, level: 4 })).toBeInTheDocument();

    // ItemList will initially show loading or table content.
    expect(await screen.findByText(/Item Master List/i, {}, { timeout: 3000 })).toBeInTheDocument();

    // Check if mocked data appears
    await waitFor(() => {
      expect(screen.getByText('Test Item 1')).toBeInTheDocument();
      expect(screen.getByText('SKU002')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  test('ItemList shows add new item button', async () => {
    render(<ItemsPage />);
    expect(await screen.findByRole('button', { name: /Add New Item/i })).toBeInTheDocument();
  });
});
