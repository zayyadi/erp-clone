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
import { updateCoA, fetchCoAs, selectCoALoadingUpdate, selectCoAError, resetCoAErrors, selectCoAList } from './coaSlice';
import { AccountType, UpdateChartOfAccountRequest, ChartOfAccount } from './coaTypes';

interface CoAEditFormProps {
  open: boolean;
  onClose: () => void;
  account: ChartOfAccount | null;
}

// Schema for update might be slightly different (e.g., account_code not editable)
const schema = yup.object().shape({
  account_name: yup.string().required('Account name is required').min(3, 'Min 3 chars').max(100, 'Max 100 chars'),
  account_type: yup.string<AccountType>().oneOf(Object.values(AccountType)).required('Account type is required'),
  description: yup.string().optional().max(255, 'Max 255 chars'),
  is_active: yup.boolean().optional(),
  is_control_account: yup.boolean().optional(),
  parent_account_id: yup.string().nullable().optional(),
  // account_code is typically not editable after creation, so it's not in the update schema here.
  // If it were, it would need to be added.
});

const CoAEditForm: React.FC<CoAEditFormProps> = ({ open, onClose, account }) => {
  const dispatch: AppDispatch = useDispatch();
  const loading = useSelector(selectCoALoadingUpdate);
  const error = useSelector(selectCoAError);
  const allAccounts = useSelector(selectCoAList); // For Parent Account Autocomplete

  const { control, handleSubmit, reset, setValue, formState: { errors, isDirty, isValid } } = useForm<UpdateChartOfAccountRequest>({
    resolver: yupResolver(schema),
    defaultValues: {
      account_name: '',
      account_type: AccountType.Asset,
      description: '',
      is_active: true,
      is_control_account: false,
      parent_account_id: null,
    },
    mode: 'onChange',
  });

  useEffect(() => {
    if (open) {
      dispatch(resetCoAErrors());
      if (account) {
        reset({ // Use reset to update all form values and validation state
          account_name: account.account_name,
          account_type: account.account_type,
          description: account.description || '',
          is_active: account.is_active,
          is_control_account: account.is_control_account,
          parent_account_id: account.parent_account_id || null,
        });
      }
      if (allAccounts.length === 0) {
        dispatch(fetchCoAs({ page: 1, limit: 1000 }));
      }
    }
  }, [open, account, reset, dispatch, allAccounts.length]);

  const onSubmit: SubmitHandler<UpdateChartOfAccountRequest> = async (data) => {
    if (!account) return;

    const payload: UpdateChartOfAccountRequest = {
        ...data,
        is_active: data.is_active ?? true,
        is_control_account: data.is_control_account ?? false,
        parent_account_id: data.parent_account_id === '' ? null : data.parent_account_id,
    };

    const resultAction = await dispatch(updateCoA({ id: account.id, coaData: payload }));
    if (updateCoA.fulfilled.match(resultAction)) {
      onClose();
    }
  };

  if (!account) return null; // Or a loading/error state if account is expected but not provided

  return (
    <Dialog open={open} onClose={onClose} maxWidth="sm" fullWidth>
      <DialogTitle>Edit Account: {account.account_code} - {account.account_name}</DialogTitle>
      <form onSubmit={handleSubmit(onSubmit)}>
        <DialogContent>
          {error && <Alert severity="error" sx={{ mb: 2 }}>{error}</Alert>}
          <Grid container spacing={2} sx={{mt:1}}>
            {/* Account Code (Display Only) */}
            <Grid item xs={12} sm={6}>
              <TextField
                label="Account Code (Read-only)"
                value={account.account_code}
                variant="outlined"
                fullWidth
                InputProps={{
                  readOnly: true,
                }}
                disabled
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
                        options={allAccounts.filter(acc => acc.is_control_account && acc.id !== account.id)} // Exclude self
                        getOptionLabel={(option: ChartOfAccount) => `${option.account_code} - ${option.account_name}`}
                        value={allAccounts.find(acc => acc.id === field.value) || null}
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
        <DialogActions sx={{p:2}}>
          <Button onClick={onClose} color="secondary" variant="outlined">
            Cancel
          </Button>
          <Button type="submit" variant="contained" color="primary" disabled={loading || !isDirty || !isValid}>
            {loading ? <CircularProgress size={24} color="inherit" /> : 'Save Changes'}
          </Button>
        </DialogActions>
      </form>
    </Dialog>
  );
};

export default CoAEditForm;
