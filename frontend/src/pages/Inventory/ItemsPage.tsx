import React from 'react';
import { Box, Typography, Paper, Container } from '@mui/material';
import ItemList from '../../features/inventory/items/ItemList'; // We will create this next

const ItemsPage: React.FC = () => {
  return (
    <Container maxWidth="lg">
      <Box sx={{ my: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Inventory Items
        </Typography>
        <Paper elevation={3} sx={{ p: 2 }}>
          <ItemList />
        </Paper>
      </Box>
    </Container>
  );
};

export default ItemsPage;
