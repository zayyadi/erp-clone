import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import axios from 'axios';
import {
  CoaState,
  ChartOfAccount,
  CreateChartOfAccountRequest,
  UpdateChartOfAccountRequest,
  ListChartOfAccountsRequest,
  PaginatedCoAResponse,
} from './coaTypes';
import { RootState, AppDispatch } from '../../../store'; // Added AppDispatch
import { showNotification } from '../../notifications/notificationSlice';

const API_BASE_URL = '/api/v1/accounting/accounts';

// Initial state
const initialState: CoaState = {
  accounts: [],
  selectedAccount: null,
  loading: false,
  error: null,
  page: 1,
  limit: 20,
  total: 0,
  loadingCreate: false,
  loadingUpdate: false,
  loadingDelete: false,
  loadingFetch: false,
};

// Async Thunks
export const fetchCoAs = createAsyncThunk<PaginatedCoAResponse, ListChartOfAccountsRequest, { rejectValue: string }>(
  'coa/fetchCoAs',
  async (params, { rejectWithValue }) => {
    try {
      const response = await axios.get<PaginatedCoAResponse>(API_BASE_URL, { params });
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || error.message || 'Failed to fetch chart of accounts');
    }
  }
);

export const fetchCoABById = createAsyncThunk<ChartOfAccount, string, { rejectValue: string }>(
  'coa/fetchCoABById',
  async (id, { rejectWithValue }) => {
    try {
      const response = await axios.get<ChartOfAccount>(`${API_BASE_URL}/${id}`);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || error.message || 'Failed to fetch account');
    }
  }
);

export const fetchCoABByCode = createAsyncThunk<ChartOfAccount, string, { rejectValue: string }>(
  'coa/fetchCoABByCode',
  async (code, { rejectWithValue }) => {
    try {
      const response = await axios.get<ChartOfAccount>(`${API_BASE_URL}/code/${code}`);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || error.message || 'Failed to fetch account by code');
    }
  }
);

export const createCoA = createAsyncThunk<ChartOfAccount, CreateChartOfAccountRequest, { dispatch: AppDispatch, rejectValue: string }>(
  'coa/createCoA',
  async (coaData, { dispatch, rejectWithValue }) => {
    try {
      const response = await axios.post<ChartOfAccount>(API_BASE_URL, coaData);
      dispatch(showNotification({ message: 'Account created successfully', severity: 'success' }));
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to create account';
      dispatch(showNotification({ message: errorMessage, severity: 'error' }));
      return rejectWithValue(errorMessage);
    }
  }
);

export const updateCoA = createAsyncThunk<ChartOfAccount, { id: string; coaData: UpdateChartOfAccountRequest }, { dispatch: AppDispatch, rejectValue: string }>(
  'coa/updateCoA',
  async ({ id, coaData }, { dispatch, rejectWithValue }) => {
    try {
      const response = await axios.put<ChartOfAccount>(`${API_BASE_URL}/${id}`, coaData);
      dispatch(showNotification({ message: 'Account updated successfully', severity: 'success' }));
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to update account';
      dispatch(showNotification({ message: errorMessage, severity: 'error' }));
      return rejectWithValue(errorMessage);
    }
  }
);

export const deleteCoA = createAsyncThunk<{ id: string }, string, { dispatch: AppDispatch, rejectValue: string }>(
  'coa/deleteCoA',
  async (id, { dispatch, rejectWithValue }) => {
    try {
      await axios.delete(`${API_BASE_URL}/${id}`);
      dispatch(showNotification({ message: 'Account deleted successfully', severity: 'success' }));
      return { id };
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to delete account';
      dispatch(showNotification({ message: errorMessage, severity: 'error' }));
      return rejectWithValue(errorMessage);
    }
  }
);

