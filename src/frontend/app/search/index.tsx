/**
 * Search screen — Search & Request.
 *
 * Allows users to search for media across all plugins with a requests.* capability
 * and submit requests for items not yet in the library.
 * Spec: specs/features/requests.md
 */

import React, { useCallback, useRef, useState } from 'react';
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

import { searchMedia, submitRequest } from '../../api/requests';
import { MediaItem } from '../../types/requests';

// ---------------------------------------------------------------------------
// Type badge
// ---------------------------------------------------------------------------

const TYPE_LABEL: Record<MediaItem['type'], string> = {
  movie: 'Movie',
  show: 'TV Show',
  audiobook: 'Audiobook',
  ebook: 'Ebook',
};

function TypeBadge({ type }: { type: MediaItem['type'] }) {
  return (
    <View style={styles.typeBadge}>
      <Text style={styles.typeBadgeText}>{TYPE_LABEL[type]}</Text>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Poster
// ---------------------------------------------------------------------------

function Poster({ url }: { url?: string }) {
  if (url) {
    return <Image source={{ uri: url }} style={styles.poster} resizeMode="cover" />;
  }
  return <View style={[styles.poster, styles.posterPlaceholder]} />;
}

// ---------------------------------------------------------------------------
// Request button
// ---------------------------------------------------------------------------

interface RequestButtonProps {
  item: MediaItem;
  onDone: (itemId: string, result: 'ok' | 'error', message: string) => void;
}

function RequestButton({ item, onDone }: RequestButtonProps) {
  const [inFlight, setInFlight] = useState(false);

  const handlePress = useCallback(async () => {
    setInFlight(true);
    try {
      await submitRequest(item);
      onDone(item.id, 'ok', 'Requested!');
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Request failed';
      onDone(item.id, 'error', msg);
    } finally {
      setInFlight(false);
    }
  }, [item, onDone]);

  return (
    <TouchableOpacity
      style={[styles.requestButton, inFlight && styles.requestButtonDisabled]}
      onPress={() => { void handlePress(); }}
      disabled={inFlight}
      accessibilityLabel={`Request ${item.title}`}
    >
      {inFlight ? (
        <ActivityIndicator size="small" color="#fff" />
      ) : (
        <Text style={styles.requestButtonText}>Request</Text>
      )}
    </TouchableOpacity>
  );
}

// ---------------------------------------------------------------------------
// Media item row
// ---------------------------------------------------------------------------

interface MediaRowProps {
  item: MediaItem;
  feedback: { result: 'ok' | 'error'; message: string } | undefined;
  onRequestDone: (itemId: string, result: 'ok' | 'error', message: string) => void;
}

function MediaRow({ item, feedback, onRequestDone }: MediaRowProps) {
  return (
    <View style={styles.mediaRow}>
      <Poster url={item.poster_url} />
      <View style={styles.mediaInfo}>
        <Text style={styles.mediaTitle} numberOfLines={2}>{item.title}</Text>
        <View style={styles.mediaSubRow}>
          {item.year !== undefined ? (
            <Text style={styles.mediaYear}>{item.year}</Text>
          ) : null}
          <TypeBadge type={item.type} />
        </View>
        {/* TODO: show Available/Requested state by checking library.exists + pending requests (spec §7-8). */}
        {/* TODO: show detail/confirmation modal before submitting (spec §user-flow-5). */}
        {feedback ? (
          <Text
            style={[
              styles.feedbackText,
              feedback.result === 'ok' ? styles.feedbackOk : styles.feedbackError,
            ]}
          >
            {feedback.message}
          </Text>
        ) : (
          <RequestButton item={item} onDone={onRequestDone} />
        )}
      </View>
    </View>
  );
}

// ---------------------------------------------------------------------------
// Main screen
// ---------------------------------------------------------------------------

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
    (itemId: string, result: 'ok' | 'error', message: string) => {
      setFeedback((prev) => ({ ...prev, [itemId]: { result, message } }));
    },
    [],
  );

  return (
    <View style={styles.container}>
      {/* Search bar */}
      <View style={styles.searchBar}>
        <TextInput
          style={styles.searchInput}
          value={query}
          onChangeText={setQuery}
          placeholder="Search movies, shows, audiobooks…"
          placeholderTextColor="#9ca3af"
          returnKeyType="search"
          onSubmitEditing={() => { void handleSearch(query); }}
          autoCapitalize="none"
          autoCorrect={false}
        />
        <TouchableOpacity
          style={[styles.searchButton, !query.trim() && styles.searchButtonDisabled]}
          onPress={() => { void handleSearch(query); }}
          disabled={!query.trim() || loading}
        >
          <Text style={styles.searchButtonText}>Search</Text>
        </TouchableOpacity>
      </View>

      {/* Body */}
      {loading ? (
        <View style={styles.centered}>
          <ActivityIndicator size="large" />
        </View>
      ) : error ? (
        <View style={styles.centered}>
          <Text style={styles.errorText}>{error}</Text>
          <TouchableOpacity
            style={styles.retryButton}
            onPress={() => { void handleSearch(query); }}
          >
            <Text style={styles.retryText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <>
          {/* TODO: group results by media type: Movies, Shows, Audiobooks, Ebooks (spec §user-flow-4). */}
          <FlatList
            data={results}
            keyExtractor={(item) => item.id}
            renderItem={({ item }) => (
              <MediaRow
                item={item}
                feedback={feedback[item.id]}
                onRequestDone={handleRequestDone}
              />
            )}
            ItemSeparatorComponent={() => <View style={styles.separator} />}
            ListEmptyComponent={
              <View style={styles.centered}>
                <Text style={styles.emptyText}>
                  {hasSearched
                    ? 'No results found.'
                    : 'Search for movies and TV shows to request.'}
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

const styles = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: '#f9fafb',
  },
  searchBar: {
    flexDirection: 'row',
    padding: 12,
    backgroundColor: '#fff',
    borderBottomWidth: 1,
    borderBottomColor: '#e5e7eb',
    gap: 8,
  },
  searchInput: {
    flex: 1,
    height: 40,
    borderWidth: 1,
    borderColor: '#d1d5db',
    borderRadius: 8,
    paddingHorizontal: 12,
    fontSize: 15,
    color: '#111827',
    backgroundColor: '#f9fafb',
  },
  searchButton: {
    paddingHorizontal: 16,
    height: 40,
    backgroundColor: '#3b82f6',
    borderRadius: 8,
    alignItems: 'center',
    justifyContent: 'center',
  },
  searchButtonDisabled: {
    backgroundColor: '#93c5fd',
  },
  searchButtonText: {
    color: '#fff',
    fontWeight: '600',
    fontSize: 14,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 24,
  },
  mediaRow: {
    flexDirection: 'row',
    backgroundColor: '#fff',
    padding: 12,
    gap: 12,
  },
  poster: {
    width: 56,
    height: 84,
    borderRadius: 4,
  },
  posterPlaceholder: {
    backgroundColor: '#d1d5db',
  },
  mediaInfo: {
    flex: 1,
    justifyContent: 'space-between',
  },
  mediaTitle: {
    fontSize: 15,
    fontWeight: '600',
    color: '#111827',
    marginBottom: 4,
  },
  mediaSubRow: {
    flexDirection: 'row',
    alignItems: 'center',
    gap: 8,
    marginBottom: 8,
  },
  mediaYear: {
    fontSize: 13,
    color: '#6b7280',
  },
  typeBadge: {
    backgroundColor: '#e5e7eb',
    paddingHorizontal: 8,
    paddingVertical: 2,
    borderRadius: 4,
  },
  typeBadgeText: {
    fontSize: 12,
    color: '#374151',
    fontWeight: '500',
  },
  requestButton: {
    alignSelf: 'flex-start',
    paddingVertical: 6,
    paddingHorizontal: 14,
    backgroundColor: '#3b82f6',
    borderRadius: 6,
    minWidth: 80,
    alignItems: 'center',
  },
  requestButtonDisabled: {
    backgroundColor: '#93c5fd',
  },
  requestButtonText: {
    color: '#fff',
    fontSize: 13,
    fontWeight: '600',
  },
  feedbackText: {
    fontSize: 13,
    fontWeight: '500',
  },
  feedbackOk: {
    color: '#16a34a',
  },
  feedbackError: {
    color: '#dc2626',
  },
  separator: {
    height: 1,
    backgroundColor: '#f3f4f6',
  },
  emptyText: {
    fontSize: 15,
    color: '#6b7280',
    textAlign: 'center',
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
});
