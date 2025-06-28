import React from 'react';
import { Box, Typography, Paper, Container } from '@mui/material';
import CoAList from '../../features/accounting/coa/CoAList'; // We will create this next

const ChartOfAccountsPage: React.FC = () => {
  return (
    <Container maxWidth="lg">
      <Box sx={{ my: 4 }}>
        <Typography variant="h4" component="h1" gutterBottom>
          Chart of Accounts
        </Typography>
        <Paper elevation={3} sx={{ p: 2 }}>
          {/* CoACreateForm could be a modal triggered from CoAList or a separate component here */}
          {/* For now, CoAList will contain the trigger for creating */}
          <CoAList />
        </Paper>
      </Box>
    </Container>
  );
};

export default ChartOfAccountsPage;
