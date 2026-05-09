/**
 * Activity screen — Cross-Service Event Timeline.
 *
 * Shows a unified feed of activity events from all configured plugins.
 * Auto-refreshes every 30 seconds (ADR 0003 shared cadence).
 * Supports client-side filtering by event type category.
 * Spec: specs/features/activity-feed.md
 */

import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';

import { getActivity } from '../../api/activity';
import type { ActivityEvent, ActivityEventType } from '../../types/plugin';
import { colors, radius, spacing, type } from '@/theme/tokens';

// ---------------------------------------------------------------------------
// Types
// ---------------------------------------------------------------------------

type FilterKey = 'all' | 'downloaded' | 'added' | 'requested' | 'other';

interface FilterOption {
  key: FilterKey;
  label: string;
}

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const REFRESH_INTERVAL_MS = 30_000;

const FILTERS: FilterOption[] = [
  { key: 'all', label: 'All' },
  { key: 'downloaded', label: 'Downloaded' },
  { key: 'added', label: 'Added' },
  { key: 'requested', label: 'Requested' },
  { key: 'other', label: 'Other' },
];

const EVENT_ICON: Record<ActivityEventType, React.ComponentProps<typeof Ionicons>['name']> = {
  downloaded: 'cloud-download',
  added_to_library: 'library',
  requested: 'add-circle',
  approved: 'checkmark-circle',
  declined: 'close-circle',
  failed: 'warning',
  deleted: 'trash',
};

const EVENT_ICON_COLOR: Record<ActivityEventType, string> = {
  downloaded: colors.tertiary,
  added_to_library: colors.statusHealthy,
  requested: colors.primary,
  approved: colors.statusHealthy,
  declined: colors.statusUnreachable,
  failed: colors.statusDegraded,
  deleted: colors.outline,
};

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

function matchesFilter(event: ActivityEvent, filter: FilterKey): boolean {
  if (filter === 'all') return true;
  if (filter === 'downloaded') return event.type === 'downloaded';
  if (filter === 'added') return event.type === 'added_to_library';
  if (filter === 'requested') {
    return event.type === 'requested' || event.type === 'approved' || event.type === 'declined';
  }
  // 'other' = everything that doesn't match the above categories
  return (
    event.type !== 'downloaded' &&
    event.type !== 'added_to_library' &&
    event.type !== 'requested' &&
    event.type !== 'approved' &&
    event.type !== 'declined'
  );
}

function relativeTime(iso: string, now: number): string {
  try {
    const diffMs = now - new Date(iso).getTime();
    const diffSec = Math.floor(diffMs / 1000);
    if (diffSec < 60) return 'just now';
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffHr = Math.floor(diffMin / 60);
    if (diffHr < 24) return `${diffHr}h ago`;
    return `${Math.floor(diffHr / 24)}d ago`;
  } catch {
    return iso;
  }
}

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

function FilterChips({
  active,
  onSelect,
}: {
  active: FilterKey;
  onSelect: (key: FilterKey) => void;
}) {
  return (
    <ScrollView
      horizontal
      showsHorizontalScrollIndicator={false}
      contentContainerStyle={s.chipsRow}
    >
      {FILTERS.map((f) => (
        <TouchableOpacity
          key={f.key}
          style={[s.chip, active === f.key && s.chipActive]}
          onPress={() => onSelect(f.key)}
          accessibilityRole="button"
          accessibilityState={{ selected: active === f.key }}
        >
          <Text style={[s.chipText, active === f.key && s.chipTextActive]}>
            {f.label}
          </Text>
        </TouchableOpacity>
      ))}
    </ScrollView>
  );
}

function FailedPluginsBanner({ plugins }: { plugins: string[] }) {
  if (plugins.length === 0) return null;
  return (
    <View style={s.warningBanner}>
      <Ionicons name="warning-outline" size={14} color={colors.statusDegraded} />
      <Text style={s.warningBannerText} numberOfLines={2}>
        Some services unavailable: {plugins.join(', ')}
      </Text>
    </View>
  );
}

