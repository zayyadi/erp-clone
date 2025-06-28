import React, { ReactElement } from 'react';
import { render, RenderOptions } from '@testing-library/react';
import { Provider } from 'react-redux';
import { BrowserRouter as Router } from 'react-router-dom';
import { ThemeProvider } from '@mui/material/styles';
import { configureStore, Store } from '@reduxjs/toolkit';

import theme from './theme';
import coaReducer from './features/accounting/coa/coaSlice';
import itemReducer from './features/inventory/items/itemSlice';
import notificationReducer from './features/notifications/notificationSlice';
import { RootState } from './store'; // Assuming RootState is exported from your main store

// Define a type for the preloaded state if you need it for specific tests
type PreloadedState = Partial<RootState>;

interface ExtendedRenderOptions extends Omit<RenderOptions, 'queries'> {
  preloadedState?: PreloadedState;
  store?: Store<RootState>; // Allow passing a specific store instance
}

// This utility function wraps the component with necessary providers
const renderWithProviders = (
  ui: ReactElement,
  {
    preloadedState,
    // Automatically create a store instance if no store was passed in
    store = configureStore({
        reducer: {
            coa: coaReducer,
            item: itemReducer,
            notification: notificationReducer
        },
        preloadedState
    }),
    ...renderOptions
  }: ExtendedRenderOptions = {}
) => {
  function Wrapper({ children }: { children: React.ReactNode }): JSX.Element {
    return (
      <Provider store={store}>
        <ThemeProvider theme={theme}>
          <Router>{children}</Router>
        </ThemeProvider>
      </Provider>
    );
  }
  return render(ui, { wrapper: Wrapper, ...renderOptions });
};

// Re-export everything from testing-library
export * from '@testing-library/react';
// Override render method
export { renderWithProviders as render };
