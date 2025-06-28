import { createSlice, createAsyncThunk, PayloadAction } from '@reduxjs/toolkit';
import axios from 'axios';
import {
  ItemState,
  Item,
  CreateItemRequest,
  UpdateItemRequest,
  ListItemRequest,
  PaginatedItemResponse,
} from './itemTypes';
import { RootState, AppDispatch } from '../../../store'; // Added AppDispatch
import { showNotification } from '../../notifications/notificationSlice'; // Added showNotification

const API_BASE_URL = '/api/v1/inventory/items';

const initialState: ItemState = {
  items: [],
  selectedItem: null,
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
export const fetchItems = createAsyncThunk<PaginatedItemResponse, ListItemRequest, { rejectValue: string }>(
  'items/fetchItems',
  async (params, { rejectWithValue }) => {
    try {
      const response = await axios.get<PaginatedItemResponse>(API_BASE_URL, { params });
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || error.message || 'Failed to fetch items');
    }
  }
);

export const fetchItemById = createAsyncThunk<Item, string, { rejectValue: string }>(
  'items/fetchItemById',
  async (id, { rejectWithValue }) => {
    try {
      const response = await axios.get<Item>(`${API_BASE_URL}/${id}`);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || error.message || 'Failed to fetch item');
    }
  }
);

export const fetchItemBySKU = createAsyncThunk<Item, string, { rejectValue: string }>(
  'items/fetchItemBySKU',
  async (sku, { rejectWithValue }) => {
    try {
      const response = await axios.get<Item>(`${API_BASE_URL}/sku/${sku}`);
      return response.data;
    } catch (error: any) {
      return rejectWithValue(error.response?.data?.message || error.message || 'Failed to fetch item by SKU');
    }
  }
);

export const createItem = createAsyncThunk<Item, CreateItemRequest, { dispatch: AppDispatch, rejectValue: string }>(
  'items/createItem',
  async (itemData, { dispatch, rejectWithValue }) => {
    try {
      const response = await axios.post<Item>(API_BASE_URL, itemData);
      dispatch(showNotification({ message: 'Item created successfully', severity: 'success' }));
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to create item';
      dispatch(showNotification({ message: errorMessage, severity: 'error' }));
      return rejectWithValue(errorMessage);
    }
  }
);

export const updateItem = createAsyncThunk<Item, { id: string; itemData: UpdateItemRequest }, { dispatch: AppDispatch, rejectValue: string }>(
  'items/updateItem',
  async ({ id, itemData }, { dispatch, rejectWithValue }) => {
    try {
      const response = await axios.put<Item>(`${API_BASE_URL}/${id}`, itemData);
      dispatch(showNotification({ message: 'Item updated successfully', severity: 'success' }));
      return response.data;
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to update item';
      dispatch(showNotification({ message: errorMessage, severity: 'error' }));
      return rejectWithValue(errorMessage);
    }
  }
);

export const deleteItem = createAsyncThunk<{ id: string }, string, { dispatch: AppDispatch, rejectValue: string }>(
  'items/deleteItem',
  async (id, { dispatch, rejectWithValue }) => {
    try {
      await axios.delete(`${API_BASE_URL}/${id}`);
      dispatch(showNotification({ message: 'Item deleted successfully', severity: 'success' }));
      return { id };
    } catch (error: any) {
      const errorMessage = error.response?.data?.message || error.message || 'Failed to delete item';
      dispatch(showNotification({ message: errorMessage, severity: 'error' }));
      return rejectWithValue(errorMessage);
    }
  }
);

