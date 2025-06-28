import React from 'react';
import { useSelector, useDispatch } from 'react-redux';
import Snackbar from '@mui/material/Snackbar';
import Alert, { AlertColor } from '@mui/material/Alert';
import { RootState, AppDispatch } from '../../store';
import { hideNotification } from './notificationSlice';

const GlobalSnackbar: React.FC = () => {
  const dispatch: AppDispatch = useDispatch();
  const { open, message, severity, duration } = useSelector((state: RootState) => state.notification);

  const handleClose = (event?: React.SyntheticEvent | Event, reason?: string) => {
    if (reason === 'clickaway') {
      return;
    }
    dispatch(hideNotification());
  };

  return (
    <Snackbar
      open={open}
      autoHideDuration={duration}
      onClose={handleClose}
      anchorOrigin={{ vertical: 'top', horizontal: 'center' }}
    >
      {/* Wrapping Alert in a Snackbar like this is standard.
          The key is to ensure Alert's props are correctly passed.
          Snackbar will not render if message is empty or open is false.
      */}
      <Alert onClose={handleClose} severity={severity as AlertColor} sx={{ width: '100%' }} variant="filled">
        {message}
      </Alert>
    </Snackbar>
  );
};

export default GlobalSnackbar;
