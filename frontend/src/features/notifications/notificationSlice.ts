import { createSlice, PayloadAction } from '@reduxjs/toolkit';
import { AlertColor } from '@mui/material/Alert';

interface NotificationState {
  open: boolean;
  message: string;
  severity: AlertColor;
  duration: number | null; // null for persistent, otherwise ms
}

const initialState: NotificationState = {
  open: false,
  message: '',
  severity: 'info', // Default severity
  duration: 6000, // Default duration
};

const notificationSlice = createSlice({
  name: 'notification',
  initialState,
  reducers: {
    showNotification(state, action: PayloadAction<{ message: string; severity?: AlertColor; duration?: number | null }>) {
      state.open = true;
      state.message = action.payload.message;
      state.severity = action.payload.severity || 'info';
      state.duration = action.payload.duration === undefined ? 6000 : action.payload.duration;
    },
    hideNotification(state) {
      state.open = false;
    },
  },
});

export const { showNotification, hideNotification } = notificationSlice.actions;
export default notificationSlice.reducer;
