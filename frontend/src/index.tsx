import React from 'react';
import ReactDOM from 'react-dom/client';
import { Provider } from 'react-redux';
import { ThemeProvider } from '@mui/material/styles';
import CssBaseline from '@mui/material/CssBaseline';
import * as Sentry from "@sentry/react";

import App from './App';
import store from './store';
import theme from './theme'; // Import the custom theme
import reportWebVitals from './reportWebVitals';
import './index.css';

// Initialize Sentry
// TODO: Replace with actual DSN and configure environment, release, etc.
// Ensure this is done securely and not hardcoded for production builds.
Sentry.init({
  dsn: "YOUR_SENTRY_DSN_GOES_HERE", // Replace with your actual DSN
  integrations: [
    new Sentry.BrowserTracing({
      tracePropagationTargets: ["localhost", /^https:\/\/yourserver\.io\/api/],
    }),
    new Sentry.Replay(),
  ],
  tracesSampleRate: 1.0,
  replaysSessionSampleRate: 0.1,
  replaysOnErrorSampleRate: 1.0,
  // environment: process.env.NODE_ENV, // Consider uncommenting and configuring
  // release: "your-project-name@your-version", // Consider uncommenting and configuring
});


const root = ReactDOM.createRoot(
  document.getElementById('root') as HTMLElement
);
root.render(
  <React.StrictMode>
    <Provider store={store}>
      <ThemeProvider theme={theme}> {/* Use the imported custom theme */}
        <CssBaseline /> {/* Normalizes styles and applies background color from theme */}
        <App />
      </ThemeProvider>
    </Provider>
  </React.StrictMode>
);

// If you want to start measuring performance in your app, pass a function
// to log results (for example: reportWebVitals(console.log))
// or send to an analytics endpoint. Learn more: https://bit.ly/CRA-vitals
reportWebVitals();
