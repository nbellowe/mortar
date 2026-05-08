import { Stack } from 'expo-router';

/**
 * Root layout for the Mortar app.
 * Uses Expo Router's Stack navigator as the top-level navigator.
 * Feature screens are registered as child routes within this layout.
 */
export default function RootLayout() {
  return (
    <Stack
      screenOptions={{
        headerShown: true,
        headerTitle: 'Mortar',
      }}
    />
  );
}
