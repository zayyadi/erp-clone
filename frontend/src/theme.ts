import { createTheme } from '@mui/material/styles';
import { red } from '@mui/material/colors';

// A custom theme for this app
const theme = createTheme({
  palette: {
    primary: {
      main: '#556cd6',
    },
    secondary: {
      main: '#19857b',
    },
    error: {
      main: red.A400,
    },
    background: {
      default: '#fff',
    },
  },
  typography: {
    fontFamily: [
      '-apple-system',
      'BlinkMacSystemFont',
      '"Segoe UI"',
      'Roboto',
      '"Helvetica Neue"',
      'Arial',
      'sans-serif',
      '"Apple Color Emoji"',
      '"Segoe UI Emoji"',
      '"Segoe UI Symbol"',
    ].join(','),
    h4: {
      // Default (desktop)
      fontSize: '2.125rem', // MUI default: 2.125rem (34px)
      fontWeight: 400,
      lineHeight: 1.235,
      letterSpacing: '0.00735em',
      // Mobile
      '@media (max-width:600px)': { // MUI 'sm' breakpoint
        fontSize: '1.75rem', // Approx 28px
      },
    },
    h6: {
      // Default (desktop) for AppBar title
      fontSize: '1.25rem', // MUI default: 1.25rem (20px)
      // Mobile
      '@media (max-width:600px)': {
        fontSize: '1.1rem', // Approx 18px
      },
    }
  },
  // Example of component default props
  // components: {
  //   MuiTextField: {
  //     defaultProps: {
  //       size: 'small',
  //       variant: 'outlined',
  //     },
  //   },
  //   MuiButton: {
  //     defaultProps: {
  //       size: 'small',
  //     }
  //   }
  // }
});

export default theme;
