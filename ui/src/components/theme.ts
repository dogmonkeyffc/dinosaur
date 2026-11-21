import { createTheme } from '@mui/material/styles';

const schemes = {
  light: {
    primary: '#426836',
    onPrimary: '#ffffff',
    surface: '#f8faf2',
    surfaceContainer: '#edf2e6',
    surfaceContainerHigh: '#e5ebde',
    onSurface: '#191d17',
    onSurfaceVariant: '#43483e',
    secondaryContainer: '#dce8d2',
    onSecondaryContainer: '#263420',
    outlineVariant: '#c4cbbd',
  },
  dark: {
    primary: '#a9d18f',
    onPrimary: '#18350f',
    surface: '#121510',
    surfaceContainer: '#1e231c',
    surfaceContainerHigh: '#282e25',
    onSurface: '#e1e6db',
    onSurfaceVariant: '#c4cbbd',
    secondaryContainer: '#384832',
    onSecondaryContainer: '#dce8d2',
    outlineVariant: '#43483e',
  },
};

function materialTheme(mode: keyof typeof schemes) {
  const colors = schemes[mode];
  return createTheme({
    cssVariables: true,
    palette: {
      mode,
      primary: { main: colors.primary, contrastText: colors.onPrimary },
      background: { default: colors.surface, paper: colors.surface },
      text: { primary: colors.onSurface, secondary: colors.onSurfaceVariant },
      divider: colors.outlineVariant,
    },
    components: {
      MuiCssBaseline: {
        styleOverrides: {
          ':root': {
            '--md-surface-container': colors.surfaceContainer,
            '--md-surface-container-high': colors.surfaceContainerHigh,
            '--md-secondary-container': colors.secondaryContainer,
            '--md-on-secondary-container': colors.onSecondaryContainer,
          },
        },
      },
    },
  });
}

export const lightTheme = materialTheme('light');
export const darkTheme = materialTheme('dark');
