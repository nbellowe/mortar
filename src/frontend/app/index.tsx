import React, { useCallback, useEffect, useRef, useState } from "react";
import {
  ActivityIndicator,
  Image,
  Linking,
  ScrollView,
  StyleSheet,
  Text,
  TouchableOpacity,
  View,
} from "react-native";
import { Link } from "expo-router";
import { Ionicons } from "@expo/vector-icons";

import { getHome, HomeResponse } from "../api/home";
import { playLibraryItem } from "../api/library";
import { useAuth } from "../components/auth-context";
import type { ContinueWatchingItem, MediaItem } from "../types/plugin";
import { colors, radius, spacing, type } from "@/theme/tokens";

function SectionHeader({
  icon,
  title,
  href,
}: {
  icon: React.ComponentProps<typeof Ionicons>["name"];
  title: string;
  href?: "/" | `/${string}`;
}) {
  return (
    <View style={s.sectionHeader}>
      <View style={s.sectionHeaderLeft}>
        <Ionicons name={icon} size={20} color={colors.primaryFixedDim} />
        <Text style={s.sectionTitle}>{title}</Text>
      </View>
      {href ? (
        <Link href={href} style={s.sectionAction}>
          View all
        </Link>
      ) : null}
    </View>
  );
}

function InlineMessage({
  icon,
  title,
  body,
}: {
  icon: React.ComponentProps<typeof Ionicons>["name"];
  title: string;
  body: string;
}) {
  return (
    <View style={s.infoCard}>
      <Ionicons name={icon} size={24} color={colors.outline} />
      <View style={s.infoBody}>
        <Text style={s.infoTitle}>{title}</Text>
        <Text style={s.infoText}>{body}</Text>
      </View>
    </View>
  );
}

function PosterCard({
  item,
  subtitle,
  badge,
  onPress,
}: {
  item: MediaItem;
  subtitle?: string;
  badge?: string;
  onPress: () => void;
}) {
  const [imgFailed, setImgFailed] = useState(false);

  return (
    <TouchableOpacity style={s.posterCard} onPress={onPress}>
      <View style={s.posterImage}>
        {item.poster_url && !imgFailed ? (
          <Image
            source={{ uri: item.poster_url }}
            style={s.posterImageFill}
            resizeMode="cover"
            onError={() => setImgFailed(true)}
          />
        ) : (
          <Ionicons name="film-outline" size={28} color={colors.outline} />
        )}
      </View>
      {badge ? (
        <View style={s.badge}>
          <Text style={s.badgeText}>{badge}</Text>
        </View>
      ) : null}
      <Text style={s.posterTitle} numberOfLines={1}>
        {item.title}
      </Text>
      <Text style={s.posterSubtitle}>{subtitle ?? item.year ?? "—"}</Text>
    </TouchableOpacity>
  );
}

function ProgressBar({ progress }: { progress: number }) {
  return (
    <View style={s.progressTrack}>
      <View
        style={[
          s.progressFill,
          { width: `${Math.min(Math.max(progress, 0), 1) * 100}%` },
        ]}
      />
    </View>
  );
}

function ContinueWatchingCard({
  entry,
  onPress,
}: {
  entry: ContinueWatchingItem;
  onPress: () => void;
}) {
  return (
    <View style={s.continueCard}>
      <PosterCard
        item={entry.item}
        subtitle="Resume"
        badge={`${Math.round(entry.progress * 100)}%`}
        onPress={onPress}
      />
      <ProgressBar progress={entry.progress} />
    </View>
  );
}

function HealthBadge({ summary }: { summary: HomeResponse["health_summary"] }) {
  if (!summary.any_unreachable) return null;
  return (
    <View style={s.healthBadge}>
      <Ionicons name="warning" size={14} color={colors.statusUnreachable} />
      <Text style={s.healthBadgeText}>
        {summary.unreachable_count} service
        {summary.unreachable_count !== 1 ? "s" : ""} unreachable
      </Text>
    </View>
  );
}

