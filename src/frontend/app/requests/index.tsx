/**
 * Requests screen — My Requests.
 *
 * Shows the current user's media requests with status badges.
 * Auto-refreshes every 30 seconds.
 * Requester ID is hardcoded to "anonymous" until auth is implemented.
 * Spec: specs/features/requests.md
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

import { fetchRequests } from '../../api/requests';
import { Request, RequestStatus } from '../../types/requests';
import { colors, radius, spacing, type } from '@/theme/tokens';

const STATUS_COLOR: Record<RequestStatus, string> = {
  pending: colors.statusDegraded,
  approved: colors.tertiary,
  available: colors.statusHealthy,
  declined: colors.statusUnreachable,
  failed: colors.statusUnknown,
};

const STATUS_LABEL: Record<RequestStatus, string> = {
  pending: 'Pending',
  approved: 'Approved',
  available: 'Available',
  declined: 'Declined',
  failed: 'Failed',
};

const STATUS_ICON: Record<RequestStatus, React.ComponentProps<typeof Ionicons>['name']> = {
  pending: 'time-outline',
  approved: 'checkmark-circle-outline',
  available: 'checkmark-circle',
  declined: 'close-circle-outline',
  failed: 'alert-circle-outline',
};

function StatusBadge({ status }: { status: RequestStatus }) {
  const color = STATUS_COLOR[status];
  return (
    <View style={[s.statusBadge, { borderColor: color }]}>
      <Ionicons name={STATUS_ICON[status]} size={13} color={color} />
      <Text style={[s.statusBadgeText, { color }]}>{STATUS_LABEL[status]}</Text>
    </View>
  );
}

function formatDate(iso: string): string {
  try {
    return new Date(iso).toLocaleDateString(undefined, {
      year: 'numeric', month: 'short', day: 'numeric',
    });
  } catch {
    return iso;
  }
}

function RequestRow({ request }: { request: Request }) {
  return (
    <View style={s.row}>
      <View style={s.rowLeft}>
        <View style={s.rowIcon}>
          <Ionicons name="film-outline" size={20} color={colors.onSurfaceVariant} />
        </View>
        <View style={s.rowInfo}>
          <Text style={s.rowTitle} numberOfLines={2}>{request.item.title}</Text>
          <Text style={s.rowDate}>Submitted {formatDate(request.submitted_at)}</Text>
        </View>
      </View>
      <StatusBadge status={request.status} />
    </View>
  );
}

const REFRESH_INTERVAL_MS = 30_000;

export default function RequestsScreen() {
  const [requests, setRequests] = useState<Request[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const load = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    try {
      const data = await fetchRequests({ signal: controller.signal });
      setRequests(data);
      setError(null);
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      setError(err instanceof Error ? err.message : 'Failed to load requests');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    void load();
    return () => { abortRef.current?.abort(); };
  }, [load]);

  // TODO: extract into shared usePollingInterval hook (ADR 0003)
  useEffect(() => {
    const id = setInterval(() => { void load(); }, REFRESH_INTERVAL_MS);
    return () => clearInterval(id);
  }, [load]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Ionicons name="receipt-outline" size={22} color={colors.primaryFixedDim} />
        <Text style={s.topBarTitle}>My Requests</Text>
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
          data={requests}
          keyExtractor={(item) => item.id}
          renderItem={({ item }) => <RequestRow request={item} />}
          ItemSeparatorComponent={() => <View style={s.separator} />}
          contentContainerStyle={s.listContent}
          ListEmptyComponent={
            <View style={s.centered}>
              <Ionicons name="receipt-outline" size={48} color={colors.outlineVariant} />
              <Text style={s.emptyTitle}>No requests yet</Text>
              <Text style={s.emptyBody}>Search for movies, shows, and audiobooks to request them.</Text>
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
  listContent: {
    paddingVertical: spacing.base,
    paddingBottom: spacing.xl,
  },
  row: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingVertical: spacing.md,
    paddingHorizontal: spacing.gutter,
    gap: spacing.sm,
  },
  rowLeft: {
    flexDirection: 'row',
    alignItems: 'center',
    flex: 1,
    gap: spacing.sm,
  },
  rowIcon: {
    width: 40,
    height: 40,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.lg,
    alignItems: 'center',
    justifyContent: 'center',
  },
  rowInfo: {
    flex: 1,
  },
  rowTitle: {
    ...type.labelMd,
    color: colors.onSurface,
    marginBottom: 2,
  },
  rowDate: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
    fontWeight: '400',
  },
  statusBadge: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 4,
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: radius.full,
    borderWidth: 1,
  },
  statusBadgeText: {
    ...type.labelSm,
  },
  separator: {
    height: StyleSheet.hairlineWidth,
    backgroundColor: colors.outlineVariant,
    marginHorizontal: spacing.gutter,
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
