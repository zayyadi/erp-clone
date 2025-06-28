import { configureStore } from '@reduxjs/toolkit';
import coaReducer from './features/accounting/coa/coaSlice';
import itemReducer from './features/inventory/items/itemSlice';
import notificationReducer from './features/notifications/notificationSlice'; // Import notification reducer
// Import other reducers here as they are created

export const store = configureStore({
  reducer: {
    coa: coaReducer,
    item: itemReducer,
    notification: notificationReducer, // Add notification reducer
    // example: exampleReducer,
    // Add other reducers here:
    // journal: journalReducer,
    // warehouse: warehouseReducer,
    // etc.
  },
});

// Infer the `RootState` and `AppDispatch` types from the store itself
export type RootState = ReturnType<typeof store.getState>;
export type AppDispatch = typeof store.dispatch;

export default store;
