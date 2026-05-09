import React, {
  useCallback,
  useEffect,
  useMemo,
  useRef,
  useState,
} from "react";
import {
  ActivityIndicator,
  FlatList,
  Image,
  Linking,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { Ionicons } from "@expo/vector-icons";

import { browseLibrary, playLibraryItem } from "../../api/library";
import type { MediaItem } from "../../types/plugin";
import { colors, radius, spacing, type } from "@/theme/tokens";

type SortKey = "added" | "title" | "year";
type TypeKey = "all" | "movie" | "show";

const SORT_OPTIONS: { key: SortKey; label: string }[] = [
  { key: "added", label: "Recently added" },
  { key: "title", label: "Title" },
  { key: "year", label: "Year" },
];

const TYPE_OPTIONS: { key: TypeKey; label: string }[] = [
  { key: "all", label: "All" },
  { key: "movie", label: "Movies" },
  { key: "show", label: "Shows" },
];

function FilterChip({
  label,
  active,
  onPress,
}: {
  label: string;
  active: boolean;
  onPress: () => void;
}) {
  return (
    <TouchableOpacity
      style={[s.chip, active && s.chipActive]}
      onPress={onPress}
    >
      <Text style={[s.chipText, active && s.chipTextActive]}>{label}</Text>
    </TouchableOpacity>
  );
}

function LibraryCard({
  item,
  onPress,
}: {
  item: MediaItem;
  onPress: () => void;
}) {
  const [failed, setFailed] = useState(false);
  return (
    <TouchableOpacity style={s.card} onPress={onPress}>
      <View style={s.cardImage}>
        {item.poster_url && !failed ? (
          <Image
            source={{ uri: item.poster_url }}
            style={s.cardImageFill}
            resizeMode="cover"
            onError={() => setFailed(true)}
          />
        ) : (
          <Ionicons name="film-outline" size={30} color={colors.outline} />
        )}
      </View>
      <Text style={s.cardTitle} numberOfLines={2}>
        {item.title}
      </Text>
      <Text style={s.cardMeta}>
        {[item.year ?? "—", item.type].join(" · ")}
      </Text>
    </TouchableOpacity>
  );
}

export default function LibraryScreen() {
  const [items, setItems] = useState<MediaItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [loadingMore, setLoadingMore] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [requiresLink, setRequiresLink] = useState(false);
  const [pluginDisplayName, setPluginDisplayName] = useState<
    string | undefined
  >();
  const [genres, setGenres] = useState<string[]>([]);
  const [selectedGenre, setSelectedGenre] = useState<string | undefined>();
  const [selectedType, setSelectedType] = useState<TypeKey>("all");
  const [sortKey, setSortKey] = useState<SortKey>("added");
  const [page, setPage] = useState(1);
  const [total, setTotal] = useState(0);
  const abortRef = useRef<AbortController | null>(null);

  const canLoadMore = items.length < total;

  const load = useCallback(
    async (nextPage: number, append: boolean) => {
      abortRef.current?.abort();
      const controller = new AbortController();
      abortRef.current = controller;
      if (append) {
        setLoadingMore(true);
      } else {
        setLoading(true);
      }
      try {
        const response = await browseLibrary({
          page: nextPage,
          pageSize: 24,
          sort: sortKey,
          type: selectedType === "all" ? undefined : selectedType,
          genre: selectedGenre,
          signal: controller.signal,
        });
        setRequiresLink(response.requires_link);
        setPluginDisplayName(response.plugin_display_name);
        setGenres(response.available_genres);
        setTotal(response.total);
        setPage(response.page);
        setItems((prev) =>
          append
            ? [
                ...prev,
                ...response.items.filter(
                  (item) => !prev.some((existing) => existing.id === item.id),
                ),
              ]
            : response.items,
        );
        setError(null);
      } catch (err) {
        if ((err as Error).name === "AbortError") return;
        setError(err instanceof Error ? err.message : "Failed to load library");
      } finally {
        setLoading(false);
        setLoadingMore(false);
      }
    },
    [selectedGenre, selectedType, sortKey],
  );

  const play = useCallback(async (itemId: string) => {
    const result = await playLibraryItem(itemId);
    await Linking.openURL(result.url);
  }, []);

  useEffect(() => {
    void load(1, false);
    return () => {
      abortRef.current?.abort();
    };
  }, [load]);

  const filterSummary = useMemo(() => {
    if (selectedGenre) return `Filtered by ${selectedGenre}`;
    if (selectedType !== "all")
      return `Showing ${selectedType === "movie" ? "movies" : "shows"}`;
    return "Browse what is already in Jellyfin";
  }, [selectedGenre, selectedType]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Ionicons name="film" size={22} color={colors.primaryFixedDim} />
        <View style={s.topBarCopy}>
          <Text style={s.topBarTitle}>Library</Text>
          <Text style={s.topBarBody}>{filterSummary}</Text>
        </View>
      </View>

      {loading ? (
        <View style={s.centered}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : requiresLink ? (
        <View style={s.centered}>
          <View style={s.infoCard}>
            <Ionicons
              name="person-circle-outline"
              size={36}
              color={colors.outline}
            />
            <Text style={s.infoTitle}>Linked account required</Text>
            <Text style={s.infoBody}>
              Browse & Play is link-gated for{" "}
              {pluginDisplayName ?? "your library plugin"} in v1.
            </Text>
          </View>
        </View>
      ) : error ? (
        <View style={s.centered}>
          <View style={s.infoCard}>
            <Ionicons name="warning-outline" size={36} color={colors.error} />
            <Text style={s.infoTitle}>Library could not load</Text>
            <Text style={s.infoBody}>{error}</Text>
          </View>
          <TouchableOpacity
            style={s.retryBtn}
            onPress={() => {
              void load(1, false);
            }}
          >
            <Text style={s.retryBtnText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <FlatList
          data={items}
          keyExtractor={(item) => item.id}
          numColumns={2}
          columnWrapperStyle={s.gridRow}
          renderItem={({ item }) => (
            <LibraryCard
              item={item}
              onPress={() => {
                void play(item.id);
              }}
            />
          )}
          contentContainerStyle={s.listContent}
          ListHeaderComponent={
            <View style={s.filtersWrap}>
              <View style={s.filterGroup}>
                {TYPE_OPTIONS.map((option) => (
                  <FilterChip
                    key={option.key}
                    label={option.label}
                    active={selectedType === option.key}
                    onPress={() => {
                      setSelectedType(option.key);
                      setSelectedGenre(undefined);
                    }}
                  />
                ))}
              </View>
              <View style={s.filterGroup}>
                {SORT_OPTIONS.map((option) => (
                  <FilterChip
                    key={option.key}
                    label={option.label}
                    active={sortKey === option.key}
                    onPress={() => setSortKey(option.key)}
                  />
                ))}
              </View>
              {genres.length > 0 ? (
                <View style={s.filterGroup}>
                  <FilterChip
                    label="All genres"
                    active={!selectedGenre}
                    onPress={() => setSelectedGenre(undefined)}
                  />
                  {genres.slice(0, 8).map((genre) => (
                    <FilterChip
                      key={genre}
                      label={genre}
                      active={selectedGenre === genre}
                      onPress={() => setSelectedGenre(genre)}
                    />
                  ))}
                </View>
              ) : null}
            </View>
          }
          ListFooterComponent={
            canLoadMore ? (
              <TouchableOpacity
                style={[s.loadMoreBtn, loadingMore && s.loadMoreBtnDisabled]}
                onPress={() => {
                  void load(page + 1, true);
                }}
                disabled={loadingMore}
              >
                {loadingMore ? (
                  <ActivityIndicator
                    size="small"
                    color={colors.onPrimaryContainer}
                  />
                ) : (
                  <Text style={s.loadMoreText}>Load more</Text>
                )}
              </TouchableOpacity>
            ) : null
          }
          ListEmptyComponent={
            <View style={s.centered}>
              <Ionicons
                name="film-outline"
                size={48}
                color={colors.outlineVariant}
              />
              <Text style={s.infoTitle}>No library matches</Text>
              <Text style={s.infoBody}>
                Try a different type, sort, or genre filter.
              </Text>
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
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.base,
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
    backgroundColor: colors.surface,
  },
  topBarCopy: {
    flex: 1,
  },
  topBarTitle: {
    ...type.headlineLg,
    color: colors.onSurface,
  },
  topBarBody: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  centered: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: spacing.gutter,
    gap: spacing.sm,
  },
  infoCard: {
    alignItems: "center",
    justifyContent: "center",
    gap: spacing.sm,
    padding: spacing.gutter,
    borderRadius: radius.xl,
    backgroundColor: colors.surfaceContainerLow,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  infoTitle: {
    ...type.headlineMd,
    color: colors.onSurface,
    textAlign: "center",
  },
  infoBody: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
    textAlign: "center",
  },
  filtersWrap: {
    gap: spacing.base,
    paddingBottom: spacing.md,
  },
  filterGroup: {
    flexDirection: "row",
    flexWrap: "wrap",
    gap: spacing.xs,
  },
  chip: {
    borderRadius: radius.full,
    paddingHorizontal: spacing.sm,
    paddingVertical: 6,
    backgroundColor: colors.surfaceContainer,
  },
  chipActive: {
    backgroundColor: colors.primaryFixed,
  },
  chipText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  chipTextActive: {
    color: colors.onPrimaryFixed,
  },
  listContent: {
    padding: spacing.gutter,
    paddingBottom: spacing.xl,
  },
  gridRow: {
    justifyContent: "space-between",
    gap: spacing.base,
    marginBottom: spacing.base,
  },
  card: {
    flex: 1,
    maxWidth: "48%",
    gap: spacing.xs,
  },
  cardImage: {
    aspectRatio: 2 / 3,
    borderRadius: radius.xl,
    overflow: "hidden",
    backgroundColor: colors.surfaceContainer,
    alignItems: "center",
    justifyContent: "center",
  },
  cardImageFill: {
    width: "100%",
    height: "100%",
  },
  cardTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  cardMeta: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
    textTransform: "capitalize",
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
  loadMoreBtn: {
    borderRadius: radius.full,
    backgroundColor: colors.primaryContainer,
    alignItems: "center",
    justifyContent: "center",
    minHeight: 44,
    marginTop: spacing.md,
  },
  loadMoreBtnDisabled: {
    opacity: 0.65,
  },
  loadMoreText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
});