function EventRow({ event, now }: { event: ActivityEvent; now: number }) {
  const iconName = EVENT_ICON[event.type];
  const iconColor = EVENT_ICON_COLOR[event.type];
  return (
    <View style={s.row}>
      <View style={[s.rowIconWrap, { borderColor: `${iconColor}44` }]}>
        <Ionicons name={iconName} size={18} color={iconColor} />
      </View>
      <View style={s.rowBody}>
        <Text style={s.rowMessage} numberOfLines={3}>
          {event.message}
        </Text>
        <View style={s.rowMeta}>
          <View style={s.pluginBadge}>
            <Text style={s.pluginBadgeText} numberOfLines={1}>
              {event.source_plugin}
            </Text>
          </View>
          <Text style={s.rowTimestamp}>{relativeTime(event.timestamp, now)}</Text>
        </View>
      </View>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

export default function ActivityScreen() {
  const [events, setEvents] = useState<ActivityEvent[]>([]);
  const [failedPlugins, setFailedPlugins] = useState<string[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [filter, setFilter] = useState<FilterKey>('all');
  const [now, setNow] = useState<number>(Date.now());
  const abortRef = useRef<AbortController | null>(null);

  const load = useCallback(async (isInitial = false) => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    try {
      const data = await getActivity(undefined, controller.signal);
      setEvents(data.events);
      setFailedPlugins(data.failed_plugins);
      setNow(Date.now());
      setError(null);
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      if (isInitial) {
        setError(err instanceof Error ? err.message : 'Failed to load activity');
      }
      // On poll failures, silently retain existing data
    } finally {
      if (isInitial) setLoading(false);
    }
  }, []);

  // Initial load
  useEffect(() => {
    void load(true);
    return () => { abortRef.current?.abort(); };
  }, [load]);

  // Polling — also refreshes `now` so relative timestamps stay current
  useEffect(() => {
    const id = setInterval(() => {
      void load(false);
    }, REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [load]);

  const filteredEvents = events.filter((e) => matchesFilter(e, filter));

  return (
    <View style={s.container}>
      {/* Top bar */}
      <View style={s.topBar}>
        <Ionicons name="list" size={22} color={colors.primaryFixedDim} />
        <Text style={s.topBarTitle}>Activity</Text>
      </View>

      {loading ? (
        <View style={s.centered}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : error ? (
        <View style={s.centered}>
          <View style={s.errorBanner}>
            <Ionicons name="warning-outline" size={20} color={colors.error} />
            <Text style={s.errorText}>{error}</Text>
          </View>
          <TouchableOpacity
            style={s.retryBtn}
            onPress={() => { setLoading(true); void load(true); }}
          >
            <Text style={s.retryBtnText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <>
          {/* Filter chips */}
          <FilterChips active={filter} onSelect={setFilter} />

          {/* Degraded services notice */}
          <FailedPluginsBanner plugins={failedPlugins} />

          {/* Event list */}
          <FlatList
            data={filteredEvents}
            keyExtractor={(item) => item.id}
            renderItem={({ item }) => <EventRow event={item} now={now} />}
            ItemSeparatorComponent={() => <View style={s.separator} />}
            contentContainerStyle={s.listContent}
            ListEmptyComponent={
              <View style={s.centered}>
                <Ionicons name="list-outline" size={48} color={colors.outlineVariant} />
                <Text style={s.emptyTitle}>
                  {filter === 'all' ? 'No activity yet' : 'No matching events'}
                </Text>
                <Text style={s.emptyBody}>
                  {filter === 'all'
                    ? 'Activity from your connected services will appear here.'
                    : 'Try a different filter to see more events.'}
                </Text>
              </View>
            }
          />
        </>
      )}
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
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
    backgroundColor: colors.surface,
  },
  topBarTitle: {
    ...type.headlineLg,
    color: colors.onSurface,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing.gutter,
    gap: spacing.sm,
  },
  // Filter chips
  chipsRow: {
    flexDirection: 'row',
    gap: spacing.base,
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.sm,
  },
  chip: {
    paddingHorizontal: spacing.sm,
    paddingVertical: 6,
    borderRadius: radius.full,
    borderWidth: 1,
    borderColor: colors.outlineVariant,
    backgroundColor: colors.surfaceContainerLow,
  },
  chipActive: {
    backgroundColor: colors.primaryContainer,
    borderColor: colors.primary,
  },
  chipText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  chipTextActive: {
    color: colors.onPrimaryContainer,
  },
  // Failed plugins warning
  warningBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
    marginHorizontal: spacing.gutter,
    marginBottom: spacing.base,
    paddingHorizontal: spacing.sm,
    paddingVertical: 6,
    borderRadius: radius.md,
    backgroundColor: `${colors.statusDegraded}22`,
    borderWidth: 1,
    borderColor: `${colors.statusDegraded}44`,
  },
  warningBannerText: {
    ...type.labelSm,
    color: colors.statusDegraded,
    flex: 1,
    fontWeight: '400',
  },
  // List
  listContent: {
    paddingBottom: spacing.xl,
  },
  separator: {
    height: StyleSheet.hairlineWidth,
    backgroundColor: colors.outlineVariant,
    marginHorizontal: spacing.gutter,
  },
  // Event row
  row: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.gutter,
    gap: spacing.sm,
  },
  rowIconWrap: {
    width: 38,
    height: 38,
    borderRadius: radius.lg,
    borderWidth: 1,
    backgroundColor: colors.surfaceContainerHigh,
    alignItems: 'center',
    justifyContent: 'center',
    flexShrink: 0,
  },
  rowBody: {
    flex: 1,
    gap: 4,
  },
  rowMessage: {
    ...type.labelMd,
    color: colors.onSurface,
    fontWeight: '400',
  },
  rowMeta: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
    flexWrap: 'wrap',
  },
  pluginBadge: {
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: radius.sm,
    backgroundColor: colors.surfaceContainerHighest,
  },
  pluginBadgeText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  rowTimestamp: {
    fontSize: 11,
    lineHeight: 14,
    fontWeight: '400',
    color: colors.outline,
  },
  // Error state
  errorBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
    backgroundColor: colors.errorContainer,
    paddingHorizontal: spacing.md,
    paddingVertical: 10,
    borderRadius: radius.lg,
  },
  errorText: {
    ...type.bodyMd,
    color: colors.onErrorContainer,
    flex: 1,
  },
  retryBtn: {
    paddingVertical: 10,
    paddingHorizontal: 20,
    backgroundColor: colors.primaryContainer,
    borderRadius: radius.full,
  },
  retryBtnText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
  // Empty state
  emptyTitle: {
    ...type.headlineMd,
    color: colors.onSurface,
    textAlign: 'center',
  },
  emptyBody: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
    textAlign: 'center',
  },
});
