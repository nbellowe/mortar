import React, { useState } from "react";
import {
  ActivityIndicator,
  StyleSheet,
  Text,
  TextInput,
  TouchableOpacity,
  View,
} from "react-native";

import { useAuth } from "./auth-context";
import { colors, radius, spacing, type } from "@/theme/tokens";

export function LoginScreen() {
  const { login } = useAuth();
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState<string | null>(null);
  const [submitting, setSubmitting] = useState(false);

  async function handleSubmit() {
    setSubmitting(true);
    setError(null);
    try {
      await login(username.trim(), password);
    } catch (err) {
      setError(err instanceof Error ? err.message : "Sign-in failed");
    } finally {
      setSubmitting(false);
    }
  }

  return (
    <View style={s.container}>
      <View style={s.card}>
        <Text style={s.eyebrow}>Mortar</Text>
        <Text style={s.title}>Sign in to your media front door</Text>
        <Text style={s.body}>
          Use the Mortar username and password configured by your operator.
        </Text>

        <View style={s.form}>
          <TextInput
            autoCapitalize="none"
            autoCorrect={false}
            placeholder="Username"
            placeholderTextColor={colors.outline}
            style={s.input}
            value={username}
            onChangeText={setUsername}
          />
          <TextInput
            autoCapitalize="none"
            autoCorrect={false}
            placeholder="Password"
            placeholderTextColor={colors.outline}
            style={s.input}
            value={password}
            onChangeText={setPassword}
            secureTextEntry
            onSubmitEditing={() => {
              void handleSubmit();
            }}
          />
          {error ? <Text style={s.error}>{error}</Text> : null}
          <TouchableOpacity
            style={[
              s.button,
              (!username.trim() || !password) && s.buttonDisabled,
            ]}
            onPress={() => {
              void handleSubmit();
            }}
            disabled={!username.trim() || !password || submitting}
          >
            {submitting ? (
              <ActivityIndicator
                size="small"
                color={colors.onPrimaryContainer}
              />
            ) : (
              <Text style={s.buttonText}>Sign in</Text>
            )}
          </TouchableOpacity>
        </View>
      </View>
    </View>
  );
}

const s = StyleSheet.create({
  container: {
    flex: 1,
    backgroundColor: colors.background,
    alignItems: "center",
    justifyContent: "center",
    padding: spacing.gutter,
  },
  card: {
    width: "100%",
    maxWidth: 420,
    backgroundColor: colors.surfaceContainerLow,
    borderRadius: radius.xl,
    borderWidth: StyleSheet.hairlineWidth,
    borderColor: colors.outlineVariant,
    padding: spacing.gutter,
    gap: spacing.base,
  },
  eyebrow: {
    ...type.labelMd,
    color: colors.primary,
    textTransform: "uppercase",
    letterSpacing: 1,
  },
  title: {
    ...type.headlineLg,
    color: colors.onSurface,
  },
  body: {
    ...type.bodyMd,
    color: colors.onSurfaceVariant,
  },
  form: {
    gap: spacing.base,
    marginTop: spacing.base,
  },
  input: {
    borderRadius: radius.lg,
    borderWidth: 1,
    borderColor: colors.outlineVariant,
    backgroundColor: colors.surface,
    color: colors.onSurface,
    paddingHorizontal: spacing.md,
    paddingVertical: spacing.sm,
  },
  button: {
    borderRadius: radius.full,
    backgroundColor: colors.primaryContainer,
    alignItems: "center",
    justifyContent: "center",
    minHeight: 44,
  },
  buttonDisabled: {
    opacity: 0.55,
  },
  buttonText: {
    ...type.labelMd,
    color: colors.onPrimaryContainer,
  },
  error: {
    ...type.labelSm,
    color: colors.error,
  },
});