// Slice
const coaSlice = createSlice({
  name: 'coa',
  initialState,
  reducers: {
    resetCoAErrors: (state) => {
      state.error = null;
    },
    setSelectedAccount: (state, action: PayloadAction<ChartOfAccount | null>) => {
      state.selectedAccount = action.payload;
    },
    setCurrentPage: (state, action: PayloadAction<number>) => {
        state.page = action.payload;
    },
    setPageLimit: (state, action: PayloadAction<number>) => {
        state.limit = action.payload;
    }
  },
  extraReducers: (builder) => {
    // fetchCoAs
    builder
      .addCase(fetchCoAs.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchCoAs.fulfilled, (state, action: PayloadAction<PaginatedCoAResponse>) => {
        state.loading = false;
        state.accounts = action.payload.data;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
        state.total = action.payload.total;
      })
      .addCase(fetchCoAs.rejected, (state, action) => {
        state.loading = false;
        state.error = action.rejectWithValue as string;
      });

    // fetchCoABById
    builder
      .addCase(fetchCoABById.pending, (state) => {
        state.loadingFetch = true;
        state.error = null;
        state.selectedAccount = null;
      })
      .addCase(fetchCoABById.fulfilled, (state, action: PayloadAction<ChartOfAccount>) => {
        state.loadingFetch = false;
        state.selectedAccount = action.payload;
      })
      .addCase(fetchCoABById.rejected, (state, action) => {
        state.loadingFetch = false;
        state.error = action.rejectWithValue as string;
      });

    // fetchCoABByCode
    builder
      .addCase(fetchCoABByCode.pending, (state) => {
        state.loadingFetch = true;
        state.error = null;
        state.selectedAccount = null;
      })
      .addCase(fetchCoABByCode.fulfilled, (state, action: PayloadAction<ChartOfAccount>) => {
        state.loadingFetch = false;
        state.selectedAccount = action.payload;
      })
      .addCase(fetchCoABByCode.rejected, (state, action) => {
        state.loadingFetch = false;
        state.error = action.rejectWithValue as string;
      });

    // createCoA
    builder
      .addCase(createCoA.pending, (state) => {
        state.loadingCreate = true;
        state.error = null;
      })
      .addCase(createCoA.fulfilled, (state, action: PayloadAction<ChartOfAccount>) => {
        state.loadingCreate = false;
        // Optionally add to list or refetch:
        // state.accounts.unshift(action.payload);
        // For now, assume refetch or navigation will handle list update
      })
      .addCase(createCoA.rejected, (state, action) => {
        state.loadingCreate = false;
        state.error = action.rejectWithValue as string;
      });

    // updateCoA
    builder
      .addCase(updateCoA.pending, (state) => {
        state.loadingUpdate = true;
        state.error = null;
      })
      .addCase(updateCoA.fulfilled, (state, action: PayloadAction<ChartOfAccount>) => {
        state.loadingUpdate = false;
        const index = state.accounts.findIndex(acc => acc.id === action.payload.id);
        if (index !== -1) {
          state.accounts[index] = action.payload;
        }
        if (state.selectedAccount?.id === action.payload.id) {
          state.selectedAccount = action.payload;
        }
      })
      .addCase(updateCoA.rejected, (state, action) => {
        state.loadingUpdate = false;
        state.error = action.rejectWithValue as string;
      });

    // deleteCoA
    builder
      .addCase(deleteCoA.pending, (state) => {
        state.loadingDelete = true;
        state.error = null;
      })
      .addCase(deleteCoA.fulfilled, (state, action: PayloadAction<{ id: string }>) => {
        state.loadingDelete = false;
        state.accounts = state.accounts.filter(acc => acc.id !== action.payload.id);
        if (state.selectedAccount?.id === action.payload.id) {
          state.selectedAccount = null;
        }
        // Adjust total count if needed, or refetch
        state.total = Math.max(0, state.total -1);
      })
      .addCase(deleteCoA.rejected, (state, action) => {
        state.loadingDelete = false;
        state.error = action.rejectWithValue as string;
      });
  },
});

export const { resetCoAErrors, setSelectedAccount, setCurrentPage, setPageLimit } = coaSlice.actions;

// Selectors
export const selectCoAList = (state: RootState) => state.coa.accounts;
export const selectCoALoading = (state: RootState) => state.coa.loading;
export const selectCoAError = (state: RootState) => state.coa.error;
export const selectSelectedCoAccount = (state: RootState) => state.coa.selectedAccount;
export const selectCoAPagination = (state: RootState) => ({
    page: state.coa.page,
    limit: state.coa.limit,
    total: state.coa.total,
});
export const selectCoALoadingCreate = (state: RootState) => state.coa.loadingCreate;
export const selectCoALoadingUpdate = (state: RootState) => state.coa.loadingUpdate;
export const selectCoALoadingDelete = (state: RootState) => state.coa.loadingDelete;
export const selectCoALoadingFetch = (state: RootState) => state.coa.loadingFetch;


export default coaSlice.reducer;
