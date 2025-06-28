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
import {
  fetchItems, deleteItem, setCurrentItemPage, setItemPageLimit, resetItemErrors,
  selectItemList, selectItemLoading, selectItemError, selectItemPagination
} from './itemSlice';
import { ItemType, Item as InventoryItem, ListItemRequest, ValuationMethod } from './itemTypes'; // Renamed Item to InventoryItem to avoid conflict
import ItemCreateForm from './ItemCreateForm';
import ItemEditForm from './ItemEditForm';

const ItemList: React.FC = () => {
  const dispatch: AppDispatch = useDispatch();
  const items = useSelector(selectItemList);
  const loading = useSelector(selectItemLoading);
  const error = useSelector(selectItemError);
  const { page, limit, total } = useSelector(selectItemPagination);

  const [filters, setFilters] = useState<ListItemRequest>({
    name: '',
    sku: '',
    item_type: '',
    is_active: '', // "all", "true", "false"
  });

  const [openCreateModal, setOpenCreateModal] = useState(false);
  const [openEditModal, setOpenEditModal] = useState(false);
  const [editingItem, setEditingItem] = useState<InventoryItem | null>(null);

  const loadItems = useCallback((resetError: boolean = false) => {
    if(resetError) {
      dispatch(resetItemErrors());
    }
    const params: ListItemRequest = { page, limit };
    if (filters.name) params.name = filters.name;
    if (filters.sku) params.sku = filters.sku;
    if (filters.item_type) params.item_type = filters.item_type as ItemType;

    if (filters.is_active === 'true') params.is_active = true;
    else if (filters.is_active === 'false') params.is_active = false;

    dispatch(fetchItems(params));
  }, [dispatch, page, limit, filters]);

  useEffect(() => {
    loadItems(true); // Reset error on initial load
  }, [loadItems]);

  const handleChangePage = (event: unknown, newPage: number) => {
    dispatch(setCurrentItemPage(newPage + 1));
  };

  const handleChangeRowsPerPage = (event: React.ChangeEvent<HTMLInputElement>) => {
    dispatch(setItemPageLimit(parseInt(event.target.value, 10)));
    dispatch(setCurrentItemPage(1));
  };

  const handleFilterChange = (event: React.ChangeEvent<HTMLInputElement | HTMLTextAreaElement> | any) => {
    const { name, value } = event.target;
    setFilters(prev => ({ ...prev, [name]: value }));
  };

  const handleApplyFilters = () => {
    dispatch(setCurrentItemPage(1));
    loadItems();
  };

  const handleDelete = (id: string) => {
    if (window.confirm('Are you sure you want to delete this item?')) {
      dispatch(deleteItem(id));
      // Refetch or rely on slice logic
    }
  };

  const handleOpenCreateModal = () => {
    dispatch(resetItemErrors());
    setOpenCreateModal(true);
  };
  const handleCloseCreateModal = () => {
    setOpenCreateModal(false);
    loadItems();
  };

  const handleOpenEditModal = (item: InventoryItem) => {
    dispatch(resetItemErrors());
    setEditingItem(item);
    setOpenEditModal(true);
  };
  const handleCloseEditModal = () => {
    setOpenEditModal(false);
    setEditingItem(null);
    loadItems();
  };

  if (loading && items.length === 0 && !error) {
    return <Box sx={{ display: 'flex', justifyContent: 'center', alignItems: 'center', height: '200px' }}><CircularProgress /></Box>;
  }

  return (
    <Box sx={{ width: '100%' }}>
      <Typography variant="h6" gutterBottom>Item Master List</Typography>

      <Grid container spacing={2} sx={{ mb: 2 }}>
        <Grid item xs={12} sm={6} md={3}>
          <TextField label="Item Name" name="name" value={filters.name} onChange={handleFilterChange} variant="outlined" fullWidth size="small"/>
        </Grid>
        <Grid item xs={12} sm={6} md={2}>
          <TextField label="SKU" name="sku" value={filters.sku} onChange={handleFilterChange} variant="outlined" fullWidth size="small"/>
        </Grid>
        <Grid item xs={12} sm={6} md={2}>
          <FormControl fullWidth size="small" variant="outlined">
            <InputLabel>Item Type</InputLabel>
            <Select name="item_type" value={filters.item_type} onChange={handleFilterChange} label="Item Type">
              <MenuItem value=""><em>All Types</em></MenuItem>
              {Object.values(ItemType).map(type => (
                <MenuItem key={type} value={type}>{type}</MenuItem>
              ))}
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={6} md={2}>
           <FormControl fullWidth size="small" variant="outlined">
            <InputLabel>Status</InputLabel>
            <Select name="is_active" value={filters.is_active} onChange={handleFilterChange} label="Status">
              <MenuItem value=""><em>All</em></MenuItem>
              <MenuItem value="true">Active</MenuItem>
              <MenuItem value="false">Inactive</MenuItem>
            </Select>
          </FormControl>
        </Grid>
        <Grid item xs={12} sm={6} md={3} sx={{display: 'flex', gap: 1, alignItems: 'center'}}>
          <Button variant="contained" onClick={handleApplyFilters} size="medium">Filter</Button>
          <Tooltip title="Refresh Data">
            <IconButton onClick={() => loadItems(true)} color="primary">
              <RefreshIcon />
            </IconButton>
          </Tooltip>
        </Grid>
      </Grid>

      <Box sx={{ display: 'flex', justifyContent: 'flex-end', mb: 2 }}>
        <Button variant="contained" color="primary" startIcon={<AddIcon />} onClick={handleOpenCreateModal}>
          Add New Item
        </Button>
      </Box>

      {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
      {loading && <CircularProgress sx={{display: 'block', margin: 'auto', mb: 1}} size={24} />}

      <TableContainer component={Paper}>
        <Table sx={{ minWidth: 650 }} aria-label="inventory items table">
          <TableHead sx={{ backgroundColor: (theme) => theme.palette.action.hover }}>
            <TableRow>
              <TableCell>SKU</TableCell>
              <TableCell>Name</TableCell>
              <TableCell>Type</TableCell>
              <TableCell>Unit</TableCell>
              <TableCell>Sales Price</TableCell>
              <TableCell>Purchase Price</TableCell>
              <TableCell>Valuation</TableCell>
              <TableCell>Active</TableCell>
              <TableCell align="right">Actions</TableCell>
            </TableRow>
          </TableHead>
          <TableBody>
            {items.map((item) => (
              <TableRow key={item.id}>
                <TableCell>{item.sku}</TableCell>
                <TableCell>{item.name}</TableCell>
                <TableCell>{item.item_type}</TableCell>
                <TableCell>{item.unit_of_measure}</TableCell>
                <TableCell>{item.sales_price !== null ? item.sales_price?.toFixed(2) : '-'}</TableCell>
                <TableCell>{item.purchase_price !== null ? item.purchase_price?.toFixed(2) : '-'}</TableCell>
                <TableCell>{item.valuation_method || '-'}</TableCell>
                <TableCell>{item.is_active ? 'Yes' : 'No'}</TableCell>
                <TableCell align="right">
                  <Tooltip title="Edit Item">
                    <IconButton size="small" color="primary" onClick={() => handleOpenEditModal(item)}>
                      <EditIcon />
                    </IconButton>
                  </Tooltip>
                  <Tooltip title="Delete Item">
                    <IconButton size="small" color="error" onClick={() => handleDelete(item.id)}>
                      <DeleteIcon />
                    </IconButton>
                  </Tooltip>
                </TableCell>
              </TableRow>
            ))}
             {items.length === 0 && !loading && (
              <TableRow>
                <TableCell colSpan={9} align="center">
                  No items found.
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
        page={page - 1}
        onPageChange={handleChangePage}
        onRowsPerPageChange={handleChangeRowsPerPage}
      />

      <ItemCreateForm open={openCreateModal} onClose={handleCloseCreateModal} />
      {editingItem && <ItemEditForm open={openEditModal} onClose={handleCloseEditModal} item={editingItem} />}
    </Box>
  );
};

export default ItemList;
