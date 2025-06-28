import React from 'react';
import { render, screen, within } from './test-utils'; // Import custom render
import App from './App';
import userEvent from '@testing-library/user-event';

describe('App Component', () => {
  test('renders main application layout and default page (Dashboard)', () => {
    render(<App />);

    // Check for AppBar title
    expect(screen.getByRole('banner', { name: /ERP System/i })).toBeInTheDocument(); // More accessible way to find AppBar

    // Check for Navigation Drawer (assuming it's open by default on test environment screen size)
    // The drawer itself might not have an explicit role, look for its content.
    const navigation = screen.getByRole('navigation'); // Drawer has 'navigation' role via <nav> or similar
    expect(within(navigation).getByText(/Dashboard/i)).toBeInTheDocument();
    expect(within(navigation).getByText(/Inventory/i)).toBeInTheDocument();
    expect(within(navigation).getByText(/Accounting/i)).toBeInTheDocument();

    // Check for initial page content (Dashboard)
    // The main content area might be identified by a role or a test ID if added.
    // For now, let's assume the Dashboard title is rendered.
    expect(screen.getByRole('heading', { name: /Dashboard/i, level: 4 })).toBeInTheDocument();
  });

  test('navigates to Chart of Accounts page when "Accounting" is clicked', async () => {
    render(<App />);

    const navigation = screen.getByRole('navigation');
    const accountingLink = within(navigation).getByText(/Accounting/i);
    await userEvent.click(accountingLink);

    // Check for Chart of Accounts page title
    expect(screen.getByRole('heading', { name: /Chart of Accounts/i, level: 4 })).toBeInTheDocument();
  });

  test('navigates to Inventory Items page when "Inventory" is clicked', async () => {
    render(<App />);

    const navigation = screen.getByRole('navigation');
    const inventoryLink = within(navigation).getByText(/Inventory/i);
    await userEvent.click(inventoryLink);

    // Check for Inventory Items page title
    expect(screen.getByRole('heading', { name: /Inventory Items/i, level: 4 })).toBeInTheDocument();
  });

  test('AppBar menu button toggles drawer on smaller screens (conceptual)', () => {
    // This test is more conceptual because JSDOM doesn't truly simulate screen sizes for useMediaQuery.
    // We'd typically mock useMediaQuery for this.
    // For this basic test, we'll just check if the button exists.
    // Actual drawer toggling logic is complex to test without deeper mocking.
    render(<App />);
    const menuButton = screen.getByRole('button', { name: /open drawer/i });
    expect(menuButton).toBeInTheDocument();
    // Further interaction testing (await userEvent.click(menuButton)) would require mocking useMediaQuery
    // and verifying changes in drawer's 'open' state or visibility, which is advanced for "basic" tests.
  });

});
