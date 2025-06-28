import React, { useEffect, useState, useCallback } from 'react';
import { useDispatch, useSelector } from 'react-redux';
import {
  Table, TableBody, TableCell, TableContainer, TableHead, TableRow, Paper, Button,
  TablePagination, CircularProgress, Alert, Box, Typography, IconButton, Tooltip,
  TextField, Grid, MenuItem, Select, FormControl, InputLabel
} from '@mui/material';
import AddIcon from '@mui/icons-material/Add';
import EditIcon from '@mui/icons-material/Edit';
import DeleteIcon from '@mui/icons-material/Delete';
import RefreshIcon from '@mui/icons-material/Refresh';

import { AppDispatch, RootState } from '../../../store';
import { fetchCoAs, deleteCoA, setCurrentPage, setPageLimit, selectCoAList, selectCoALoading, selectCoAError, selectCoAPagination, resetCoAErrors } from './coaSlice';
import { AccountType, ChartOfAccount, ListChartOfAccountsRequest } from './coaTypes';
import CoACreateForm from './CoACreateForm';
import CoAEditForm from './CoAEditForm';

const CoAList: React.FC = () => {
  const dispatch: AppDispatch = useDispatch();
  const accounts = useSelector(selectCoAList);
  const loading = useSelector(selectCoALoading);
  const error = useSelector(selectCoAError);
  const { page, limit, total } = useSelector(selectCoAPagination);

  const [filters, setFilters] = useState<ListChartOfAccountsRequest>({
    account_name: '',
    account_type: '',
    is_active: '', // Changed to string to accommodate "all" option
  });

  const [openCreateModal, setOpenCreateModal] = useState(false);
  const [openEditModal, setOpenEditModal] = useState(false);
  const [editingAccount, setEditingAccount] = useState<ChartOfAccount | null>(null);

  const loadCoAs = useCallback((resetError: boolean = false) => {
    if(resetError) {
      dispatch(resetCoAErrors());
    }
    const params: ListChartOfAccountsRequest = { page, limit };
    if (filters.account_name) params.account_name = filters.account_name;
    if (filters.account_type) params.account_type = filters.account_type as AccountType;

    if (filters.is_active === 'true') params.is_active = true;
    else if (filters.is_active === 'false') params.is_active = false;
    // If 'all' or empty, don't send is_active param to backend, so it uses default behavior

    dispatch(fetchCoAs(params));
  }, [dispatch, page, limit, filters]);

  useEffect(() => {
    loadCoAs();
  }, [loadCoAs]);

  const handleChangePage = (event: unknown, newPage: number) => {
    dispatch(setCurrentPage(newPage + 1)); // API is 1-based, MUI TablePagination is 0-based
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    dispatch(setPageLimit(parseInt(event.target.value, 10)));
    dispatch(setCurrentPage(1)); // Reset to first page
  };

  const handleFilterChange = (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | any) => {
    // For Select components, event might not be typical HTML event
    const { name, value } = event.target;
    setFilters(prev => ({ ...prev, [name]: value }));
  };

  const handleApplyFilters = () => {
    dispatch(setCurrentPage(1)); // Reset to page 1 when filters change
    loadCoAs(); // Reload with current page (now 1) and new filters
  };

  const handleDelete = (id: string) => {
    if (window.confirm('Are you sure you want to delete this account?')) {
      dispatch(deleteCoA(id)).then(() => {
        // Optional: show success message or refetch list if not automatically updated by slice logic
        // loadCoAs(); // If deleteCoA doesn't update the list sufficiently
      });
    }
  };

  const handleOpenCreateModal = () => {
    dispatch(resetCoAErrors()); // Clear errors before opening
    setOpenCreateModal(true);
  };
  const handleCloseCreateModal = () => {
    setOpenCreateModal(false);
    // Optionally refetch or rely on optimistic updates / cache invalidation
    // For simplicity, a refetch might be good if createCoA doesn't update list
    loadCoAs();
  };

  const handleOpenEditModal = (account: ChartOfAccount) => {
    dispatch(resetCoAErrors()); // Clear errors
    setEditingAccount(account);
    setOpenEditModal(true);
  };
  const handleCloseEditModal = () => {
    setOpenEditModal(false);
    setEditingAccount(null);
    // Refetch on close to see updated data, or rely on slice update logic
    loadCoAs();
  };


  if (loading && accounts.length === 0 && !error) { // Show full page loader only on initial load and no error
    return <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}><CircularProgress /></Box>;
  }

  return (
    <Box sx={{ width: '100%' }}>
      <Typography variant="h6" gutterBottom>Accounts List</Typography>

      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={12} sm={3}>
          <TextField
            label="Account Name"
            name="account_name"
            value={filters.account_name}
            onChange={handleFilterChange}
            variant="outlined"
            fullWidth
            size="small"
          />
        </Grid>
        <Grid item xs={12} sm={3}>
          <FormControl fullWidth size="small" variant="outlined">
            <InputLabel>Account Type</InputLabel>
            <Select
              name="account_type"
              value={filters.account_type}
              onChange={handleFilterChange}
              label="Account Type"
            >
              <MenuItem value=""><em>All Types</em></MenuItem>
              {Object.values(AccountType).map(type => (
                <MenuItem key={type} value={type}>{type}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={3}>
           <FormControl fullWidth size="small" variant="outlined">
            <InputLabel>Status</InputLabel>
            <Select
              name="is_active"
              value={filters.is_active}
              onChange={handleFilterChange}
              label="Status"
            >
              <MenuItem value=""><em>All</em></MenuItem>
              <MenuItem value="true">Active</MenuItem>
              <MenuItem value="false">Inactive</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={3} sx={{display: 'flex', gap: 1}}>
          <Button variant="contained" onClick={handleApplyFilters} size="medium">Filter</Button>
          <Tooltip title="Refresh Data">
            <IconButton onClick={loadCoAs} color="primary">
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Grid>
      </Grid>

      <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 2 }}>
        <Button
          variant="contained"
          color="primary"
          startIcon={<AddIcon />}
          onClick={handleOpenCreateModal}
        >
          Add New Account
        </Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      {loading && <CircularProgress sx={{display: 'block', margin: 'auto', mb: 1}} size={24} />}


      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="chart of accounts table">
          <TableHead>
            <TableRow sx={{ backgroundColor: (theme) => theme.palette.action.hover }}>
              <TableCell>Code</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Active</TableCell>
              <TableCell>Control Account</TableCell>
              <TableCell>Description</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {accounts.map((account) => (
              <TableRow key={account.id}>
                <TableCell>{account.account_code}</TableCell>
                <TableCell>{account.account_name}</TableCell>
                <TableCell>{account.account_type}</TableCell>
                <TableCell>{account.is_active ? 'Yes' : 'No'}</TableCell>
                <TableCell>{account.is_control_account ? 'Yes' : 'No'}</TableCell>
                <TableCell sx={{maxWidth: 200, overflow: 'hidden', textOverflow: 'ellipsis', whiteSpace: 'nowrap'}}>
                  <Tooltip title={account.description || ''}>
                    <span>{account.description || '-'}</span>
                  </Tooltip>
                </TableCell>
                <TableCell align="right">
                  <Tooltip title="Edit Account">
                    <IconButton
                      size="small"
                      color="primary"
                      onClick={() => handleOpenEditModal(account)}
                    >
                      <EditIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Delete Account">
                    <IconButton size="small" color="error" onClick={() => handleDelete(account.id)}>
                      <DeleteIcon />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
             {accounts.length === 0 && !loading && (
              <TableRow>
                <TableCell colSpan={7} align="center">
                  No accounts found.
                </TableCell>
              </TableRow>
            )}
          </TableBody>
        </Table>
      </TableContainer>
      <TablePagination
        rowsPerPageOptions={[5, 10, 20, 50]}
        component="div"
        count={total}
        rowsPerPage={limit}
        page={page - 1} // MUI TablePagination is 0-based
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />

      <CoACreateForm open={openCreateModal} onClose={handleCloseCreateModal} />
      {editingAccount && <CoAEditForm open={openEditModal} onClose={handleCloseEditModal} account={editingAccount} />}
    </Box>
  );
};

export default CoAList;
