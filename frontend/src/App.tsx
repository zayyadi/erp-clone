import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link as RouterLink } from 'react-router-dom';
import {
  Box, AppBar, Toolbar, Typography, Drawer, List, ListItem, ListItemButton, ListItemIcon, ListItemText,
  Container, CssBaseline, IconButton, Paper, Grid, useTheme, useMediaQuery
} from '@mui/material';
import MenuIcon from '@mui/icons-material/Menu';
import HomeIcon from '@mui/icons-material/Home';
import StorefrontIcon from '@mui/icons-material/Storefront';
import AccountBalanceIcon from '@mui/icons-material/AccountBalance';
import AssessmentIcon from '@mui/icons-material/Assessment';
import { styled } from '@mui/material/styles';

const drawerWidth = 240;

const Main = styled('main', { shouldForwardProp: (prop) => prop !== 'open' })<{
  open?: boolean;
}>(({ theme, open }) => ({
  flexGrow: 1,
  padding: theme.spacing(3),
  transition: theme.transitions.create('margin', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  marginLeft: 0,
  // Apply margin shift only if not mobile and drawer is open
  ...(!props.isMobile && props.open && {
    marginLeft: `${drawerWidth}px`,
    transition: theme.transitions.create('margin', {
      easing: theme.transitions.easing.easeOut,
      duration: theme.transitions.duration.enteringScreen,
    }),
  }),
  position: 'relative',
}));

// Add isMobile to Main component props
interface MainProps {
  open?: boolean;
  isMobile?: boolean;
}

const Main = styled('main', { shouldForwardProp: (prop) => prop !== 'open' && prop !== 'isMobile' })<MainProps>(
  ({ theme, open, isMobile }) => ({ // Destructure isMobile here
  flexGrow: 1,
  padding: theme.spacing(3),
  transition: theme.transitions.create('margin', {
    easing: theme.transitions.easing.sharp,
    duration: theme.transitions.duration.leavingScreen,
  }),
  marginLeft: 0,
  ...(!isMobile && open && { // Use destructured isMobile and open
    marginLeft: `${drawerWidth}px`,
    transition: theme.transitions.create('margin', {
      easing: theme.transitions.easing.easeOut,
      duration: theme.transitions.duration.enteringScreen,
    }),
  }),
  position: 'relative',
}));


const DrawerHeader = styled('div')(({ theme }) => ({
  display: 'flex',
  alignItems: 'center',
  padding: theme.spacing(0, 1),
  ...theme.mixins.toolbar,
  justifyContent: 'flex-end',
}));

// Placeholder components for pages
const DashboardPage = () => <Typography variant="h4">Dashboard</Typography>;
// const InventoryPage = () => <Typography variant="h4">Inventory Management</Typography>; // To be replaced by actual page or sub-routes
const ReportsPage = () => <Typography variant="h4">Reports Section</Typography>;

// Import actual pages
import ChartOfAccountsPage from './pages/Accounting/ChartOfAccountsPage';
import ItemsPage from './pages/Inventory/ItemsPage'; // Import ItemsPage
import GlobalSnackbar from './features/notifications/GlobalSnackbar'; // Import GlobalSnackbar


function App() {
  const theme = useTheme();
  const isMobile = useMediaQuery(theme.breakpoints.down('sm')); // Check for screens smaller than 'sm' (600px)
  const [open, setOpen] = React.useState(!isMobile); // Drawer closed by default on mobile, open on desktop

  React.useEffect(() => {
    // Adjust drawer state if screen size changes (e.g., rotate tablet)
    if (isMobile) {
      setOpen(false);
    } else {
      setOpen(true);
    }
  }, [isMobile]);

  const handleDrawerToggle = () => {
    setOpen(!open);
  };

  const menuItems = [
    { text: 'Dashboard', icon: <HomeIcon />, path: '/' },
    { text: 'Inventory', icon: <StorefrontIcon />, path: '/inventory' },
    { text: 'Accounting', icon: <AccountBalanceIcon />, path: '/accounting' },
    { text: 'Reports', icon: <AssessmentIcon />, path: '/reports' },
  ];

  return (
    <Router>
      <Box sx={{ display: 'flex' }}>
        <CssBaseline />
        <GlobalSnackbar /> {/* Add GlobalSnackbar here */}
        <AppBar
          position="fixed"
          sx={{
            zIndex: (theme) => theme.zIndex.drawer + 1,
            transition: (theme) => theme.transitions.create(['margin', 'width'], {
              easing: theme.transitions.easing.sharp,
              duration: theme.transitions.duration.leavingScreen,
            }),
            // Only apply margin/width shift if not mobile and drawer is open
            ...(!isMobile && open && {
              width: `calc(100% - ${drawerWidth}px)`,
              marginLeft: `${drawerWidth}px`,
              transition: (theme) => theme.transitions.create(['margin', 'width'], {
                easing: theme.transitions.easing.easeOut,
                duration: theme.transitions.duration.enteringScreen,
              }),
            }),
          }}
        >
          <Toolbar>
            <IconButton
              color="inherit"
              aria-label="open drawer"
              edge="start"
              onClick={handleDrawerToggle}
              // Always show menu button on mobile, hide on desktop when drawer is open & persistent
              sx={{ mr: 2, ...(open && !isMobile && { display: 'none' }) }}
            >
              <MenuIcon />
            </IconButton>
            <Typography variant="h6" noWrap component="div">
              ERP System
            </Typography>
          </Toolbar>
        </AppBar>
        <Drawer
          variant={isMobile ? "temporary" : "persistent"}
          anchor="left"
          open={open}
          onClose={isMobile ? handleDrawerToggle : undefined} // Close on backdrop click on mobile
          ModalProps={{
            keepMounted: true, // Better open performance on mobile.
          }}
          sx={{
            width: drawerWidth,
            flexShrink: 0,
            [`& .MuiDrawer-paper`]: { width: drawerWidth, boxSizing: 'border-box' },
          }}
        >
          <DrawerHeader>
            {/* Show close button inside drawer only if it's persistent and open, or if mobile (temporary ones often have it in header) */}
            <IconButton onClick={handleDrawerToggle}>
              <MenuIcon /> {/* Or use ChevronLeftIcon theme.direction === 'rtl' ? <ChevronRightIcon /> : <ChevronLeftIcon /> */}
            </IconButton>
          </DrawerHeader>
          {/* <Divider /> */}
          <List onClick={isMobile ? handleDrawerToggle : undefined} /* Close drawer on mobile after item click */ >
            {menuItems.map((item) => (
              <ListItem key={item.text} disablePadding component={RouterLink} to={item.path} sx={{ color: 'inherit', textDecoration: 'none' }}>
                <ListItemButton>
                  <ListItemIcon>{item.icon}</ListItemIcon>
                  <ListItemText primary={item.text} />
                </ListItemButton>
              </ListItem>
            ))}
          </List>
        </Drawer>
        <Main open={open} isMobile={isMobile} sx={{ width: '100%'}}> {/* Pass isMobile prop */}
          <DrawerHeader /> {/* Necessary to prevent content from being hidden under the AppBar */}
          <Container maxWidth="xl"> {/* Changed to xl for wider content area, can be "lg" or false */}
            <Paper elevation={3} sx={{ p: 2, display: 'flex', flexDirection: 'column', minHeight: 'calc(100vh - 64px - 48px)' }}> {/* 64px AppBar, 48px padding */}
              <Routes>
                <Route path="/" element={<DashboardPage />} />

                {/* Inventory Routes */}
                <Route path="/inventory" element={<ItemsPage />} /> {/* Default to Items List for now */}
                <Route path="/inventory/items" element={<ItemsPage />} />
                {/* <Route path="/inventory/warehouses" element={<WarehousesPage />} /> */}
                {/* <Route path="/inventory/adjustments" element={<AdjustmentsPage />} /> */}
                {/* <Route path="/inventory/levels" element={<InventoryLevelsPage />} /> */}

                {/* Accounting Routes */}
                <Route path="/accounting" element={<ChartOfAccountsPage />} /> {/* Default to CoA */}
                <Route path="/accounting/chart-of-accounts" element={<ChartOfAccountsPage />} />
                {/* <Route path="/accounting/journal-entries" element={<JournalEntriesPage />} /> */}
                {/* <Route path="/accounting/reports/trial-balance" element={<TrialBalancePage />} /> */}

                <Route path="/reports" element={<ReportsPage />} />
                {/* Add more specific routes for sub-modules later */}
              </Routes>
            </Paper>
            {/* Example Grid from before, can be removed or adapted */}
            <Grid container spacing={3} sx={{mt: 2}}>
              <Grid item xs={12} md={6}>
                 <Paper sx={{ p: 2, height: 100 }}>
                    <Typography>Sample Widget Area 1</Typography>
                 </Paper>
              </Grid>
              <Grid item xs={12} md={6}>
                 <Paper sx={{ p: 2, height: 100 }}>
                    <Typography>Sample Widget Area 2</Typography>
                 </Paper>
              </Grid>
            </Grid>
          </Container>
        </Main>
      </Box>
    </Router>
  );
}

export default App;
