import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
// Sentry lines removed
import App from './App';
import store from './store';
import reportWebVitals from './reportWebVitals';
import './index.css';

// Sentry Init block removed

// Basic Material UI theme
const theme = createTheme({
  // You can customize your theme here later based on UI/UX guidelines
  palette: {
    primary: {
      main: '#1976d2', // Example primary color
    },
    secondary: {
      main: '#dc004e', // Example secondary color
    },
  },
  typography: {
    // Example: Adjust font size based on 72-80 chars per line guideline
    body1: {
      fontSize: '1rem', // Default, adjust as needed
    },
  },
  // Example: 8px baseline grid (spacing unit)
  spacing: 8,
});

const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <ThemeProvider theme={theme}>
        <CssBaseline /> {/* Normalizes styles and applies background color */}
        <App />
      </ThemeProvider>
    </Provider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