// Slice
const itemSlice = createSlice({
  name: 'items',
  initialState,
  reducers: {
    resetItemErrors: (state) => {
      state.error = null;
    },
    setSelectedItem: (state, action: PayloadAction<Item | null>) => {
      state.selectedItem = action.payload;
    },
    setCurrentItemPage: (state, action: PayloadAction<number>) => {
        state.page = action.payload;
    },
    setItemPageLimit: (state, action: PayloadAction<number>) => {
        state.limit = action.payload;
    }
  },
  extraReducers: (builder) => {
    // fetchItems
    builder
      .addCase(fetchItems.pending, (state) => {
        state.loading = true;
        state.error = null;
      })
      .addCase(fetchItems.fulfilled, (state, action: PayloadAction<PaginatedItemResponse>) => {
        state.loading = false;
        state.items = action.payload.data;
        state.page = action.payload.page;
        state.limit = action.payload.limit;
        state.total = action.payload.total;
      })
      .addCase(fetchItems.rejected, (state, action) => {
        state.loading = false;
        state.error = action.rejectWithValue as string;
      });

    // fetchItemById / fetchItemBySKU (shared loading/error states for single fetch)
    [fetchItemById, fetchItemBySKU].forEach(thunk => {
        builder
            .addCase(thunk.pending, (state) => {
                state.loadingFetch = true;
                state.error = null;
                state.selectedItem = null;
            })
            .addCase(thunk.fulfilled, (state, action: PayloadAction<Item>) => {
                state.loadingFetch = false;
                state.selectedItem = action.payload;
            })
            .addCase(thunk.rejected, (state, action) => {
                state.loadingFetch = false;
                state.error = action.rejectWithValue as string;
            });
    });

    // createItem
    builder
      .addCase(createItem.pending, (state) => {
        state.loadingCreate = true;
        state.error = null;
      })
      .addCase(createItem.fulfilled, (state, action: PayloadAction<Item>) => {
        state.loadingCreate = false;
        // state.items.unshift(action.payload); // Or refetch
      })
      .addCase(createItem.rejected, (state, action) => {
        state.loadingCreate = false;
        state.error = action.rejectWithValue as string;
      });

    // updateItem
    builder
      .addCase(updateItem.pending, (state) => {
        state.loadingUpdate = true;
        state.error = null;
      })
      .addCase(updateItem.fulfilled, (state, action: PayloadAction<Item>) => {
        state.loadingUpdate = false;
        const index = state.items.findIndex(item => item.id === action.payload.id);
        if (index !== -1) {
          state.items[index] = action.payload;
        }
        if (state.selectedItem?.id === action.payload.id) {
          state.selectedItem = action.payload;
        }
      })
      .addCase(updateItem.rejected, (state, action) => {
        state.loadingUpdate = false;
        state.error = action.rejectWithValue as string;
      });

    // deleteItem
    builder
      .addCase(deleteItem.pending, (state) => {
        state.loadingDelete = true;
        state.error = null;
      })
      .addCase(deleteItem.fulfilled, (state, action: PayloadAction<{ id: string }>) => {
        state.loadingDelete = false;
        state.items = state.items.filter(item => item.id !== action.payload.id);
        if (state.selectedItem?.id === action.payload.id) {
          state.selectedItem = null;
        }
        state.total = Math.max(0, state.total -1);
      })
      .addCase(deleteItem.rejected, (state, action) => {
        state.loadingDelete = false;
        state.error = action.rejectWithValue as string;
      });
  },
});

export const { resetItemErrors, setSelectedItem, setCurrentItemPage, setItemPageLimit } = itemSlice.actions;

// Selectors
export const selectItemList = (state: RootState) => state.item.items; // Ensure store is updated with 'item' key
export const selectItemLoading = (state: RootState) => state.item.loading;
export const selectItemError = (state: RootState) => state.item.error;
export const selectSelectedItem = (state: RootState) => state.item.selectedItem;
export const selectItemPagination = (state: RootState) => ({
    page: state.item.page,
    limit: state.item.limit,
    total: state.item.total,
});
export const selectItemLoadingCreate = (state: RootState) => state.item.loadingCreate;
export const selectItemLoadingUpdate = (state: RootState) => state.item.loadingUpdate;
export const selectItemLoadingDelete = (state: RootState) => state.item.loadingDelete;
export const selectItemLoadingFetch = (state: RootState) => state.item.loadingFetch;

export default itemSlice.reducer;
