import React from 'react';
import {
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, radius, spacing, type } from '@/theme/tokens';

function SectionHeader({ icon, title, action }: { icon: React.ComponentProps<typeof Ionicons>['name']; title: string; action?: string }) {
  return (
    <View style={s.sectionHeader}>
      <View style={s.sectionHeaderLeft}>
        <Ionicons name={icon} size={20} color={colors.primaryFixedDim} />
        <Text style={s.sectionTitle}>{title}</Text>
      </View>
      {action ? <Text style={s.sectionAction}>{action}</Text> : null}
    </View>
  );
}

function ContinueWatchingPlaceholder() {
  return (
    <View style={s.linkedAccountCard}>
      <Ionicons name="person-circle-outline" size={40} color={colors.outline} />
      <Text style={s.linkedAccountTitle}>Link your account to see Continue Watching</Text>
      <Text style={s.linkedAccountBody}>
        Connect your Jellyfin account to track what you've been watching.
      </Text>
    </View>
  );
}

function PosterPlaceholder({ title, subtitle }: { title: string; subtitle: string }) {
  return (
    <View style={s.posterCard}>
      <View style={s.posterImage}>
        <Ionicons name="film-outline" size={32} color={colors.outline} />
      </View>
      <Text style={s.posterTitle} numberOfLines={1}>{title}</Text>
      <Text style={s.posterSubtitle}>{subtitle}</Text>
    </View>
  );
}

const PLACEHOLDER_RECENTLY_ADDED = [
  { id: '1', title: 'Recently Added', subtitle: '—' },
  { id: '2', title: 'Coming Soon', subtitle: '—' },
];

export default function HomeScreen() {
  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Text style={s.topBarTitle}>Mortar</Text>
      </View>

      <ScrollView contentContainerStyle={s.scroll}>
        <View style={s.hero}>
          <Text style={s.heroGreeting}>Good evening.</Text>
          <Text style={s.heroSubtitle}>Your library is ready.</Text>
        </View>

        <View style={s.section}>
          <SectionHeader icon="play-circle-outline" title="Continue Watching" />
          <ContinueWatchingPlaceholder />
        </View>

        <View style={s.section}>
          <SectionHeader icon="star-outline" title="Recently Added" action="View all" />
          <ScrollView horizontal showsHorizontalScrollIndicator={false} contentContainerStyle={s.posterRow}>
            {PLACEHOLDER_RECENTLY_ADDED.map((item) => (
              <PosterPlaceholder key={item.id} title={item.title} subtitle={item.subtitle} />
            ))}
            <View style={s.posterCardAdd}>
              <Ionicons name="add-circle-outline" size={28} color={colors.outline} />
              <Text style={s.posterAddLabel}>Request</Text>
            </View>
          </ScrollView>
        </View>
      </ScrollView>
    </View>
  );
}

const s = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  topBar: {
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
    backgroundColor: colors.surface,
  },
  topBarTitle: {
    ...type.headlineLg,
    color: colors.primary,
    letterSpacing: -0.5,
  },
  scroll: {
    padding: spacing.gutter,
    gap: 32,
    paddingBottom: spacing.xl,
  },
  hero: {
    gap: 6,
    paddingTop: 8,
  },
  heroGreeting: {
    ...type.displayLg,
    color: colors.onSurface,
  },
  heroSubtitle: {
    ...type.bodyLg,
    color: colors.onSurfaceVariant,
  },
  section: {
    gap: 16,
  },
  sectionHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
  },
  sectionHeaderLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
  },
  sectionTitle: {
    ...type.headlineMd,
    color: colors.onSurface,
  },
  sectionAction: {
    ...type.labelMd,
    color: colors.primary,
  },
  linkedAccountCard: {
    backgroundColor: colors.surfaceContainer,
    borderRadius: radius.xl,
    padding: spacing.gutter,
    alignItems: 'center',
    gap: spacing.sm,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  linkedAccountTitle: {
    ...type.labelMd,
    color: colors.onSurface,
    textAlign: 'center',
  },
  linkedAccountBody: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
    textAlign: 'center',
  },
  posterRow: {
    gap: 16,
    paddingRight: spacing.gutter,
  },
  posterCard: {
    width: 120,
    gap: 8,
  },
  posterImage: {
    width: 120,
    height: 180,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.lg,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  posterTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  posterSubtitle: {
    ...type.labelSm,
    color: colors.outline,
  },
  posterCardAdd: {
    width: 120,
    height: 180,
    backgroundColor: colors.surfaceContainerLow,
    borderRadius: radius.lg,
    borderWidth: 1.5,
    borderStyle: 'dashed',
    borderColor: colors.outlineVariant,
    alignItems: 'center',
    justifyContent: 'center',
    gap: 8,
  },
  posterAddLabel: {
    ...type.labelSm,
    color: colors.outline,
  },
});
