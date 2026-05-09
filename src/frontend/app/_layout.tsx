import React from "react";
import {
  Platform,
  ActivityIndicator,
  Pressable,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
  useWindowDimensions,
} from "react-native";
import { Link, Tabs, usePathname } from "expo-router";
import { Ionicons } from "@expo/vector-icons";
import { AuthProvider, useAuth } from "../components/auth-context";
import { LoginScreen } from "../components/login-screen";
import { colors, radius, spacing } from "@/theme/tokens";

const SIDEBAR_WIDTH = 280;
const DESKTOP_BREAKPOINT = 768;

type IoniconName = React.ComponentProps<typeof Ionicons>["name"];

type NavItem = {
  label: string;
  href: "/" | `/${string}`;
  icon: IoniconName;
  iconActive: IoniconName;
};

const SIDEBAR_ITEMS: NavItem[] = [
  { label: "Home", href: "/", icon: "home-outline", iconActive: "home" },
  {
    label: "Library",
    href: "/library",
    icon: "film-outline",
    iconActive: "film",
  },
  {
    label: "Search",
    href: "/search",
    icon: "search-outline",
    iconActive: "search",
  },
  {
    label: "Requests",
    href: "/requests",
    icon: "receipt-outline",
    iconActive: "receipt",
  },
  {
    label: "Activity",
    href: "/activity",
    icon: "list-outline",
    iconActive: "list",
  },
  {
    label: "Downloads",
    href: "/downloads",
    icon: "cloud-download-outline",
    iconActive: "cloud-download",
  },
  {
    label: "Health",
    href: "/health",
    icon: "heart-outline",
    iconActive: "heart",
  },
];

const TAB_ITEMS: NavItem[] = [
  { label: "Home", href: "/", icon: "home-outline", iconActive: "home" },
  {
    label: "Search",
    href: "/search",
    icon: "search-outline",
    iconActive: "search",
  },
  {
    label: "Requests",
    href: "/requests",
    icon: "receipt-outline",
    iconActive: "receipt",
  },
  {
    label: "Health",
    href: "/health",
    icon: "heart-outline",
    iconActive: "heart",
  },
];

function SidebarNav({ canViewHealth }: { canViewHealth: boolean }) {
  const pathname = usePathname();
  const { user, logout } = useAuth();
  const items = SIDEBAR_ITEMS.filter(
    (item) => canViewHealth || item.href !== "/health",
  );
  return (
    <View style={s.sidebar}>
      <View style={s.sidebarBrand}>
        <Text style={s.brandText}>Mortar</Text>
        {user ? <Text style={s.sidebarUser}>{user.username}</Text> : null}
      </View>
      <ScrollView contentContainerStyle={s.navList}>
        {items.map(({ label, href, icon, iconActive }) => {
          const active =
            href === "/" ? pathname === "/" : pathname.startsWith(href);
          // Flatten avoids nested style arrays when Link asChild merges its own style prop
          const pressableStyle = StyleSheet.flatten([
            s.navItem,
            active && s.navItemActive,
          ]);
          const labelStyle = StyleSheet.flatten([
            s.navLabel,
            active && s.navLabelActive,
          ]);
          return (
            <Link key={href} href={href} replace asChild>
              <Pressable style={pressableStyle}>
                <Ionicons
                  name={active ? iconActive : icon}
                  size={22}
                  color={
                    active ? colors.onPrimaryFixed : colors.onSurfaceVariant
                  }
                />
                <Text style={labelStyle}>{label}</Text>
              </Pressable>
            </Link>
          );
        })}
      </ScrollView>
      <TouchableOpacity
        style={s.signOutBtn}
        onPress={() => {
          void logout();
        }}
      >
        <Ionicons
          name="log-out-outline"
          size={18}
          color={colors.onSurfaceVariant}
        />
        <Text style={s.signOutText}>Sign out</Text>
      </TouchableOpacity>
    </View>
  );
}

function RootLayoutContent() {
  const { width } = useWindowDimensions();
  const { loading, user } = useAuth();
  const isDesktop = Platform.OS === "web" && width >= DESKTOP_BREAKPOINT;
  const canViewHealth = user?.role === "admin";

  if (loading) {
    return (
      <View style={s.loadingWrap}>
        <ActivityIndicator size="large" color={colors.primary} />
      </View>
    );
  }

  if (!user) {
    return <LoginScreen />;
  }

  const tabItems = TAB_ITEMS.filter(
    (item) => canViewHealth || item.href !== "/health",
  );

  return (
    <View style={s.root}>
      {isDesktop && <SidebarNav canViewHealth={canViewHealth} />}
      <View style={s.main}>
        <Tabs
          tabBar={isDesktop ? () => null : undefined}
          screenOptions={{
            headerShown: false,
            tabBarStyle: s.tabBar,
            tabBarActiveTintColor: colors.primary,
            tabBarInactiveTintColor: colors.outline,
          }}
        >
          {tabItems.map(({ label, href, icon, iconActive }) => {
            const name = href === "/" ? "index" : `${href.slice(1)}/index`;
            return (
              <Tabs.Screen
                key={href}
                name={name}
                options={{
                  tabBarLabel: label,
                  tabBarIcon: ({ color, focused }) => (
                    <Ionicons
                      name={focused ? iconActive : icon}
                      size={24}
                      color={color}
                    />
                  ),
                }}
              />
            );
          })}
          <Tabs.Screen
            name="activity/index"
            options={{ tabBarButton: () => null }}
          />
          <Tabs.Screen
            name="downloads/index"
            options={{ tabBarButton: () => null }}
          />
          <Tabs.Screen
            name="library/index"
            options={{ tabBarButton: () => null }}
          />
          <Tabs.Screen
            name="health/index"
            options={{ tabBarButton: () => null }}
          />
        </Tabs>
      </View>
    </View>
  );
}

export default function RootLayout() {
  return (
    <AuthProvider>
      <RootLayoutContent />
    </AuthProvider>
  );
}

const s = StyleSheet.create({
  root: {
    flex: 1,
    flexDirection: "row",
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
    fontWeight: "700",
    color: colors.primary,
    letterSpacing: -0.5,
  },
  sidebarUser: {
    marginTop: 6,
    fontSize: 13,
    color: colors.onSurfaceVariant,
  },
  navList: {
    padding: spacing.sm,
    gap: 4,
    flexGrow: 1,
  },
  navItem: {
    flexDirection: "row",
    alignItems: "center",
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
    fontWeight: "600",
    color: colors.onSurfaceVariant,
  },
  navLabelActive: {
    color: colors.onPrimaryFixed,
    fontWeight: "700",
  },
  main: {
    flex: 1,
    backgroundColor: colors.background,
  },
  signOutBtn: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.xs,
    margin: spacing.sm,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    borderRadius: radius.full,
    backgroundColor: colors.surfaceContainer,
  },
  signOutText: {
    fontSize: 13,
    fontWeight: "600",
    color: colors.onSurfaceVariant,
  },
  loadingWrap: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    backgroundColor: colors.background,
  },
  tabBar: {
    backgroundColor: colors.surfaceContainerLowest,
    borderTopColor: colors.outlineVariant,
    borderTopWidth: StyleSheet.hairlineWidth,
  },
});
