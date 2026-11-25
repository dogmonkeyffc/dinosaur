import type { ReactNode } from 'react';
import { Platform, Pressable, ScrollView, StatusBar, Text, View } from 'react-native';
import { pageStyles as css } from './pageStyles';

export function MobilePage({
  title,
  close,
  children,
}: {
  title: string;
  close?: () => void;
  children: ReactNode;
}) {
  return (
    <View style={[css.screen, { paddingTop: Platform.OS === 'android' ? StatusBar.currentHeight : 52 }]}>
      <StatusBar barStyle="light-content" />
      <View style={css.toolbar}>
        {close ? (
          <Pressable accessibilityRole="button" accessibilityLabel="返回" onPress={close} style={css.back}>
            <Text style={css.backIcon}>←</Text>
          </Pressable>
        ) : (
          <View style={css.back} />
        )}
        <Text accessibilityRole="header" style={css.title}>
          {title}
        </Text>
      </View>
      <ScrollView keyboardShouldPersistTaps="handled" contentContainerStyle={css.content}>
        {children}
      </ScrollView>
    </View>
  );
}

export function MobileButton({
  children,
  press,
  disabled = false,
  primary = false,
}: {
  children: string;
  press: () => void;
  disabled?: boolean;
  primary?: boolean;
}) {
  return (
    <Pressable
      accessibilityRole="button"
      disabled={disabled}
      onPress={press}
      style={({ pressed }) => [css.button, primary && css.primary, (disabled || pressed) && css.dim]}
    >
      <Text style={[css.buttonText, primary && css.primaryText]}>{children}</Text>
    </Pressable>
  );
}
