import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { ThemeProvider, createTheme } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import * as Sentry from "@sentry/react";
import App from './App';
import store from './store';
import reportWebVitals from './reportWebVitals';
import './index.css';

// Initialize Sentry
// TODO: Replace with actual DSN and configure environment, release, etc.
// Ensure this is done securely and not hardcoded for production builds.
Sentry.init({
  dsn: "YOUR_SENTRY_DSN_GOES_HERE", // Replace with your actual DSN
  integrations: [
    new Sentry.BrowserTracing({
      // Set 'tracePropagationTargets' to control for which URLs distributed tracing should be enabled
      tracePropagationTargets: ["localhost", /^https:\/\/yourserver\.io\/api/],
    }),
    new Sentry.Replay(),
  ],
  // Performance Monitoring
  tracesSampleRate: 1.0, // Capture 100% of the transactions, reduce in production!
  // Session Replay
  replaysSessionSampleRate: 0.1, // This sets the sample rate at 10%. You may want to change it to 100% while in development and then sample at a lower rate in production.
  replaysOnErrorSampleRate: 1.0, // If you're not already sampling the entire session, change the sample rate to 100% when sampling sessions where errors occur.
  // TODO: Set environment, release version, etc.
  // environment: process.env.NODE_ENV,
  // release: "my-project-name@1.0.0",
});

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
