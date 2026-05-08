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

import { fetchHealth } from '../../api/health';
import { PluginHealth } from '../../types/health';

// ---------------------------------------------------------------------------
// Status badge
// ---------------------------------------------------------------------------

const STATUS_COLOR: Record<PluginHealth['status'], string> = {
  healthy: '#22c55e',    // green
  degraded: '#eab308',   // yellow
  unreachable: '#ef4444', // red
  unknown: '#9ca3af',    // grey
};

function StatusDot({ status }: { status: PluginHealth['status'] }) {
  return <View style={[styles.dot, { backgroundColor: STATUS_COLOR[status] }]} />;
}

// ---------------------------------------------------------------------------
// Formatting helpers
// ---------------------------------------------------------------------------

function formatCheckedAt(iso: string): string {
  try {
    const d = new Date(iso);
    const diffMs = Date.now() - d.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    if (diffSec < 60) return `${diffSec}s ago`;
    const diffMin = Math.floor(diffSec / 60);
    if (diffMin < 60) return `${diffMin}m ago`;
    const diffHr = Math.floor(diffMin / 60);
    return `${diffHr}h ago`;
  } catch {
    return iso;
  }
}

function statusLabel(status: PluginHealth['status']): string {
  switch (status) {
    case 'healthy': return 'Healthy';
    case 'degraded': return 'Degraded';
    case 'unreachable': return 'Unreachable';
    case 'unknown': return 'Unknown';
  }
}

// ---------------------------------------------------------------------------
// Summary banner
// ---------------------------------------------------------------------------

function SummaryBanner({ plugins }: { plugins: PluginHealth[] }) {
  const unreachable = plugins.filter((p) => p.status === 'unreachable').length;
  const total = plugins.length;

  if (total === 0) return null;

  const allHealthy = unreachable === 0 && plugins.every((p) => p.status !== 'unknown');

  return (
    <View style={[styles.banner, allHealthy ? styles.bannerGood : styles.bannerBad]}>
      <Text style={styles.bannerText}>
        {allHealthy
          ? 'All services healthy'
          : `${unreachable} of ${total} service${unreachable !== 1 ? 's' : ''} unreachable`}
      </Text>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Plugin row
// ---------------------------------------------------------------------------

function PluginRow({ plugin }: { plugin: PluginHealth }) {
  return (
    <View style={styles.row}>
      <View style={styles.rowLeft}>
        <StatusDot status={plugin.status} />
        <View style={styles.rowInfo}>
          <Text style={styles.rowName}>{plugin.display_name}</Text>
          <Text style={styles.rowMeta}>
            {plugin.plugin_type} · {statusLabel(plugin.status)}
          </Text>
          {plugin.detail ? (
            <Text style={styles.rowDetail} numberOfLines={2}>
              {plugin.detail}
            </Text>
          ) : null}
        </View>
      </View>
      <View style={styles.rowRight}>
        <Text style={styles.rowLatency}>{plugin.latency_ms} ms</Text>
        <Text style={styles.rowChecked}>{formatCheckedAt(plugin.checked_at)}</Text>
      </View>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Main screen
// ---------------------------------------------------------------------------

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

  // Initial load
  useEffect(() => {
    void load();
    return () => {
      abortRef.current?.abort();
    };
  }, [load]);

  // Auto-refresh every 60 seconds
  useEffect(() => {
    const id = setInterval(() => {
      void load();
    }, REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [load]);

  if (loading) {
    return (
      <View style={styles.centered}>
        <ActivityIndicator size="large" />
      </View>
    );
  }

  if (error) {
    return (
      <View style={styles.centered}>
        <Text style={styles.errorText}>{error}</Text>
        <TouchableOpacity style={styles.retryButton} onPress={() => { setLoading(true); void load(); }}>
          <Text style={styles.retryText}>Retry</Text>
        </TouchableOpacity>
      </View>
    );
  }

  return (
    <View style={styles.container}>
      <SummaryBanner plugins={plugins} />
      <FlatList
        data={plugins}
        keyExtractor={(item) => item.plugin_id}
        renderItem={({ item }) => <PluginRow plugin={item} />}
        ItemSeparatorComponent={() => <View style={styles.separator} />}
        ListEmptyComponent={
          <View style={styles.centered}>
            <Text style={styles.emptyText}>No plugins configured.</Text>
          </View>
        }
      />
    </View>
  );
}

// ---------------------------------------------------------------------------
// Styles
// ---------------------------------------------------------------------------

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
  },
  banner: {
    paddingVertical: 10,
    paddingHorizontal: 16,
  },
  bannerGood: {
    backgroundColor: '#dcfce7',
  },
  bannerBad: {
    backgroundColor: '#fee2e2',
  },
  bannerText: {
    fontSize: 14,
    fontWeight: '600',
    color: '#111827',
  },
  row: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    justifyContent: 'space-between',
    padding: 16,
    backgroundColor: '#fff',
  },
  rowLeft: {
    flexDirection: 'row',
    alignItems: 'flex-start',
    flex: 1,
  },
  dot: {
    width: 12,
    height: 12,
    borderRadius: 6,
    marginTop: 3,
    marginRight: 12,
  },
  rowInfo: {
    flex: 1,
  },
  rowName: {
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
  },
  rowMeta: {
    fontSize: 13,
    color: '#6b7280',
    marginTop: 2,
  },
  rowDetail: {
    fontSize: 12,
    color: '#ef4444',
    marginTop: 4,
  },
  rowRight: {
    alignItems: 'flex-end',
    marginLeft: 12,
  },
  rowLatency: {
    fontSize: 14,
    fontWeight: '500',
    color: '#374151',
  },
  rowChecked: {
    fontSize: 12,
    color: '#9ca3af',
    marginTop: 2,
  },
  separator: {
    height: 1,
    backgroundColor: '#f3f4f6',
  },
  errorText: {
    fontSize: 15,
    color: '#ef4444',
    textAlign: 'center',
    marginBottom: 16,
  },
  retryButton: {
    paddingVertical: 8,
    paddingHorizontal: 20,
    backgroundColor: '#3b82f6',
    borderRadius: 6,
  },
  retryText: {
    color: '#fff',
    fontSize: 14,
    fontWeight: '600',
  },
  emptyText: {
    fontSize: 15,
    color: '#6b7280',
  },
});
