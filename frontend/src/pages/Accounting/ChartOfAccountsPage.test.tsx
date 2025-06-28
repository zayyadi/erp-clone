import React from 'react';
import { render, screen, waitFor } from '../../test-utils'; // Use custom render
import ChartOfAccountsPage from './ChartOfAccountsPage';
import { rest } from 'msw';
import { setupServer } from 'msw/node';
import { PaginatedCoAResponse } from '../../features/accounting/coa/coaTypes';

// Mock API calls for CoAList
const server = setupServer(
  rest.get('/api/v1/accounting/accounts', (req, res, ctx) => {
    const response: PaginatedCoAResponse = {
      data: [
        { id: '1', account_code: '101', account_name: 'Cash', account_type: 'Asset', is_active: true, is_control_account: false, created_at: '', updated_at: '' },
        { id: '2', account_code: '201', account_name: 'Accounts Payable', account_type: 'Liability', is_active: true, is_control_account: false, created_at: '', updated_at: '' },
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

describe('ChartOfAccountsPage', () => {
  test('renders page title and CoAList component (which shows loading/data)', async () => {
    render(<ChartOfAccountsPage />);

    expect(screen.getByRole('heading', { name: /Chart of Accounts/i, level: 4 })).toBeInTheDocument();

    // CoAList will initially show loading or table content.
    // Wait for an element that appears after data loading (e.g., one of the mocked accounts)
    // Or at least the "Accounts List" title from within CoAList
    expect(await screen.findByText(/Accounts List/i, {}, { timeout: 3000 })).toBeInTheDocument();

    // Check if mocked data appears
    await waitFor(() => {
      expect(screen.getByText('Cash')).toBeInTheDocument();
      expect(screen.getByText('Accounts Payable')).toBeInTheDocument();
    }, { timeout: 3000 });

  });

  test('CoAList shows add new account button', async () => {
    render(<ChartOfAccountsPage />);
    // Wait for the list component to be fully rendered
    expect(await screen.findByRole('button', { name: /Add New Account/i })).toBeInTheDocument();
  });
});
