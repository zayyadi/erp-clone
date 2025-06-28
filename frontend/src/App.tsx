import React from 'react';
import { BrowserRouter as Router, Routes, Route, Link as RouterLink } from 'react-router-dom';
import { Box, AppBar, Toolbar, Typography, Drawer, List, ListItemButton, ListItemText, Container, Grid, Paper, useTheme } from '@mui/material'; // Added useTheme back
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
  marginLeft: `-${drawerWidth}px`,
  ...(open && {
    transition: theme.transitions.create('margin', {
      easing: theme.transitions.easing.easeOut,
      duration: theme.transitions.duration.enteringScreen,
    }),
    marginLeft: 0,
  }),
}));

const DrawerHeader = styled('div')(({ theme }) => ({
  display: 'flex',
  alignItems: 'center',
  padding: theme.spacing(0, 1),
  // necessary for content to be below app bar
  ...theme.mixins.toolbar,
  justifyContent: 'flex-end',
}));

function App() {
  const theme = useTheme(); // Re-added useTheme
  const [open, setOpen] = React.useState(true); // setOpen might be used for a toggle button

  // Example: Responsive breakpoints (from plan: 900px/600px)
  // These can be used with MUI's `useMediaQuery` hook or directly in sx props.
  // import useMediaQuery from '@mui/material/useMediaQuery';
  // const isMobile = useMediaQuery(theme.breakpoints.down('sm')); // 600px
  // const isTablet = useMediaQuery(theme.breakpoints.down('md')); // 900px

  return (
    <Router>
      <Box sx={{ display: 'flex' }}>
        <AppBar position="fixed" sx={{ zIndex: theme.zIndex.drawer + 1 }}> {/* Use theme here */}
          <Toolbar>
            <Typography variant="h6" noWrap component="div">
              ERP System
            </Typography>
          </Toolbar>
        </AppBar>
        <Drawer
          sx={{
            width: drawerWidth,
            flexShrink: 0,
            '& .MuiDrawer-paper': {
              width: drawerWidth,
              boxSizing: 'border-box',
            },
          }}
          variant="persistent"
          anchor="left"
          open={open} // Controlled by state
        >
          <DrawerHeader>
            {/* Could add an IconButton here to close the drawer using setOpen(!open) */}
            <Typography>Navigation</Typography>
          </DrawerHeader>
          <List>
            {['Dashboard', 'Inventory', 'Accounting', 'Reports'].map((text) => (
              <ListItem button key={text} component={Link} to={`/${text.toLowerCase()}`}>
                <ListItemText primary={text} />
              </ListItem>
            ))}
          </List>
        </Drawer>
        <Main open={open}>
          <DrawerHeader /> {/* This is to offset content below AppBar */}
          <Container maxWidth="lg"> {/* Responsive container */}
            <Grid container spacing={3}> {/* 12-column grid with spacing */}
              <Grid item xs={12}>
                <Paper sx={{ p: 2, display: 'flex', flexDirection: 'column' }}>
                  {/* Content for different routes will go here */}
                  <Routes>
                    <Route path="/" element={<Typography variant="h4">Dashboard</Typography>} />
                    <Route path="/dashboard" element={<Typography variant="h4">Dashboard</Typography>} />
                    <Route path="/inventory" element={<Typography variant="h4">Inventory Management</Typography>} />
                    <Route path="/accounting" element={<Typography variant="h4">Accounting Module</Typography>} />
                    <Route path="/reports" element={<Typography variant="h4">Reports Section</Typography>} />
                  </Routes>
                </Paper>
              </Grid>
              {/* Example of more grid items for layout */}
              <Grid item xs={12} md={6}>
                 <Paper sx={{ p: 2, height: 200 }}>
                    <Typography>Widget 1 (Responsive: 12 cols on small, 6 on medium+)</Typography>
                 </Paper>
              </Grid>
              <Grid item xs={12} md={6}>
                 <Paper sx={{ p: 2, height: 200 }}>
                    <Typography>Widget 2 (Responsive: 12 cols on small, 6 on medium+)</Typography>
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
