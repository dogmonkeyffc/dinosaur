import { useState, type ReactNode } from 'react';
import {
  ActivityIndicator,
  Platform,
  Pressable,
  ScrollView,
  StatusBar,
  Text,
  TextInput,
  View,
  type TextInputProps,
} from 'react-native';
import { colors, styles } from './theme';

export function Screen({
  children,
  title,
  tv = false,
  scrollY = 0,
  scrolled,
}: {
  children: ReactNode;
  title?: string;
  tv?: boolean;
  scrollY?: number;
  scrolled?: (offset: number) => void;
}) {
  return (
    <View
      style={[
        styles.screen,
        { paddingTop: tv ? 0 : Platform.OS === 'android' ? StatusBar.currentHeight : 52 },
      ]}
    >
      <StatusBar hidden={tv} barStyle="light-content" />
      <ScrollView
        keyboardShouldPersistTaps="handled"
        contentOffset={{ x: 0, y: scrollY }}
        onScroll={(event) => scrolled?.(event.nativeEvent.contentOffset.y)}
        scrollEventThrottle={100}
        contentContainerStyle={[styles.content, tv && { padding: 40 }]}
      >
        {title && (
          <Text accessibilityRole="header" style={styles.title}>
            {title}
          </Text>
        )}
        {children}
      </ScrollView>
    </View>
  );
}

export function Button({
  children,
  onPress,
  disabled,
  primary,
  preferred,
  label,
}: {
  children: ReactNode;
  onPress: () => void;
  disabled?: boolean;
  primary?: boolean;
  preferred?: boolean;
  label?: string;
}) {
  const [focused, setFocused] = useState(false);
  return (
    <Pressable
      accessibilityRole="button"
      accessibilityLabel={label}
      accessibilityState={{ disabled: Boolean(disabled) }}
      onPress={onPress}
      disabled={disabled}
      hasTVPreferredFocus={preferred}
      onFocus={() => setFocused(true)}
      onBlur={() => setFocused(false)}
      style={({ pressed }) => ({
        minHeight: 52,
        paddingHorizontal: 20,
        paddingVertical: 13,
        justifyContent: 'center',
        alignItems: 'center',
        borderRadius: 12,
        borderWidth: 2,
        borderColor: focused ? '#ffffff' : primary ? colors.accent : colors.border,
        backgroundColor: primary || focused ? colors.accent : colors.surface,
        opacity: disabled ? 0.45 : pressed ? 0.7 : 1,
      })}
    >
      <Text
        style={{
          color: primary || focused ? colors.background : colors.text,
          fontWeight: '600',
          fontSize: 16,
        }}
      >
        {children}
      </Text>
    </Pressable>
  );
}

export function Input(props: TextInputProps) {
  return (
    <TextInput
      placeholderTextColor={colors.muted}
      autoCapitalize="none"
      autoCorrect={false}
      {...props}
      style={[styles.input, props.style]}
    />
  );
}
export function ErrorText({ error }: { error?: string }) {
  return error ? (
    <Text accessibilityRole="alert" style={styles.error}>
      {error}
    </Text>
  ) : null;
}
export function Busy() {
  return <ActivityIndicator color={colors.accent} size="large" />;
}
