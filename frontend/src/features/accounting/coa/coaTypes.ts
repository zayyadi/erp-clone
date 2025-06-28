// Based on backend models and DTOs (internal/accounting/models & service/dto)

export enum AccountType {
  Asset = "Asset",
  Liability = "Liability",
  Equity = "Equity",
  Revenue = "Revenue",
  Expense = "Expense",
  ContraAsset = "Contra Asset",
  ContraLiability = "Contra Liability",
  ContraEquity = "Contra Equity",
  ContraRevenue = "Contra Revenue",
  ContraExpense = "Contra Expense",
}

export interface ChartOfAccount {
  id: string;
  account_code: string;
  account_name: string;
  account_type: AccountType;
  description?: string;
  is_active: boolean;
  is_control_account: boolean;
  parent_account_id?: string | null;
  created_at: string;
  updated_at: string;
  // children?: ChartOfAccount[]; // For hierarchical display, if needed later
  // balance?: number; // If reports include balances directly
}

export interface CreateChartOfAccountRequest {
  account_code: string;
  account_name: string;
  account_type: AccountType;
  description?: string;
  is_active?: boolean; // Defaults to true on backend
  is_control_account?: boolean; // Defaults to false
  parent_account_id?: string | null;
}

export interface UpdateChartOfAccountRequest {
  account_name?: string;
  account_type?: AccountType;
  description?: string;
  is_active?: boolean;
  is_control_account?: boolean;
  parent_account_id?: string | null;
}

export interface ListChartOfAccountsRequest {
  page?: number;
  limit?: number;
  account_name?: string;
  account_type?: AccountType | string; // string for query param flexibility
  is_active?: boolean | string; // string for query param flexibility
}

export interface PaginatedCoAResponse {
  data: ChartOfAccount[];
  page: number;
  limit: number;
  total: number;
}

// For Redux state
export interface CoaState {
  accounts: ChartOfAccount[];
  selectedAccount: ChartOfAccount | null;
  loading: boolean;
  error: string | null | undefined; // Allow for undefined if using RTK error handling
  page: number;
  limit: number;
  total: number;
  // For specific operations
  loadingCreate: boolean;
  loadingUpdate: boolean;
  loadingDelete: boolean;
  loadingFetch: boolean; // For fetching single account
}
