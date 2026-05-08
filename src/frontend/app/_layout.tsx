import { Tabs } from 'expo-router';

/**
 * Root layout for the Mortar app.
 * Uses Expo Router's Tabs navigator for bottom navigation.
 * Tabs: Home, Search, My Requests, Health.
 */
export default function RootLayout() {
  return (
    <Tabs
      screenOptions={{
        headerShown: true,
        headerTitle: 'Mortar',
        tabBarActiveTintColor: '#3b82f6',
        tabBarInactiveTintColor: '#6b7280',
        tabBarStyle: {
          backgroundColor: '#fff',
          borderTopColor: '#e5e7eb',
        },
      }}
    >
      <Tabs.Screen
        name="index"
        options={{
          title: 'Home',
          tabBarLabel: 'Home',
          headerTitle: 'Mortar',
        }}
      />
      <Tabs.Screen
        name="search/index"
        options={{
          title: 'Search',
          tabBarLabel: 'Search',
          headerTitle: 'Search',
        }}
      />
      <Tabs.Screen
        name="requests/index"
        options={{
          title: 'My Requests',
          tabBarLabel: 'Requests',
          headerTitle: 'My Requests',
        }}
      />
      <Tabs.Screen
        name="health/index"
        options={{
          title: 'Health',
          tabBarLabel: 'Health',
          headerTitle: 'Service Health',
        }}
      />
    </Tabs>
  );
}
