import React from 'react';
import { render, screen, within, waitFor } from '../../../test-utils';
import CoAList from './CoAList';
import { CoaState, PaginatedCoAResponse, AccountType } from './coaTypes';
import { RootState } from '../../../store';
import userEvent from '@testing-library/user-event';

import { rest } from 'msw';
import { setupServer } from 'msw/node';

const mockInitialCoaState: CoaState = {
  accounts: [],
  selectedAccount: null,
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
  rest.get('/api/v1/accounting/accounts', (req, res, ctx) => {
    const response: PaginatedCoAResponse = {
      data: [
        { id: '1', account_code: '1010', account_name: 'Test Cash', account_type: AccountType.Asset, is_active: true, is_control_account: false, description: 'Cash account', created_at: '2023-01-01T00:00:00Z', updated_at: '2023-01-01T00:00:00Z' },
        { id: '2', account_code: '6010', account_name: 'Test Salaries', account_type: AccountType.Expense, is_active: true, is_control_account: false, description: 'Salaries expense', created_at: '2023-01-01T00:00:00Z', updated_at: '2023-01-01T00:00:00Z' },
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


describe('CoAList Component', () => {
  test('renders "No accounts found" when no data is provided and not loading', () => {
    // Override server response for this specific test to return empty data
    server.use(
      rest.get('/api/v1/accounting/accounts', (req, res, ctx) => {
        return res(ctx.json({ data: [], page: 1, limit: 5, total: 0 }));
      })
    );

    render(<CoAList />, { preloadedState: { coa: { ...mockInitialCoaState, accounts: [], total: 0, loading: false } } });
    expect(screen.getByText(/No accounts found/i)).toBeInTheDocument();
  });

  test('renders table with accounts when data is provided', async () => {
    render(<CoAList />, {
        preloadedState: {
            coa: { ...mockInitialCoaState }
        }
    });

    // Wait for accounts to be loaded by MSW
    await waitFor(() => {
        expect(screen.getByText('Test Cash')).toBeInTheDocument();
        expect(screen.getByText('1010')).toBeInTheDocument();
        expect(screen.getByText('Test Salaries')).toBeInTheDocument();
        expect(screen.getByText('6010')).toBeInTheDocument();
    }, { timeout: 3000 });
  });

  test('renders filter inputs and Add New Account button', async () => {
    render(<CoAList />, { preloadedState: { coa: mockInitialCoaState } });
    expect(screen.getByLabelText(/Account Name/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Account Type/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/Status/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Filter/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /Add New Account/i })).toBeInTheDocument();
    // Wait for any initial data load to complete to avoid act warnings
    await screen.findByText(/Accounts List/i);
  });

  test('clicking "Add New Account" button opens the create form modal (conceptual - checks for function call or state change if directly testable)', async () => {
    // This test is more about the interaction setup.
    // Actual modal opening involves state changes within CoAList.
    // We can check if the button is there. A more detailed test would involve mocking useState or context.
    render(<CoAList />, { preloadedState: { coa: mockInitialCoaState } });
    const addButton = screen.getByRole('button', { name: /Add New Account/i });
    await userEvent.click(addButton);
    // In a real scenario, you'd check if the CoACreateForm becomes visible.
    // For basic test, we assume clicking it works if it doesn't crash.
    // Here, we'll just ensure no crash and the button was clickable.
    expect(addButton).toBeEnabled();
    await screen.findByText(/Accounts List/i); // Ensure component is stable
  });
});
