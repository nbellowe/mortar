import React from 'react';
import {
  Platform,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  View,
  useWindowDimensions,
} from 'react-native';
import { Link, Tabs, usePathname } from 'expo-router';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius, spacing } from '@/theme/tokens';

const SIDEBAR_WIDTH = 280;
const DESKTOP_BREAKPOINT = 768;

type IoniconName = React.ComponentProps<typeof Ionicons>['name'];

type NavItem = { label: string; href: '/' | `/${string}`; icon: IoniconName; iconActive: IoniconName };

const SIDEBAR_ITEMS: NavItem[] = [
  { label: 'Home', href: '/', icon: 'home-outline', iconActive: 'home' },
  { label: 'Library', href: '/library', icon: 'film-outline', iconActive: 'film' },
  { label: 'Search', href: '/search', icon: 'search-outline', iconActive: 'search' },
  { label: 'My Requests', href: '/requests', icon: 'receipt-outline', iconActive: 'receipt' },
  { label: 'Activity', href: '/activity', icon: 'list-outline', iconActive: 'list' },
  { label: 'Downloads', href: '/downloads', icon: 'cloud-download-outline', iconActive: 'cloud-download' },
  { label: 'Health', href: '/health', icon: 'heart-outline', iconActive: 'heart' },
];

const TAB_ITEMS: NavItem[] = [
  { label: 'Home', href: '/', icon: 'home-outline', iconActive: 'home' },
  { label: 'Search', href: '/search', icon: 'search-outline', iconActive: 'search' },
  { label: 'Requests', href: '/requests', icon: 'receipt-outline', iconActive: 'receipt' },
  { label: 'Health', href: '/health', icon: 'heart-outline', iconActive: 'heart' },
];

function SidebarNav() {
  const pathname = usePathname();
  return (
    <View style={s.sidebar}>
      <View style={s.sidebarBrand}>
        <Text style={s.brandText}>Mortar</Text>
      </View>
      <ScrollView contentContainerStyle={s.navList}>
        {SIDEBAR_ITEMS.map(({ label, href, icon, iconActive }) => {
          const active = href === '/' ? pathname === '/' : pathname.startsWith(href);
          // Flatten avoids nested style arrays when Link asChild merges its own style prop
          const pressableStyle = StyleSheet.flatten([s.navItem, active && s.navItemActive]);
          const labelStyle = StyleSheet.flatten([s.navLabel, active && s.navLabelActive]);
          return (
            <Link key={href} href={href} replace asChild>
              <Pressable style={pressableStyle}>
                <Ionicons
                  name={active ? iconActive : icon}
                  size={22}
                  color={active ? colors.onPrimaryFixed : colors.onSurfaceVariant}
                />
                <Text style={labelStyle}>{label}</Text>
              </Pressable>
            </Link>
          );
        })}
      </ScrollView>
    </View>
  );
}

export default function RootLayout() {
  const { width } = useWindowDimensions();
  const isDesktop = Platform.OS === 'web' && width >= DESKTOP_BREAKPOINT;

  return (
    <View style={s.root}>
      {isDesktop && <SidebarNav />}
      <View style={s.main}>
        <Tabs
          // tabBar at the navigator level avoids React Navigation merging our
          // tabBarStyle with its own defaults into an array that expo-router rejects.
          tabBar={isDesktop ? () => null : undefined}
          screenOptions={{
            headerShown: false,
            tabBarStyle: s.tabBar,
            tabBarActiveTintColor: colors.primary,
            tabBarInactiveTintColor: colors.outline,
          }}
        >
          {TAB_ITEMS.map(({ label, href, icon, iconActive }) => {
            const name = href === '/' ? 'index' : `${href.slice(1)}/index`;
            return (
              <Tabs.Screen
                key={href}
                name={name}
                options={{
                  tabBarLabel: label,
                  tabBarIcon: ({ color, focused }) => (
                    <Ionicons name={focused ? iconActive : icon} size={24} color={color} />
                  ),
                }}
              />
            );
          })}
          <Tabs.Screen name="activity/index" options={{ tabBarButton: () => null }} />
          <Tabs.Screen name="downloads/index" options={{ tabBarButton: () => null }} />
        </Tabs>
      </View>
    </View>
  );
}

const s = StyleSheet.create({
  root: {
    flex: 1,
    flexDirection: 'row',
    backgroundColor: colors.background,
  },
  sidebar: {
    width: SIDEBAR_WIDTH,
    backgroundColor: colors.surfaceContainerLow,
    borderRightWidth: StyleSheet.hairlineWidth,
    borderRightColor: colors.outlineVariant,
  },
  sidebarBrand: {
    paddingHorizontal: spacing.gutter,
    paddingVertical: 20,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
  },
  brandText: {
    fontSize: 26,
    fontWeight: '700',
    color: colors.primary,
    letterSpacing: -0.5,
  },
  navList: {
    padding: spacing.sm,
    gap: 4,
  },
  navItem: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 16,
    paddingVertical: 12,
    paddingHorizontal: 16,
    borderRadius: radius.full,
  },
  navItemActive: {
    backgroundColor: colors.primaryFixed,
  },
  navLabel: {
    fontSize: 14,
    fontWeight: '600',
    color: colors.onSurfaceVariant,
  },
  navLabelActive: {
    color: colors.onPrimaryFixed,
    fontWeight: '700',
  },
  main: {
    flex: 1,
    backgroundColor: colors.background,
  },
  tabBar: {
    backgroundColor: colors.surfaceContainerLowest,
    borderTopColor: colors.outlineVariant,
    borderTopWidth: StyleSheet.hairlineWidth,
  },
});
