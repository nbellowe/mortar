/**
 * Home screen — Mortar dashboard.
 *
 * Shows a health badge (when any service is unreachable), a Continue Watching
 * placeholder (pending auth), and a Recently Added row backed by live API data.
 * Data is fetched once on mount — no auto-polling (this is not a live operational view).
 * Spec: specs/features/home.md
 */

import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  Image,
  ScrollView,
  StyleSheet,
  Text,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';

import { getHome, HomeResponse } from '../api/home';
import type { MediaItem } from '../types/plugin';
import { colors, radius, spacing, type } from '@/theme/tokens';

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

function SectionHeader({
  icon,
  title,
  action,
}: {
  icon: React.ComponentProps<typeof Ionicons>['name'];
  title: string;
  action?: string;
}) {
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

/** Subtle setup prompt shown in place of a Continue Watching section header. */
function ContinueWatchingPlaceholder() {
  // TODO: replace with real continue-watching data once auth is implemented (ADR 0005).
  return (
    <View style={s.linkedAccountCard}>
      <Ionicons name="person-circle-outline" size={40} color={colors.outline} />
      <Text style={s.linkedAccountTitle}>Link your account to see Continue Watching</Text>
      <Text style={s.linkedAccountBody}>
        Connect your Jellyfin account to track what you&apos;ve been watching.
      </Text>
    </View>
  );
}

/** Empty poster-shaped placeholder used during loading (skeleton stand-in). */
function PosterSkeleton() {
  return (
    <View style={s.posterCard}>
      <View style={[s.posterImage, s.posterSkeleton]} />
      <View style={s.skeletonTextLine} />
      <View style={[s.skeletonTextLine, s.skeletonTextShort]} />
    </View>
  );
}

/** Single poster card backed by a real MediaItem. */
function PosterCard({ item }: { item: MediaItem }) {
  const [imgFailed, setImgFailed] = useState(false);

  return (
    <View style={s.posterCard}>
      <View style={s.posterImage}>
        {item.poster_url && !imgFailed ? (
          <Image
            source={{ uri: item.poster_url }}
            style={s.posterImageFill}
            resizeMode="cover"
            onError={() => setImgFailed(true)}
          />
        ) : (
          <Ionicons name="film-outline" size={32} color={colors.outline} />
        )}
      </View>
      <Text style={s.posterTitle} numberOfLines={1}>{item.title}</Text>
      <Text style={s.posterSubtitle}>{item.year ?? '—'}</Text>
    </View>
  );
}

/** The "+" request card always shown at the end of the Recently Added row. */
function RequestCard() {
  return (
    <View style={s.posterCardAdd}>
      <Ionicons name="add-circle-outline" size={28} color={colors.outline} />
      <Text style={s.posterAddLabel}>Request</Text>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Section: Health badge
// ---------------------------------------------------------------------------

function HealthBadge({ summary }: { summary: HomeResponse['health_summary'] }) {
  if (!summary.any_unreachable) return null;

  return (
    <View style={s.healthBadge}>
      <Ionicons name="warning" size={14} color={colors.statusUnreachable} />
      <Text style={s.healthBadgeText}>
        {summary.unreachable_count} service{summary.unreachable_count !== 1 ? 's' : ''} unreachable
      </Text>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Section: Recently Added
// ---------------------------------------------------------------------------

function RecentlyAddedRow({
  loading,
  error,
  items,
}: {
  loading: boolean;
  error: string | null;
  items: MediaItem[];
}) {
  if (loading) {
    return (
      <ScrollView
        horizontal
        showsHorizontalScrollIndicator={false}
        contentContainerStyle={s.posterRow}
      >
        <PosterSkeleton />
        <PosterSkeleton />
        <PosterSkeleton />
      </ScrollView>
    );
  }

  if (error) {
    return (
      <View style={s.inlineError}>
        <Ionicons name="alert-circle-outline" size={16} color={colors.error} />
        <Text style={s.inlineErrorText}>Could not load recent additions</Text>
      </View>
    );
  }

  if (items.length === 0) {
    return (
      <View style={s.emptyRow}>
        <Text style={s.emptyRowText}>No recent additions</Text>
      </View>
    );
  }

  return (
    <ScrollView
      horizontal
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={s.posterRow}
    >
      {items.map((item) => (
        <PosterCard key={item.id} item={item} />
      ))}
      <RequestCard />
    </ScrollView>
  );
}

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

export default function HomeScreen() {
  const [recentlyAdded, setRecentlyAdded] = useState<MediaItem[]>([]);
  const [healthSummary, setHealthSummary] = useState<HomeResponse['health_summary'] | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const load = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    try {
      const data = await getHome(controller.signal);
      setRecentlyAdded(data.recently_added);
      setHealthSummary(data.health_summary);
      setError(null);
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      setError(err instanceof Error ? err.message : 'Failed to load home data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
    return () => { abortRef.current?.abort(); };
  }, [load]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Text style={s.topBarTitle}>Mortar</Text>
      </View>

      <ScrollView contentContainerStyle={s.scroll}>
        {/* Hero */}
        <View style={s.hero}>
          <Text style={s.heroGreeting}>Good evening.</Text>
          <Text style={s.heroSubtitle}>Your library is ready.</Text>
          {healthSummary ? <HealthBadge summary={healthSummary} /> : null}
        </View>

        {/* Continue Watching — placeholder until auth is implemented */}
        <View style={s.section}>
          <ContinueWatchingPlaceholder />
        </View>

        {/* Recently Added */}
        <View style={s.section}>
          <SectionHeader icon="star-outline" title="Recently Added" action="View all" />
          <RecentlyAddedRow
            loading={loading}
            error={error}
            items={recentlyAdded}
          />
        </View>
      </ScrollView>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

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

  // Health badge — shown only when any_unreachable is true
  healthBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
    alignSelf: 'flex-start',
    marginTop: spacing.base,
    paddingHorizontal: spacing.sm,
    paddingVertical: 5,
    backgroundColor: colors.surfaceContainer,
    borderRadius: radius.full,
    borderWidth: 1,
    borderColor: colors.statusUnreachable,
  },
  healthBadgeText: {
    ...type.labelSm,
    color: colors.statusUnreachable,
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

  // Continue Watching placeholder card
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

  // Poster row
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
    overflow: 'hidden',
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  posterImageFill: {
    width: '100%',
    height: '100%',
  },
  posterTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  posterSubtitle: {
    ...type.labelSm,
    color: colors.outline,
  },

  // Skeleton placeholders
  posterSkeleton: {
    backgroundColor: colors.surfaceContainerHigh,
    opacity: 0.5,
  },
  skeletonTextLine: {
    height: 12,
    width: 90,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.sm,
    opacity: 0.5,
  },
  skeletonTextShort: {
    width: 50,
  },

  // Request card at end of row
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

  // Empty / error inline states for Recently Added
  emptyRow: {
    paddingVertical: spacing.md,
  },
  emptyRowText: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
  },
  inlineError: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.xs,
    paddingVertical: spacing.sm,
  },
  inlineErrorText: {
    ...type.bodyMd,
    color: colors.error,
  },
});
