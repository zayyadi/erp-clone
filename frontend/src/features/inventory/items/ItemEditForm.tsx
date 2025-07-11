import React, { useEffect } from 'react';
import { useForm, Controller, SubmitHandler } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import {
  Button, TextField, Dialog, DialogActions, DialogContent, DialogTitle, Grid,
  MenuItem, FormControlLabel, Checkbox, CircularProgress, Alert, Box
} from '@mui/material';
import { useDispatch, useSelector } from 'react-redux';
import { AppDispatch, RootState } from '../../../store';
import { updateItem, selectItemLoadingUpdate, selectItemError, resetItemErrors } from './itemSlice';
import { ItemType, UpdateItemRequest, Item as InventoryItem, ValuationMethod } from './itemTypes';

interface ItemEditFormProps {
  open: boolean;
  onClose: () => void;
  item: InventoryItem | null;
}

const schema = yup.object().shape({
  // SKU is typically not editable
  name: yup.string().required('Item name is required').min(3, 'Min 3 chars').max(100, 'Max 100 chars'),
  item_type: yup.string<ItemType>().oneOf(Object.values(ItemType)).required('Item type is required'),
  unit_of_measure: yup.string().required('Unit of Measure is required').max(20, 'Max 20 chars'),
  description: yup.string().optional().max(255, 'Max 255 chars'),
  purchase_price: yup.number().nullable().optional().min(0, 'Must be positive'),
  sales_price: yup.number().nullable().optional().min(0, 'Must be positive'),
  valuation_method: yup.string<ValuationMethod>().oneOf(Object.values(ValuationMethod)).nullable().optional(),
  is_active: yup.boolean().optional(),
});

const ItemEditForm: React.FC<ItemEditFormProps> = ({ open, onClose, item }) => {
  const dispatch: AppDispatch = useDispatch();
  const loading = useSelector(selectItemLoadingUpdate);
  const error = useSelector(selectItemError);

  const { control, handleSubmit, reset, formState: { errors, isDirty, isValid } } = useForm<UpdateItemRequest>({
    resolver: yupResolver(schema),
    defaultValues: { // Default values will be overridden by item data in useEffect
      name: '',
      item_type: ItemType.FinishedGood,
      unit_of_measure: 'pcs',
      description: '',
      purchase_price: null,
      sales_price: null,
      valuation_method: null,
      is_active: true,
    },
    mode: 'onChange',
  });

  useEffect(() => {
    if (open) {
      dispatch(resetItemErrors());
      if (item) {
        reset({
          name: item.name,
          item_type: item.item_type,
          unit_of_measure: item.unit_of_measure,
          description: item.description || '',
          purchase_price: item.purchase_price === undefined ? null : item.purchase_price,
          sales_price: item.sales_price === undefined ? null : item.sales_price,
          valuation_method: item.valuation_method || null,
          is_active: item.is_active,
        });
      }
    }
  }, [open, item, reset, dispatch]);

  const onSubmit: SubmitHandler<UpdateItemRequest> = async (data) => {
    if (!item) return;
    const payload: UpdateItemRequest = {
        ...data,
        is_active: data.is_active ?? true,
        purchase_price: data.purchase_price === undefined || data.purchase_price === null || isNaN(Number(data.purchase_price)) ? null : Number(data.purchase_price),
        sales_price: data.sales_price === undefined || data.sales_price === null || isNaN(Number(data.sales_price)) ? null : Number(data.sales_price),
    };
    const resultAction = await dispatch(updateItem({ id: item.id, itemData: payload }));
    if (updateItem.fulfilled.match(resultAction)) {
      onClose();
    }
  };

  if (!item) return null;

  return (
    <Dialog open={open} onClose={onClose} maxWidth="md" fullWidth>
      <DialogTitle>Edit Item: {item.sku} - {item.name}</DialogTitle>
      <form onSubmit={handleSubmit(onSubmit)}>
        <DialogContent>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <Grid container spacing={2} sx={{mt:1}}>
            <Grid item xs={12} sm={6}>
              <TextField label="SKU (Read-only)" value={item.sku} variant="outlined" fullWidth InputProps={{ readOnly: true }} disabled />
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="name" control={control} render={({ field }) => (
                <TextField {...field} label="Item Name" variant="outlined" fullWidth required error={!!errors.name} helperText={errors.name?.message} />
              )}/>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="item_type" control={control} render={({ field }) => (
                <TextField {...field} select label="Item Type" variant="outlined" fullWidth required error={!!errors.item_type} helperText={errors.item_type?.message}>
                  {Object.values(ItemType).map((type) => ( <MenuItem key={type} value={type}>{type}</MenuItem> ))}
                </TextField>
              )}/>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="unit_of_measure" control={control} render={({ field }) => (
                <TextField {...field} label="Unit of Measure (e.g., pcs, kg)" variant="outlined" fullWidth required error={!!errors.unit_of_measure} helperText={errors.unit_of_measure?.message} />
              )}/>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="purchase_price" control={control} render={({ field }) => (
                <TextField {...field} label="Purchase Price (Optional)" variant="outlined" fullWidth type="number" error={!!errors.purchase_price} helperText={errors.purchase_price?.message}
                           onChange={e => field.onChange(e.target.value === '' ? null : parseFloat(e.target.value))} value={field.value === null ? '' : field.value} inputProps={{ step: "0.01" }}/>
              )}/>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="sales_price" control={control} render={({ field }) => (
                <TextField {...field} label="Sales Price (Optional)" variant="outlined" fullWidth type="number" error={!!errors.sales_price} helperText={errors.sales_price?.message}
                           onChange={e => field.onChange(e.target.value === '' ? null : parseFloat(e.target.value))} value={field.value === null ? '' : field.value} inputProps={{ step: "0.01" }}/>
              )}/>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="valuation_method" control={control} render={({ field }) => (
                <TextField {...field} select label="Valuation Method (Optional)" variant="outlined" fullWidth error={!!errors.valuation_method} helperText={errors.valuation_method?.message} value={field.value || ''}>
                  <MenuItem value=""><em>None</em></MenuItem>
                  {Object.values(ValuationMethod).map((method) => ( <MenuItem key={method} value={method}>{method}</MenuItem> ))}
                </TextField>
              )}/>
            </Grid>
            <Grid item xs={12}>
              <Controller name="description" control={control} render={({ field }) => (
                <TextField {...field} label="Description (Optional)" variant="outlined" fullWidth multiline rows={3} error={!!errors.description} helperText={errors.description?.message} />
              )}/>
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller name="is_active" control={control} render={({ field }) => (
                <FormControlLabel control={<Checkbox {...field} checked={field.value} />} label="Item is Active" />
              )}/>
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions sx={{p:2}}>
          <Button onClick={onClose} color="secondary" variant="outlined">Cancel</Button>
          <Button type="submit" variant="contained" color="primary" disabled={loading || !isDirty || !isValid}>
            {loading ? <CircularProgress size={24} color="inherit" /> : 'Save Changes'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default ItemEditForm;