export default function HomeScreen() {
  const { user, logout } = useAuth();
  const [data, setData] = useState<HomeResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [playingId, setPlayingId] = useState<string | null>(null);
  const abortRef = useRef<AbortController | null>(null);

  const load = useCallback(async () => {
    abortRef.current?.abort();
    const controller = new AbortController();
    abortRef.current = controller;
    try {
      const home = await getHome(controller.signal);
      setData(home);
      setError(null);
    } catch (err) {
      if ((err as Error).name === "AbortError") return;
      setError(err instanceof Error ? err.message : "Failed to load home");
    } finally {
      setLoading(false);
    }
  }, []);

  const handlePlay = useCallback(async (itemId: string) => {
    setPlayingId(itemId);
    try {
      const result = await playLibraryItem(itemId);
      await Linking.openURL(result.url);
    } finally {
      setPlayingId(null);
    }
  }, []);

  useEffect(() => {
    void load();
    return () => {
      abortRef.current?.abort();
    };
  }, [load]);

  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <View>
          <Text style={s.topBarTitle}>Mortar</Text>
          {user ? (
            <Text style={s.topBarSubtitle}>Signed in as {user.username}</Text>
          ) : null}
        </View>
        <TouchableOpacity
          style={s.signOutBtn}
          onPress={() => {
            void logout();
          }}
        >
          <Text style={s.signOutBtnText}>Sign out</Text>
        </TouchableOpacity>
      </View>

      {loading ? (
        <View style={s.centered}>
          <ActivityIndicator size="large" color={colors.primary} />
        </View>
      ) : error || !data ? (
        <View style={s.centered}>
          <InlineMessage
            icon="warning-outline"
            title="Home could not load"
            body={error ?? "Try again in a moment."}
          />
          <TouchableOpacity
            style={s.retryBtn}
            onPress={() => {
              setLoading(true);
              void load();
            }}
          >
            <Text style={s.retryBtnText}>Retry</Text>
          </TouchableOpacity>
        </View>
      ) : (
        <ScrollView contentContainerStyle={s.scroll}>
          <View style={s.hero}>
            <Text style={s.heroGreeting}>Welcome back.</Text>
            <Text style={s.heroSubtitle}>
              Your household queue, library, and stack health in one place.
            </Text>
            <HealthBadge summary={data.health_summary} />
          </View>

          <View style={s.section}>
            <SectionHeader
              icon="play-circle-outline"
              title="Continue Watching"
            />
            {!data.continue_watching_enabled ? null : data.continue_watching_requires_link ? (
              <InlineMessage
                icon="person-circle-outline"
                title="Link required"
                body="Connect your Jellyfin account to unlock personalized resume items."
              />
            ) : data.continue_watching.length === 0 ? (
              <InlineMessage
                icon="sparkles-outline"
                title="Nothing to resume"
                body="Start watching something in Jellyfin and it will appear here."
              />
            ) : (
              <ScrollView
                horizontal
                showsHorizontalScrollIndicator={false}
                contentContainerStyle={s.posterRow}
              >
                {data.continue_watching.map((entry) => (
                  <ContinueWatchingCard
                    key={entry.item.id}
                    entry={entry}
                    onPress={() => {
                      void handlePlay(entry.item.id);
                    }}
                  />
                ))}
              </ScrollView>
            )}
          </View>

          <View style={s.section}>
            <SectionHeader
              icon="film-outline"
              title="Recently Added"
              href="/library"
            />
            {data.recently_added_requires_link ? (
              <InlineMessage
                icon="person-circle-outline"
                title="Link required"
                body="Browse & Play is link-gated in v1, so this row appears after you link a Jellyfin account."
              />
            ) : data.recently_added.length === 0 ? (
              <InlineMessage
                icon="film-outline"
                title="No recent additions"
                body="When new items land in Jellyfin, they will show up here."
              />
            ) : (
              <ScrollView
                horizontal
                showsHorizontalScrollIndicator={false}
                contentContainerStyle={s.posterRow}
              >
                {data.recently_added.map((item) => (
                  <PosterCard
                    key={item.id}
                    item={item}
                    onPress={() => {
                      void handlePlay(item.id);
                    }}
                    badge={playingId === item.id ? "Opening…" : undefined}
                  />
                ))}
              </ScrollView>
            )}
          </View>
        </ScrollView>
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
    paddingHorizontal: spacing.gutter,
    paddingVertical: spacing.md,
    borderBottomWidth: StyleSheet.hairlineWidth,
    borderBottomColor: colors.outlineVariant,
    backgroundColor: colors.surface,
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
    gap: spacing.base,
  },
  topBarTitle: {
    ...type.headlineLg,
    color: colors.primary,
    letterSpacing: -0.5,
  },
  topBarSubtitle: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  signOutBtn: {
    borderRadius: radius.full,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
    backgroundColor: colors.surfaceContainer,
  },
  signOutBtnText: {
    ...type.labelMd,
    color: colors.onSurfaceVariant,
  },
  centered: {
    flex: 1,
    alignItems: "center",
    justifyContent: "center",
    padding: spacing.gutter,
    gap: spacing.sm,
  },
  scroll: {
    padding: spacing.gutter,
    gap: spacing.xl,
    paddingBottom: spacing.xl,
  },
  hero: {
    gap: spacing.xs,
  },
  heroGreeting: {
    ...type.displayLg,
    color: colors.onSurface,
  },
  heroSubtitle: {
    ...type.bodyLg,
    color: colors.onSurfaceVariant,
  },
  healthBadge: {
    flexDirection: "row",
    alignItems: "center",
    gap: spacing.xs,
    alignSelf: "flex-start",
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
    gap: spacing.base,
  },
  sectionHeader: {
    flexDirection: "row",
    alignItems: "center",
    justifyContent: "space-between",
  },
  sectionHeaderLeft: {
    flexDirection: "row",
    alignItems: "center",
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
  infoCard: {
    flexDirection: "row",
    gap: spacing.base,
    padding: spacing.md,
    borderRadius: radius.xl,
    backgroundColor: colors.surfaceContainerLow,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
  },
  infoBody: {
    flex: 1,
    gap: 4,
  },
  infoTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  infoText: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
  },
  posterRow: {
    gap: spacing.base,
  },
  posterCard: {
    width: 148,
    gap: spacing.xs,
  },
  continueCard: {
    width: 148,
    gap: spacing.sm,
  },
  posterImage: {
    width: 148,
    height: 216,
    alignItems: "center",
    justifyContent: "center",
    borderRadius: radius.xl,
    overflow: "hidden",
    backgroundColor: colors.surfaceContainer,
  },
  posterImageFill: {
    width: "100%",
    height: "100%",
  },
  posterTitle: {
    ...type.labelMd,
    color: colors.onSurface,
  },
  posterSubtitle: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  badge: {
    alignSelf: "flex-start",
    borderRadius: radius.full,
    paddingHorizontal: 8,
    paddingVertical: 4,
    backgroundColor: colors.surfaceContainerHigh,
  },
  badgeText: {
    ...type.labelSm,
    color: colors.onSurfaceVariant,
  },
  progressTrack: {
    width: "100%",
    height: 6,
    backgroundColor: colors.surfaceContainerHigh,
    borderRadius: radius.full,
    overflow: "hidden",
  },
  progressFill: {
    height: "100%",
    backgroundColor: colors.primary,
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
});
