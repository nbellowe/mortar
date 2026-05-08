/**
 * Search & Request screen.
 * Searches across request-capable plugins and lets users submit requests.
 * Spec: specs/features/requests.md
 */

import React, { useCallback, useMemo, useRef, useState } from 'react';
import {
  ActivityIndicator,
  FlatList,
  Image,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from 'react-native';
import { Ionicons } from '@expo/vector-icons';

import { searchMedia, submitRequest } from '../../api/requests';
import { MediaItem } from '../../types/requests';
import { colors, radius, spacing, type } from '@/theme/tokens';

const TYPE_LABEL: Record<MediaItem['type'], string> = {
  movie: 'Movies',
  show: 'Shows',
  audiobook: 'Audiobooks',
  ebook: 'Ebooks',
};

function PosterPlaceholder() {
  return (
    <View style={s.posterPlaceholder}>
      <Ionicons name="film-outline" size={24} color={colors.outline} />
    </View>
  );
}

function Poster({ url }: { url?: string }) {
  if (url) {
    return <Image source={{ uri: url }} style={s.poster} resizeMode="cover" />;
  }
  return <PosterPlaceholder />;
}

interface RequestButtonProps {
  item: MediaItem;
  onDone: (id: string, result: 'ok' | 'error', msg: string) => void;
}

function RequestButton({ item, onDone }: RequestButtonProps) {
  const [inFlight, setInFlight] = useState(false);

  const handlePress = useCallback(async () => {
    setInFlight(true);
    try {
      await submitRequest(item);
      onDone(item.id, 'ok', 'Requested');
    } catch (err) {
      onDone(item.id, 'error', err instanceof Error ? err.message : 'Failed');
    } finally {
      setInFlight(false);
    }
  }, [item, onDone]);

  return (
    <TouchableOpacity
      style={[s.requestBtn, inFlight && s.requestBtnDisabled]}
      onPress={() => { void handlePress(); }}
      disabled={inFlight}
      accessibilityLabel={`Request ${item.title}`}
    >
      {inFlight ? (
        <ActivityIndicator size="small" color={colors.onPrimaryContainer} />
      ) : (
        <>
          <Ionicons name="add-circle-outline" size={16} color={colors.onPrimaryContainer} />
          <Text style={s.requestBtnText}>Request</Text>
        </>
      )}
    </TouchableOpacity>
  );
}

interface MediaCardProps {
  item: MediaItem;
  feedback: { result: 'ok' | 'error'; message: string } | undefined;
  onRequestDone: (id: string, result: 'ok' | 'error', msg: string) => void;
}

function MediaCard({ item, feedback, onRequestDone }: MediaCardProps) {
  return (
    <View style={s.card}>
      <Poster url={item.poster_url} />
      <View style={s.cardMeta}>
        <Text style={s.cardYear}>{item.year ?? '—'}</Text>
      </View>
      <View style={s.cardBody}>
        <Text style={s.cardTitle} numberOfLines={2}>{item.title}</Text>
        {feedback ? (
          <View style={[s.statusChip, feedback.result === 'ok' ? s.chipAvailable : s.chipError]}>
            <Text style={[s.statusChipText, feedback.result === 'ok' ? s.chipTextAvailable : s.chipTextError]}>
              {feedback.result === 'ok' ? 'Requested' : feedback.message}
            </Text>
          </View>
        ) : (
          <RequestButton item={item} onDone={onRequestDone} />
        )}
      </View>
    </View>
  );
}

type FeedbackMap = Record<string, { result: 'ok' | 'error'; message: string }>;

export default function SearchScreen() {
  const [query, setQuery] = useState('');
  const [results, setResults] = useState<MediaItem[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);
  const [feedback, setFeedback] = useState<FeedbackMap>({});
  const abortRef = useRef<AbortController | null>(null);

  const handleSearch = useCallback(async (q: string) => {
    const trimmed = q.trim();
    if (!trimmed) return;

    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;

    setLoading(true);
    setError(null);
    setHasSearched(true);
    setFeedback({});

    try {
      const data = await searchMedia(trimmed, controller.signal);
      setResults(data);
    } catch (err) {
      if ((err as Error).name === 'AbortError') return;
      setError(err instanceof Error ? err.message : 'Search failed');
      setResults([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const handleRequestDone = useCallback(
    (id: string, result: 'ok' | 'error', message: string) => {
      setFeedback((prev) => ({ ...prev, [id]: { result, message } }));
    },
    [],
  );

  const grouped = useMemo(() => {
    const groups: { type: MediaItem['type']; label: string; items: MediaItem[] }[] = [];
    const seen = new Map<MediaItem['type'], MediaItem[]>();
    for (const item of results) {
      if (!seen.has(item.type)) seen.set(item.type, []);
      seen.get(item.type)!.push(item);
    }
    for (const [t, items] of seen) {
      groups.push({ type: t, label: TYPE_LABEL[t], items });
    }
    return groups;
  }, [results]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <View style={s.searchBar}>
          <Ionicons name="search-outline" size={18} color={colors.onSurfaceVariant} style={s.searchIcon} />
          <TextInput
            style={s.searchInput}
            value={query}
            onChangeText={setQuery}
            placeholder="Search movies, shows, audiobooks…"
            placeholderTextColor={colors.outline}
            returnKeyType="search"
            onSubmitEditing={() => { void handleSearch(query); }}
            autoCapitalize="none"
            autoCorrect={false}
          />
          {query.length > 0 && (
            <TouchableOpacity onPress={() => setQuery('')} style={s.clearBtn}>
              <Ionicons name="close-circle" size={18} color={colors.outline} />
            </TouchableOpacity>
          )}
        </View>
        <TouchableOpacity
          style={[s.searchSubmit, !query.trim() && s.searchSubmitDisabled]}
          onPress={() => { void handleSearch(query); }}
          disabled={!query.trim() || loading}
        >
          <Text style={s.searchSubmitText}>Search</Text>
        </TouchableOpacity>
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
          <TouchableOpacity style={s.retryBtn} onPress={() => { void handleSearch(query); }}>
            <Text style={s.retryBtnText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : !hasSearched ? (
        <View style={s.centered}>
          <Ionicons name="search" size={48} color={colors.outlineVariant} />
          <Text style={s.emptyTitle}>Search your media stack</Text>
          <Text style={s.emptyBody}>Find movies, shows, and audiobooks to request.</Text>
        </View>
      ) : grouped.length === 0 ? (
        <View style={s.centered}>
          <Text style={s.emptyTitle}>No results found</Text>
          <Text style={s.emptyBody}>Try a different search term.</Text>
        </View>
      ) : (
        <FlatList
          data={grouped}
          keyExtractor={(item) => item.type}
          renderItem={({ item: group }) => (
            <View style={s.group}>
              <View style={s.groupHeader}>
                <Text style={s.groupTitle}>{group.label}</Text>
                <View style={s.groupCount}>
                  <Text style={s.groupCountText}>{group.items.length}</Text>
                </View>
              </View>
              <FlatList
                data={group.items}
                horizontal
                showsHorizontalScrollIndicator={false}
                keyExtractor={(item) => item.id}
                contentContainerStyle={s.cardRow}
                renderItem={({ item }) => (
                  <MediaCard
                    item={item}
                    feedback={feedback[item.id]}
                    onRequestDone={handleRequestDone}
                  />
                )}
              />
            </View>
          )}
          contentContainerStyle={s.listContent}
        />
      )}
    </View>
  );
}

const POSTER_W = 120;
const POSTER_H = 180;

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
    paddingVertical: spacing.sm,
    backgroundColor: colors.surface,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
  },
  searchBar: {
    flex: 1,
    flexDirection: 'row',
    alignItems: 'center',
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.full,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
    paddingHorizontal: 12,
    height: 44,
  },
  searchIcon: {
    marginRight: 8,
  },
  searchInput: {
    flex: 1,
    fontSize: 16,
    color: colors.onSurface,
  },
  clearBtn: {
    padding: 4,
  },
  searchSubmit: {
    paddingVertical: 10,
    paddingHorizontal: 16,
    backgroundColor: colors.primaryContainer,
    borderRadius: radius.full,
  },
  searchSubmitDisabled: {
    opacity: 0.5,
  },
  searchSubmitText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: spacing.gutter,
    gap: spacing.sm,
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
  listContent: {
    paddingVertical: spacing.gutter,
    gap: 32,
  },
  group: {
    gap: 16,
  },
  groupHeader: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: spacing.gutter,
  },
  groupTitle: {
    ...type.headlineLg,
    color: colors.onSurface,
  },
  groupCount: {
    backgroundColor: colors.surfaceContainerHigh,
    paddingHorizontal: 10,
    paddingVertical: 4,
    borderRadius: radius.full,
  },
  groupCountText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  cardRow: {
    paddingHorizontal: spacing.gutter,
    gap: spacing.md,
  },
  card: {
    width: POSTER_W,
    gap: 8,
  },
  poster: {
    width: POSTER_W,
    height: POSTER_H,
    borderRadius: radius.lg,
  },
  posterPlaceholder: {
    width: POSTER_W,
    height: POSTER_H,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.lg,
    alignItems: 'center',
    justifyContent: 'center',
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  cardMeta: {
    position: 'absolute',
    top: 8,
    right: 8,
    backgroundColor: `${colors.surface}dd`,
    paddingHorizontal: 6,
    paddingVertical: 2,
    borderRadius: radius.sm,
  },
  cardYear: {
    ...type.labelSm,
    color: colors.onSurface,
  },
  cardBody: {
    gap: 8,
  },
  cardTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  requestBtn: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'center',
    gap: 6,
    paddingVertical: 8,
    backgroundColor: colors.primaryContainer,
    borderRadius: radius.md,
  },
  requestBtnDisabled: {
    opacity: 0.6,
  },
  requestBtnText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
  statusChip: {
    paddingVertical: 8,
    paddingHorizontal: 10,
    borderRadius: radius.md,
    alignItems: 'center',
  },
  statusChipText: {
    ...type.labelSm,
  },
  chipAvailable: {
    backgroundColor: `${colors.tertiary}22`,
    borderWidth: 1,
    borderColor: `${colors.tertiary}44`,
  },
  chipTextAvailable: {
    color: colors.tertiary,
  },
  chipError: {
    backgroundColor: `${colors.error}22`,
    borderWidth: 1,
    borderColor: `${colors.error}44`,
  },
  chipTextError: {
    color: colors.error,
  },
});
