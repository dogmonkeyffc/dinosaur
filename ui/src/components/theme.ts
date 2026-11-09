import { createLightTheme, type BrandVariants } from '@fluentui/react-components';

const brand: BrandVariants = {
  10: '#08140d',
  20: '#14261b',
  30: '#1d3826',
  40: '#254831',
  50: '#2d563c',
  60: '#346448',
  70: '#3c7253',
  80: '#44815f',
  90: '#518f6c',
  100: '#639d7a',
  110: '#77ac8a',
  120: '#8dbb9b',
  130: '#a4caad',
  140: '#bcdbbf',
  150: '#d4e9d3',
  160: '#edf6eb',
};

export const dinosaurTheme = {
  ...createLightTheme(brand),
  fontFamilyBase: "'Segoe UI', 'Noto Sans SC', 'Microsoft YaHei', sans-serif",
  colorBrandBackground: brand[60],
  colorBrandBackgroundHover: brand[50],
  colorBrandBackgroundPressed: brand[40],
};
