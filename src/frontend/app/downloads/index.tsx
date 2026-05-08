import React from 'react';
import { StyleSheet, Text, View } from 'react-native';
import { Ionicons } from '@expo/vector-icons';
import { colors, spacing, type } from '@/theme/tokens';

export default function DownloadsScreen() {
  return (
    <View style={s.container}>
      <View style={s.topBar}>
        <Ionicons name="cloud-download" size={22} color={colors.primaryFixedDim} />
        <Text style={s.topBarTitle}>Downloads</Text>
      </View>
      <View style={s.centered}>
        <Ionicons name="cloud-download-outline" size={48} color={colors.outlineVariant} />
        <Text style={s.emptyTitle}>Download queue coming soon</Text>
        <Text style={s.emptyBody}>Track active and queued downloads across your connected services.</Text>
      </View>
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
