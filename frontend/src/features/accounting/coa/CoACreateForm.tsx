import React, { useEffect } from 'react';
import { useForm, Controller, SubmitHandler } from 'react-hook-form';
import { yupResolver } from '@hookform/resolvers/yup';
import * as yup from 'yup';
import {
  Button, TextField, Dialog, DialogActions, DialogContent, DialogTitle, Grid,
  MenuItem, FormControlLabel, Checkbox, CircularProgress, Alert, Box, Autocomplete
} from '@mui/material';
import { useDispatch, useSelector } from 'react-redux';
import { AppDispatch, RootState } from '../../../store';
import { createCoA, fetchCoAs, selectCoALoadingCreate, selectCoAError, resetCoAErrors, selectCoAList } from './coaSlice';
import { AccountType, CreateChartOfAccountRequest, ChartOfAccount } from './coaTypes';

interface CoACreateFormProps {
  open: boolean;
  onClose: () => void;
}

const schema = yup.object().shape({
  account_code: yup.string().required('Account code is required').min(3, 'Min 3 chars').max(20, 'Max 20 chars'),
  account_name: yup.string().required('Account name is required').min(3, 'Min 3 chars').max(100, 'Max 100 chars'),
  account_type: yup.string<AccountType>().oneOf(Object.values(AccountType)).required('Account type is required'),
  description: yup.string().optional().max(255, 'Max 255 chars'),
  is_active: yup.boolean().optional(),
  is_control_account: yup.boolean().optional(),
  parent_account_id: yup.string().nullable().optional(),
});

const CoACreateForm: React.FC<CoACreateFormProps> = ({ open, onClose }) => {
  const dispatch: AppDispatch = useDispatch();
  const loading = useSelector(selectCoALoadingCreate);
  const error = useSelector(selectCoAError);
  const accounts = useSelector(selectCoAList); // For Parent Account Autocomplete

  const { control, handleSubmit, reset, formState: { errors, isDirty, isValid } } = useForm<CreateChartOfAccountRequest>({
    resolver: yupResolver(schema),
    defaultValues: {
      account_code: '',
      account_name: '',
      account_type: AccountType.Asset, // Default type
      description: '',
      is_active: true,
      is_control_account: false,
      parent_account_id: null,
    },
    mode: 'onChange', // Validate on change to enable/disable submit button
  });

  useEffect(() => {
    if (open) {
      dispatch(resetCoAErrors()); // Clear any previous errors when dialog opens
      reset(); // Reset form fields
      // Fetch accounts if not already loaded, for parent account selection
      if (accounts.length === 0) {
        dispatch(fetchCoAs({ page: 1, limit: 1000 })); // Fetch a large list for dropdown
      }
    }
  }, [open, reset, dispatch, accounts.length]);

  const onSubmit: SubmitHandler<CreateChartOfAccountRequest> = async (data) => {
    // Ensure boolean values are set if undefined (checkboxes can be tricky)
    const payload: CreateChartOfAccountRequest = {
        ...data,
        is_active: data.is_active ?? true,
        is_control_account: data.is_control_account ?? false,
        parent_account_id: data.parent_account_id === '' ? null : data.parent_account_id,
    };

    const resultAction = await dispatch(createCoA(payload));
    if (createCoA.fulfilled.match(resultAction)) {
      onClose(); // Close dialog on success
    }
    // Error is handled by the Alert component via selector
  };

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Create New Chart of Account</DialogTitle>
      <form onSubmit={handleSubmit(onSubmit)}>
        <DialogContent>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <Grid container spacing={2} sx={{mt:1}}>
            <Grid item xs={12} sm={6}>
              <Controller
                name="account_code"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    label="Account Code"
                    variant="outlined"
                    fullWidth
                    required
                    error={!!errors.account_code}
                    helperText={errors.account_code?.message}
                  />
                )}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller
                name="account_name"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    label="Account Name"
                    variant="outlined"
                    fullWidth
                    required
                    error={!!errors.account_name}
                    helperText={errors.account_name?.message}
                  />
                )}
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                name="account_type"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    select
                    label="Account Type"
                    variant="outlined"
                    fullWidth
                    required
                    error={!!errors.account_type}
                    helperText={errors.account_type?.message}
                  >
                    {Object.values(AccountType).map((type) => (
                      <MenuItem key={type} value={type}>
                        {type}
                      </MenuItem>
                    ))}
                  </TextField>
                )}
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                name="parent_account_id"
                control={control}
                render={({ field }) => (
                    <Autocomplete
                        options={accounts.filter(acc => acc.is_control_account)} // Only allow control accounts as parents
                        getOptionLabel={(option: ChartOfAccount) => `${option.account_code} - ${option.account_name}`}
                        value={accounts.find(acc => acc.id === field.value) || null}
                        onChange={(_, newValue) => {
                            field.onChange(newValue ? newValue.id : null);
                        }}
                        onBlur={field.onBlur}
                        renderInput={(params) => (
                            <TextField
                                {...params}
                                label="Parent Account (Optional)"
                                variant="outlined"
                                error={!!errors.parent_account_id}
                                helperText={errors.parent_account_id?.message}
                            />
                        )}
                        isOptionEqualToValue={(option, value) => option.id === value.id}
                    />
                )}
              />
            </Grid>
            <Grid item xs={12}>
              <Controller
                name="description"
                control={control}
                render={({ field }) => (
                  <TextField
                    {...field}
                    label="Description (Optional)"
                    variant="outlined"
                    fullWidth
                    multiline
                    rows={3}
                    error={!!errors.description}
                    helperText={errors.description?.message}
                  />
                )}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller
                name="is_active"
                control={control}
                render={({ field }) => (
                  <FormControlLabel
                    control={<Checkbox {...field} checked={field.value} />}
                    label="Active"
                  />
                )}
              />
            </Grid>
            <Grid item xs={12} sm={6}>
              <Controller
                name="is_control_account"
                control={control}
                render={({ field }) => (
                  <FormControlLabel
                    control={<Checkbox {...field} checked={field.value} />}
                    label="Control Account"
                  />
                )}
              />
            </Grid>
          </Grid>
        </DialogContent>
        <DialogActions sx={{p: 2}}>
          <Button onClick={onClose} color="secondary" variant="outlined">
            Cancel
          </Button>
          <Button type="submit" variant="contained" color="primary" disabled={loading || !isDirty || !isValid}>
            {loading ? <CircularProgress size={24} color="inherit" /> : 'Create Account'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default CoACreateForm;
