import React, { useCallback, useMemo, useRef, useState } from "react";
import {
  ActivityIndicator,
  FlatList,
  Image,
  Linking,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";

import { playLibraryItem } from "../../api/library";
import { searchMedia, submitRequest } from "../../api/requests";
import { useAuth } from "../../components/auth-context";
import type { LibraryMatch, MediaItem, Request } from "../../types/plugin";
import { colors, radius, spacing, type } from "@/theme/tokens";

const TYPE_LABEL: Record<MediaItem["type"], string> = {
  movie: "Movies",
  show: "Shows",
  audiobook: "Audiobooks",
  ebook: "Ebooks",
};

type FeedbackMap = Record<string, string>;

function searchKey(item: MediaItem): string {
  if (item.tmdb_id) return `tmdb:${item.tmdb_id}`;
  if (item.imdb_id) return `imdb:${item.imdb_id}`;
  if (item.tvdb_id) return `tvdb:${item.tvdb_id}`;
  if (item.isbn) return `isbn:${item.isbn}`;
  if (item.asin) return `asin:${item.asin}`;
  return `id:${item.id}`;
}

function Poster({ url }: { url?: string }) {
  const [failed, setFailed] = useState(false);
  if (url && !failed) {
    return (
      <Image
        source={{ uri: url }}
        style={s.poster}
        resizeMode="cover"
        onError={() => setFailed(true)}
      />
    );
  }
  return (
    <View style={s.posterPlaceholder}>
      <Ionicons name="film-outline" size={24} color={colors.outline} />
    </View>
  );
}

function StatusChip({ color, label }: { color: string; label: string }) {
  return (
    <View style={[s.statusChip, { borderColor: color }]}>
      <Text style={[s.statusChipText, { color }]}>{label}</Text>
    </View>
  );
}

function FailedPluginsBanner({ plugins }: { plugins: string[] }) {
  if (plugins.length === 0) return null;
  return (
    <View style={s.warningBanner}>
      <Ionicons
        name="warning-outline"
        size={16}
        color={colors.statusDegraded}
      />
      <Text style={s.warningBannerText}>
        Search partial results: {plugins.join(", ")} unavailable.
      </Text>
    </View>
  );
}

function actionStateForItem(
  item: MediaItem,
  matches: Map<string, LibraryMatch>,
  requests: Map<string, Request>,
) {
  const key = searchKey(item);
  const match = matches.get(key);
  const request = requests.get(key);

  if (match) {
    return { kind: "available" as const, label: "Open in Jellyfin", match };
  }
  if (request) {
    if (
      request.status === "pending" ||
      request.status === "approved" ||
      request.status === "available"
    ) {
      return {
        kind: "requested" as const,
        label: request.status.charAt(0).toUpperCase() + request.status.slice(1),
        request,
      };
    }
    if (request.status === "failed" || request.status === "declined") {
      return { kind: "request" as const, label: "Retry" };
    }
  }
  return { kind: "request" as const, label: "Request" };
}

function ActionButton({
  item,
  actionState,
  feedback,
  requesting,
  onRequest,
  onPlay,
}: {
  item: MediaItem;
  actionState: ReturnType<typeof actionStateForItem>;
  feedback?: string;
  requesting: boolean;
  onRequest: (item: MediaItem) => void;
  onPlay: (itemId: string) => void;
}) {
  if (feedback) {
    return <StatusChip color={colors.statusHealthy} label={feedback} />;
  }

  if (actionState.kind === "available") {
    return (
      <TouchableOpacity
        style={s.availableBtn}
        onPress={() => onPlay(actionState.match.item.id)}
      >
        <Ionicons
          name="play-circle-outline"
          size={16}
          color={colors.onPrimaryContainer}
        />
        <Text style={s.availableBtnText}>{actionState.label}</Text>
      </TouchableOpacity>
    );
  }

  if (actionState.kind === "requested") {
    return (
      <StatusChip color={colors.statusDegraded} label={actionState.label} />
    );
  }

  if (requesting) {
    return <ActivityIndicator size="small" color={colors.primary} />;
  }

  return (
    <TouchableOpacity style={s.requestBtn} onPress={() => onRequest(item)}>
      <Ionicons
        name="add-circle-outline"
        size={16}
        color={colors.onPrimaryContainer}
      />
      <Text style={s.requestBtnText}>{actionState.label}</Text>
    </TouchableOpacity>
  );
}

function MediaCard({
  item,
  matches,
  requests,
  feedback,
  requesting,
  onRequest,
  onPlay,
}: {
  item: MediaItem;
  matches: Map<string, LibraryMatch>;
  requests: Map<string, Request>;
  feedback?: string;
  requesting: boolean;
  onRequest: (item: MediaItem) => void;
  onPlay: (itemId: string) => void;
}) {
  const actionState = actionStateForItem(item, matches, requests);

  return (
    <View style={s.card}>
      <Poster url={item.poster_url} />
      <View style={s.cardMeta}>
        <Text style={s.cardYear}>{item.year ?? "—"}</Text>
      </View>
      <View style={s.cardBody}>
        <Text style={s.cardTitle} numberOfLines={2}>
          {item.title}
        </Text>
        <ActionButton
          item={item}
          actionState={actionState}
          feedback={feedback}
          requesting={requesting}
          onRequest={onRequest}
          onPlay={onPlay}
        />
      </View>
    </View>
  );
}

export default function SearchScreen() {
  const { user } = useAuth();
  const [query, setQuery] = useState("");
  const [results, setResults] = useState<MediaItem[]>([]);
  const [failedPlugins, setFailedPlugins] = useState<string[]>([]);
  const [existingRequests, setExistingRequests] = useState<Request[]>([]);
  const [availableMatches, setAvailableMatches] = useState<LibraryMatch[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [hasSearched, setHasSearched] = useState(false);
  const [feedback, setFeedback] = useState<FeedbackMap>({});
  const [requestingId, setRequestingId] = useState<string | null>(null);
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
      setResults(data.items);
      setFailedPlugins(data.failed_plugins);
      setExistingRequests(data.existing_requests);
      setAvailableMatches(data.available_matches);
    } catch (err) {
      if ((err as Error).name === "AbortError") return;
      setError(err instanceof Error ? err.message : "Search failed");
      setResults([]);
      setFailedPlugins([]);
      setExistingRequests([]);
      setAvailableMatches([]);
    } finally {
      setLoading(false);
    }
  }, []);

  const matchesMap = useMemo(() => {
    const map = new Map<string, LibraryMatch>();
    for (const match of availableMatches) {
      map.set(searchKey(match.item), match);
    }
    return map;
  }, [availableMatches]);

  const requestMap = useMemo(() => {
    const map = new Map<string, Request>();
    for (const request of existingRequests) {
      map.set(searchKey(request.item), request);
    }
    return map;
  }, [existingRequests]);

  const grouped = useMemo(() => {
    const groups: {
      type: MediaItem["type"];
      label: string;
      items: MediaItem[];
    }[] = [];
    const seen = new Map<MediaItem["type"], MediaItem[]>();
    for (const item of results) {
      if (!seen.has(item.type)) seen.set(item.type, []);
      seen.get(item.type)!.push(item);
    }
    for (const [kind, items] of seen) {
      groups.push({ type: kind, label: TYPE_LABEL[kind], items });
    }
    return groups;
  }, [results]);

  const handleRequest = useCallback(async (item: MediaItem) => {
    if (requestingId != null) return;
    setRequestingId(item.id);
    try {
      const created = await submitRequest(item);
      setFeedback((prev) => ({ ...prev, [item.id]: "Requested" }));
      setExistingRequests((prev) => [created, ...prev]);
    } catch (err) {
      setFeedback((prev) => ({
        ...prev,
        [item.id]: err instanceof Error ? err.message : "Failed",
      }));
    } finally {
      setRequestingId(null);
    }
  }, [requestingId, user]);

  const handlePlay = useCallback(async (itemId: string) => {
    const result = await playLibraryItem(itemId);
    await Linking.openURL(result.url);
  }, []);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <View style={s.searchBar}>
          <Ionicons
            name="search-outline"
            size={18}
            color={colors.onSurfaceVariant}
            style={s.searchIcon}
          />
          <TextInput
            style={s.searchInput}
            value={query}
            onChangeText={setQuery}
            placeholder="Search movies, shows, audiobooks…"
            placeholderTextColor={colors.outline}
            returnKeyType="search"
            onSubmitEditing={() => {
              void handleSearch(query);
            }}
            autoCapitalize="none"
            autoCorrect={false}
          />
          {query.length > 0 ? (
            <TouchableOpacity onPress={() => setQuery("")} style={s.clearBtn}>
              <Ionicons name="close-circle" size={18} color={colors.outline} />
            </TouchableOpacity>
          ) : null}
        </View>
        <TouchableOpacity
          style={[s.searchSubmit, !query.trim() && s.searchSubmitDisabled]}
          onPress={() => {
            void handleSearch(query);
          }}
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
          <TouchableOpacity
            style={s.retryBtn}
            onPress={() => {
              void handleSearch(query);
            }}
          >
            <Text style={s.retryBtnText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : !hasSearched ? (
        <View style={s.centered}>
          <Ionicons name="search" size={48} color={colors.outlineVariant} />
          <Text style={s.emptyTitle}>Search your media stack</Text>
          <Text style={s.emptyBody}>
            Find movies, shows, audiobooks, and books in one place.
          </Text>
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
          ListHeaderComponent={<FailedPluginsBanner plugins={failedPlugins} />}
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
                    matches={matchesMap}
                    requests={requestMap}
                    feedback={feedback[item.id]}
                    requesting={requestingId === item.id}
                    onRequest={(i) => { void handleRequest(i); }}
                    onPlay={(itemId) => {
                      void handlePlay(itemId);
                    }}
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

const s = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
  },
  topBar: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.base,
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
    backgroundColor: colors.surface,
  },
  searchBar: {
    flex: 1,
    flexDirection: "row",
    alignItems: "center",
    borderRadius: radius.full,
    backgroundColor: colors.surfaceContainer,
    paddingHorizontal: spacing.sm,
  },
  searchIcon: {
    marginRight: spacing.xs,
  },
  searchInput: {
    flex: 1,
    color: colors.onSurface,
    paddingVertical: 10,
  },
  clearBtn: {
    padding: 4,
  },
  searchSubmit: {
    paddingHorizontal: spacing.md,
    paddingVertical: 10,
    borderRadius: radius.full,
    backgroundColor: colors.primaryContainer,
  },
  searchSubmitDisabled: {
    opacity: 0.55,
  },
  searchSubmitText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
  centered: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: spacing.gutter,
    gap: spacing.sm,
  },
  warningBanner: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.base,
    marginHorizontal: spacing.gutter,
    marginTop: spacing.base,
    marginBottom: spacing.base,
    padding: spacing.sm,
    borderRadius: radius.lg,
    backgroundColor: `${colors.statusDegraded}22`,
    borderWidth: 1,
    borderColor: `${colors.statusDegraded}44`,
  },
  warningBannerText: {
    ...type.labelSm,
    color: colors.onSurface,
    flex: 1,
  },
  errorBanner: {
    flexDirection: "row",
    alignItems: "center",
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
    textAlign: "center",
  },
  emptyBody: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
    textAlign: "center",
  },
  listContent: {
    paddingBottom: spacing.xl,
  },
  group: {
    gap: spacing.base,
    marginTop: spacing.base,
  },
  groupHeader: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    paddingHorizontal: spacing.gutter,
  },
  groupTitle: {
    ...type.headlineMd,
    color: colors.onSurface,
  },
  groupCount: {
    minWidth: 28,
    height: 28,
    borderRadius: radius.full,
    backgroundColor: colors.surfaceContainerHigh,
    alignItems: "center",
    justifyContent: "center",
  },
  groupCountText: {
    ...type.labelMd,
    color: colors.onSurfaceVariant,
  },
  cardRow: {
    gap: spacing.base,
    paddingHorizontal: spacing.gutter,
  },
  card: {
    width: 176,
    gap: spacing.xs,
  },
  poster: {
    width: 176,
    height: 250,
    borderRadius: radius.xl,
  },
  posterPlaceholder: {
    width: 176,
    height: 250,
    borderRadius: radius.xl,
    backgroundColor: colors.surfaceContainer,
    alignItems: "center",
    justifyContent: "center",
  },
  cardMeta: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  cardYear: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  cardBody: {
    gap: spacing.xs,
  },
  cardTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  statusChip: {
    alignSelf: "flex-start",
    paddingHorizontal: 8,
    paddingVertical: 4,
    borderRadius: radius.full,
    borderWidth: 1,
  },
  statusChipText: {
    ...type.labelSm,
  },
  requestBtn: {
    flexDirection: "row",
    alignItems: "center",
    alignSelf: "flex-start",
    gap: 4,
    backgroundColor: colors.primaryContainer,
    paddingHorizontal: spacing.sm,
    paddingVertical: 8,
    borderRadius: radius.full,
  },
  requestBtnText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
  availableBtn: {
    flexDirection: "row",
    alignItems: "center",
    alignSelf: "flex-start",
    gap: 4,
    backgroundColor: colors.primaryContainer,
    paddingHorizontal: spacing.sm,
    paddingVertical: 8,
    borderRadius: radius.full,
  },
  availableBtnText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
});
