/**
 * Downloads screen — Unified Download Queue.
 *
 * Shows active, queued, and failed downloads across all plugins with
 * downloads.read capability. Auto-refreshes every 10 seconds.
 * Spec: specs/features/download-queue.md
 */

import React, { useCallback, useEffect, useRef, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';

import { getDownloads, DownloadsResponse } from '../../api/downloads';
import type { DownloadItem } from '../../types/plugin';
import { colors, radius, spacing, type } from '@/theme/tokens';

// ---------------------------------------------------------------------------
// Formatter helpers
// ---------------------------------------------------------------------------

function formatBytes(bytes: number): string {
  if (bytes >= 1_073_741_824) {
    return `${(bytes / 1_073_741_824).toFixed(1)} GB`;
  }
  if (bytes >= 1_048_576) {
    return `${(bytes / 1_048_576).toFixed(0)} MB`;
  }
  return `${(bytes / 1_024).toFixed(0)} KB`;
}

function formatSpeed(bytesPerSec: number): string {
  if (bytesPerSec >= 1_048_576) {
    return `${(bytesPerSec / 1_048_576).toFixed(1)} MB/s`;
  }
  return `${(bytesPerSec / 1_024).toFixed(0)} KB/s`;
}

function formatEta(seconds: number | null): string {
  if (seconds === null) return 'unknown';
  if (seconds < 3_600) {
    return `${Math.ceil(seconds / 60)}m left`;
  }
  return `${Math.floor(seconds / 3_600)}h left`;
}

// ---------------------------------------------------------------------------
// Status mappings
// ---------------------------------------------------------------------------

type DownloadStatus = DownloadItem['status'];

const STATUS_COLOR: Record<DownloadStatus, string> = {
  downloading: colors.primary,
  queued: colors.onSurfaceVariant,
  processing: colors.tertiary,
  paused: colors.statusDegraded,
  failed: colors.statusUnreachable,
};

const STATUS_LABEL: Record<DownloadStatus, string> = {
  downloading: 'Downloading',
  queued: 'Queued',
  processing: 'Processing',
  paused: 'Paused',
  failed: 'Failed',
};

const STATUS_ICON: Record<DownloadStatus, React.ComponentProps<typeof Ionicons>['name']> = {
  downloading: 'cloud-download',
  queued: 'time-outline',
  processing: 'sync',
  paused: 'pause-circle-outline',
  failed: 'close-circle-outline',
};

// ---------------------------------------------------------------------------
// Sub-components
// ---------------------------------------------------------------------------

function FailedPluginsBanner({ names }: { names: string[] }) {
  if (names.length === 0) return null;
  return (
    <View style={s.warnBanner}>
      <Ionicons name="warning-outline" size={16} color={colors.statusDegraded} />
      <Text style={s.warnBannerText}>
        Some services unavailable: {names.join(', ')}
      </Text>
    </View>
  );
}

function StatusBadge({ status }: { status: DownloadStatus }) {
  const color = STATUS_COLOR[status];
  return (
    <View style={[s.statusBadge, { borderColor: color }]}>
      <Ionicons name={STATUS_ICON[status]} size={12} color={color} />
      <Text style={[s.statusBadgeText, { color }]}>{STATUS_LABEL[status]}</Text>
    </View>
  );
}

function ProgressBar({ progress }: { progress: number }) {
  const pct = Math.min(Math.max(progress, 0), 1);
  return (
    <View style={s.progressTrack}>
      <View style={[s.progressFill, { width: `${pct * 100}%` as `${number}%` }]} />
    </View>
  );
}

function DownloadRow({ item }: { item: DownloadItem }) {
  const isActive = item.status === 'downloading' || item.status === 'processing';
  const showProgress = isActive;

  return (
    <View style={[s.card, item.status === 'failed' && s.cardFailed]}>
      {/* Header row: name + status badge */}
      <View style={s.cardHeader}>
        <Text style={s.cardName} numberOfLines={2}>{item.name}</Text>
        <StatusBadge status={item.status} />
      </View>

      {/* Progress bar for active items */}
      {showProgress && (
        <ProgressBar progress={item.progress} />
      )}

      {/* Footer row: meta information */}
      <View style={s.cardMeta}>
        <View style={s.cardMetaLeft}>
          {item.size_bytes > 0 && (
            <Text style={s.metaText}>{formatBytes(item.size_bytes)}</Text>
          )}
          {isActive && item.speed_bytes_s > 0 && (
            <Text style={s.metaText}>{formatSpeed(item.speed_bytes_s)}</Text>
          )}
          {isActive && (
            <Text style={s.metaText}>{formatEta(item.eta_seconds)}</Text>
          )}
          {item.status === 'failed' && (
            <Text style={s.failedNote}>See native app for details</Text>
          )}
        </View>
        {item.source_plugin != null && item.source_plugin !== '' && (
          <View style={s.sourcePill}>
            <Text style={s.sourcePillText}>{item.source_plugin}</Text>
          </View>
        )}
      </View>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Screen
// ---------------------------------------------------------------------------

const REFRESH_INTERVAL_MS = 10_000;

export default function DownloadsScreen() {
  const [data, setData] = useState<DownloadsResponse>({ items: [], failed_plugins: [] });
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const load = useCallback(async (isInitial = false) => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    try {
      const result = await getDownloads(controller.signal);
      setData(result);
      setError(null);
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      // Only surface the error on initial load; subsequent polls preserve
      // last-known data so the UI stays usable.
      if (isInitial) {
        setError(err instanceof Error ? err.message : 'Failed to load download queue');
      }
    } finally {
      if (isInitial) setLoading(false);
    }
  }, []);

  // Initial fetch
  useEffect(() => {
    void load(true);
    return () => { abortRef.current?.abort(); };
  }, [load]);

  // Polling — only after initial load succeeds so we do not double-fetch
  useEffect(() => {
    const id = setInterval(() => { void load(false); }, REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [load]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Ionicons name="cloud-download" size={22} color={colors.primaryFixedDim} />
        <Text style={s.topBarTitle}>Downloads</Text>
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
        <FlatList
          data={data.items}
          keyExtractor={(item) => item.id}
          ListHeaderComponent={
            <FailedPluginsBanner names={data.failed_plugins} />
          }
          renderItem={({ item }) => <DownloadRow item={item} />}
          ItemSeparatorComponent={() => <View style={s.separator} />}
          contentContainerStyle={s.listContent}
          ListEmptyComponent={
            <View style={s.centered}>
              <Ionicons name="cloud-download-outline" size={48} color={colors.outlineVariant} />
              <Text style={s.emptyTitle}>Nothing downloading right now</Text>
              <Text style={s.emptyBody}>
                Active and queued downloads from your connected services will appear here.
              </Text>
            </View>
          }
        />
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

  // Top bar
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

  // Centered state wrapper (loading / error / empty)
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing.gutter,
    gap: spacing.sm,
  },

  // Warning banner (failed plugins)
  warnBanner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
    backgroundColor: `${colors.statusDegraded}22`,
    borderWidth: 1,
    borderColor: `${colors.statusDegraded}44`,
    borderRadius: radius.lg,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    marginHorizontal: spacing.gutter,
    marginTop: spacing.md,
    marginBottom: spacing.base,
  },
  warnBannerText: {
    ...type.labelMd,
    color: colors.onSurface,
    flex: 1,
  },

  // List
  listContent: {
    paddingBottom: spacing.xl,
    paddingTop: spacing.base,
    gap: spacing.base,
  },
  separator: {
    height: StyleSheet.hairlineWidth,
    backgroundColor: colors.outlineVariant,
    marginHorizontal: spacing.gutter,
  },

  // Download card
  card: {
    backgroundColor: colors.surfaceContainer,
    marginHorizontal: spacing.gutter,
    borderRadius: radius.xl,
    padding: spacing.md,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
    gap: spacing.base,
  },
  cardFailed: {
    borderColor: `${colors.statusUnreachable}44`,
    backgroundColor: `${colors.statusUnreachable}11`,
  },
  cardHeader: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    gap: spacing.sm,
  },
  cardName: {
    ...type.labelMd,
    color: colors.onSurface,
    flex: 1,
  },

  // Status badge chip
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: radius.full,
    borderWidth: 1,
    flexShrink: 0,
  },
  statusBadgeText: {
    ...type.labelSm,
  },

  // Progress bar
  progressTrack: {
    height: 4,
    backgroundColor: colors.outlineVariant,
    borderRadius: radius.full,
    overflow: 'hidden',
  },
  progressFill: {
    height: '100%',
    backgroundColor: colors.primary,
    borderRadius: radius.full,
  },

  // Card footer / meta row
  cardMeta: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    gap: spacing.sm,
  },
  cardMetaLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.sm,
    flex: 1,
    flexWrap: 'wrap',
  },
  metaText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
    fontWeight: '400',
  },
  failedNote: {
    ...type.labelSm,
    color: colors.outline,
    fontWeight: '400',
    fontStyle: 'italic',
  },

  // Source plugin chip
  sourcePill: {
    backgroundColor: colors.surfaceContainerHigh,
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: radius.sm,
  },
  sourcePillText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
    fontWeight: '400',
  },

  // Error banner
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

  // Retry button
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
