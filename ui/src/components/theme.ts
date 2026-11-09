import { createDarkTheme, createLightTheme, type BrandVariants } from '@fluentui/react-components';

const brand: BrandVariants = {
  10: '#090f1a',
  20: '#101d31',
  30: '#172a46',
  40: '#20395d',
  50: '#294977',
  60: '#30558f',
  70: '#3966a9',
  80: '#4778bf',
  90: '#5a8bd2',
  100: '#719ee0',
  110: '#8db3eb',
  120: '#a6c7f5',
  130: '#bed8fa',
  140: '#d3e5fc',
  150: '#e3efff',
  160: '#f2f7ff',
};

const cornerRadius = '8px';
const sharedTokens = {
  borderRadiusSmall: cornerRadius,
  borderRadiusMedium: cornerRadius,
  borderRadiusLarge: cornerRadius,
  borderRadiusXLarge: cornerRadius,
  fontFamilyBase: "system-ui, 'Noto Sans SC', 'Microsoft YaHei', sans-serif",
};

export const lightTheme = {
  ...createLightTheme(brand),
  ...sharedTokens,
  colorNeutralBackground1: '#ffffff',
  colorNeutralBackground2: '#edf1f6',
  colorNeutralBackground3: '#f5f7fa',
  colorNeutralForeground1: '#172536',
  colorNeutralForeground3: '#4d5d70',
  colorBrandForeground1: '#30558f',
  colorBrandBackground: brand[60],
  colorBrandBackgroundHover: brand[50],
  colorBrandBackgroundPressed: brand[40],
};

export const darkTheme = {
  ...createDarkTheme(brand),
  ...sharedTokens,
  colorNeutralBackground1: '#1a2634',
  colorNeutralBackground2: '#202e3e',
  colorNeutralBackground3: '#101923',
  colorNeutralForeground1: '#f2f6fc',
  colorNeutralForeground3: '#b2c1d2',
  colorBrandForeground1: '#a6c7f5',
  colorBrandBackground: brand[60],
  colorBrandBackgroundHover: brand[50],
  colorBrandBackgroundPressed: brand[40],
};
