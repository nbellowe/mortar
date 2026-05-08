/**
 * Health screen — Service Health Dashboard.
 *
 * Shows last-known health state for every configured plugin.
 * Auto-refreshes every 60 seconds (ADR 0003 shared cadence).
 * Spec: specs/features/health.md
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

import { fetchHealth } from '../../api/health';
import { PluginHealth } from '../../types/health';
import { colors, radius, spacing, type } from '@/theme/tokens';

const STATUS_COLOR: Record<PluginHealth['status'], string> = {
  healthy: colors.statusHealthy,
  degraded: colors.statusDegraded,
  unreachable: colors.statusUnreachable,
  unknown: colors.statusUnknown,
};

const STATUS_LABEL: Record<PluginHealth['status'], string> = {
  healthy: 'Healthy',
  degraded: 'Degraded',
  unreachable: 'Unreachable',
  unknown: 'Unknown',
};

const STATUS_ICON: Record<PluginHealth['status'], React.ComponentProps<typeof Ionicons>['name']> = {
  healthy: 'checkmark-circle',
  degraded: 'warning',
  unreachable: 'close-circle',
  unknown: 'help-circle',
};

function formatCheckedAt(iso: string): string {
  try {
    const diffSec = Math.floor((Date.now() - new Date(iso).getTime()) / 1000);
    if (diffSec < 60) return `${diffSec}s ago`;
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;
    return `${Math.floor(diffMin / 60)}h ago`;
  } catch {
    return iso;
  }
}

function SummaryBanner({ plugins }: { plugins: PluginHealth[] }) {
  if (plugins.length === 0) return null;
  const unhealthy = plugins.filter((p) => p.status !== 'healthy').length;
  const allGood = unhealthy === 0;
  return (
    <View style={[s.banner, allGood ? s.bannerGood : s.bannerBad]}>
      <Ionicons
        name={allGood ? 'checkmark-circle' : 'warning'}
        size={18}
        color={allGood ? colors.statusHealthy : colors.statusDegraded}
      />
      <Text style={s.bannerText}>
        {allGood
          ? `All ${plugins.length} services healthy`
          : `${unhealthy} of ${plugins.length} service${unhealthy !== 1 ? 's' : ''} need attention`}
      </Text>
    </View>
  );
}

function PluginCard({ plugin }: { plugin: PluginHealth }) {
  const statusColor = STATUS_COLOR[plugin.status];
  return (
    <View style={s.card}>
      <View style={s.cardLeft}>
        <Ionicons name={STATUS_ICON[plugin.status]} size={22} color={statusColor} />
        <View style={s.cardInfo}>
          <Text style={s.cardName}>{plugin.display_name}</Text>
          <Text style={s.cardMeta}>{plugin.plugin_type}</Text>
          {plugin.detail ? (
            <Text style={s.cardDetail} numberOfLines={2}>{plugin.detail}</Text>
          ) : null}
        </View>
      </View>
      <View style={s.cardRight}>
        <View style={[s.statusPill, { borderColor: statusColor }]}>
          <Text style={[s.statusPillText, { color: statusColor }]}>
            {STATUS_LABEL[plugin.status]}
          </Text>
        </View>
        <Text style={s.cardLatency}>{plugin.latency_ms} ms</Text>
        <Text style={s.cardChecked}>{formatCheckedAt(plugin.checked_at)}</Text>
      </View>
    </View>
  );
}

const REFRESH_INTERVAL_MS = 60_000;

export default function HealthScreen() {
  const [plugins, setPlugins] = useState<PluginHealth[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const load = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    try {
      const data = await fetchHealth(controller.signal);
      setPlugins(data);
      setError(null);
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      setError(err instanceof Error ? err.message : 'Failed to load health data');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
    return () => { abortRef.current?.abort(); };
  }, [load]);

  useEffect(() => {
    const id = setInterval(() => { void load(); }, REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [load]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Ionicons name="heart" size={22} color={colors.primaryFixedDim} />
        <Text style={s.topBarTitle}>Service Health</Text>
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
          <TouchableOpacity style={s.retryBtn} onPress={() => { setLoading(true); void load(); }}>
            <Text style={s.retryBtnText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <FlatList
          data={plugins}
          keyExtractor={(item) => item.plugin_id}
          ListHeaderComponent={<SummaryBanner plugins={plugins} />}
          renderItem={({ item }) => <PluginCard plugin={item} />}
          ItemSeparatorComponent={() => <View style={s.separator} />}
          contentContainerStyle={s.listContent}
          ListEmptyComponent={
            <View style={s.centered}>
              <Ionicons name="server-outline" size={48} color={colors.outlineVariant} />
              <Text style={s.emptyTitle}>No plugins configured</Text>
              <Text style={s.emptyBody}>Add plugins in your Mortar config to see health status.</Text>
            </View>
          }
        />
      )}
    </View>
  );
}

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
  banner: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: spacing.base,
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.sm,
    marginHorizontal: spacing.gutter,
    marginTop: spacing.md,
    borderRadius: radius.lg,
  },
  bannerGood: {
    backgroundColor: `${colors.statusHealthy}22`,
    borderWidth: 1,
    borderColor: `${colors.statusHealthy}44`,
  },
  bannerBad: {
    backgroundColor: `${colors.statusDegraded}22`,
    borderWidth: 1,
    borderColor: `${colors.statusDegraded}44`,
  },
  bannerText: {
    ...type.labelMd,
    color: colors.onSurface,
    flex: 1,
  },
  listContent: {
    paddingBottom: spacing.xl,
    gap: spacing.base,
    paddingTop: spacing.base,
  },
  card: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    backgroundColor: colors.surfaceContainer,
    marginHorizontal: spacing.gutter,
    borderRadius: radius.xl,
    padding: spacing.md,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  cardLeft: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    flex: 1,
    gap: spacing.sm,
  },
  cardInfo: {
    flex: 1,
  },
  cardName: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  cardMeta: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
    marginTop: 2,
    textTransform: 'uppercase',
    letterSpacing: 0.5,
  },
  cardDetail: {
    ...type.bodyMd,
    fontSize: 12,
    color: colors.error,
    marginTop: 4,
  },
  cardRight: {
    alignItems: 'flex-end',
    gap: 4,
    marginLeft: spacing.sm,
  },
  statusPill: {
    paddingHorizontal: 8,
    paddingVertical: 3,
    borderRadius: radius.full,
    borderWidth: 1,
  },
  statusPillText: {
    ...type.labelSm,
  },
  cardLatency: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  cardChecked: {
    fontSize: 11,
    color: colors.outline,
    lineHeight: 14,
    fontWeight: '400',
  },
  separator: {
    height: 0,
  },
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
