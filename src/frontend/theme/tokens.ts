export const colors = {
  background: '#1b110d',
  surface: '#1b110d',
  surfaceDim: '#1b110d',
  surfaceBright: '#433632',
  surfaceContainer: '#281d19',
  surfaceContainerLowest: '#150c09',
  surfaceContainerLow: '#241915',
  surfaceContainerHigh: '#332723',
  surfaceContainerHighest: '#3e322e',

  primary: '#ffb599',
  onPrimary: '#5a1c00',
  primaryFixed: '#ffdbce',
  onPrimaryFixed: '#3c1a00',
  primaryFixedDim: '#ffb599',
  primaryContainer: '#e17141',
  onPrimaryContainer: '#4f1800',

  secondary: '#eabe99',
  onSecondary: '#452a10',
  secondaryContainer: '#5f4024',
  onSecondaryContainer: '#d8ad89',

  tertiary: '#74d2f6',
  onTertiary: '#003544',
  tertiaryContainer: '#349cbe',
  onTertiaryContainer: '#002e3c',

  error: '#ffb4ab',
  onError: '#690005',
  errorContainer: '#93000a',
  onErrorContainer: '#ffdad6',

  onSurface: '#f3ded8',
  onSurfaceVariant: '#ddc0b6',
  onBackground: '#f3ded8',
  outline: '#a58b82',
  outlineVariant: '#56423b',

  inverseSurface: '#f3ded8',
  inverseOnSurface: '#3a2e29',
  inversePrimary: '#a04013',

  statusHealthy: '#81c784',
  statusDegraded: '#ffb74d',
  statusUnreachable: '#e57373',
  statusUnknown: '#a58b82',
} as const;

export const spacing = {
  xs: 4,
  sm: 12,
  base: 8,
  md: 24,
  lg: 48,
  xl: 80,
  gutter: 24,
  marginMobile: 16,
  marginDesktop: 64,
} as const;

export const radius = {
  sm: 4,
  md: 8,
  lg: 12,
  xl: 16,
  full: 999,
} as const;

export const type = {
  displayLg: { fontSize: 48, lineHeight: 53, fontWeight: '700' as const },
  headlineLg: { fontSize: 32, lineHeight: 38, fontWeight: '600' as const },
  headlineLgMobile: { fontSize: 28, lineHeight: 34, fontWeight: '600' as const },
  headlineMd: { fontSize: 24, lineHeight: 31, fontWeight: '600' as const },
  bodyLg: { fontSize: 18, lineHeight: 29, fontWeight: '400' as const },
  bodyMd: { fontSize: 16, lineHeight: 24, fontWeight: '400' as const },
  labelMd: { fontSize: 14, lineHeight: 20, fontWeight: '600' as const },
  labelSm: { fontSize: 12, lineHeight: 14, fontWeight: '700' as const },
} as const;
