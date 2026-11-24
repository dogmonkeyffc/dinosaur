import type { ImageStyle, TextStyle, ViewStyle } from 'react-native';

export const uiScale = 0.85;

export function dp(value: number): number {
  return Math.round(value * uiScale * 100) / 100;
}

const lengths = new Set([
  'width',
  'height',
  'minWidth',
  'maxWidth',
  'minHeight',
  'maxHeight',
  'top',
  'right',
  'bottom',
  'left',
  'start',
  'end',
  'padding',
  'paddingHorizontal',
  'paddingVertical',
  'paddingTop',
  'paddingRight',
  'paddingBottom',
  'paddingLeft',
  'paddingStart',
  'paddingEnd',
  'margin',
  'marginHorizontal',
  'marginVertical',
  'marginTop',
  'marginRight',
  'marginBottom',
  'marginLeft',
  'marginStart',
  'marginEnd',
  'gap',
  'rowGap',
  'columnGap',
  'borderWidth',
  'borderTopWidth',
  'borderRightWidth',
  'borderBottomWidth',
  'borderLeftWidth',
  'borderRadius',
  'borderTopLeftRadius',
  'borderTopRightRadius',
  'borderBottomLeftRadius',
  'borderBottomRightRadius',
  'fontSize',
  'lineHeight',
  'letterSpacing',
]);

// Scale visual lengths only; preserve flex, opacity, ratios, percentages and transforms.
export function compactStyles<T extends Record<string, ViewStyle | TextStyle | ImageStyle>>(styles: T): T {
  return Object.fromEntries(
    Object.entries(styles).map(([name, style]) => [
      name,
      Object.fromEntries(
        Object.entries(style).map(([property, value]) => [
          property,
          lengths.has(property) && typeof value === 'number' ? dp(value) : value,
        ]),
      ),
    ]),
  ) as T;
}
