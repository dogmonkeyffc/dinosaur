import { StyleSheet, View } from 'react-native';
import { compactStyles } from '../ui/density';
import { palette } from './homeStyles';

export function Icon({
  name,
  color = palette.muted,
}: {
  name: 'home' | 'library' | 'search' | 'history';
  color?: string;
}) {
  return (
    <View importantForAccessibility="no" style={css.icon}>
      {name === 'history' ? (
        <View style={[css.clock, { borderColor: color }]}>
          <View style={[css.hour, { backgroundColor: color }]} />
          <View style={[css.minute, { backgroundColor: color }]} />
        </View>
      ) : name === 'search' ? (
        <>
          <View style={[css.lens, { borderColor: color }]} />
          <View style={[css.handle, { backgroundColor: color }]} />
        </>
      ) : name === 'library' ? (
        <View style={css.grid}>
          {[0, 1, 2, 3].map((key) => (
            <View key={key} style={[css.square, { borderColor: color }]} />
          ))}
        </View>
      ) : (
        <>
          <View style={[css.roof, { borderColor: color }]} />
          <View style={[css.house, { borderColor: color }]} />
        </>
      )}
    </View>
  );
}

const css = StyleSheet.create(
  compactStyles({
    icon: { width: 24, height: 24, alignItems: 'center', justifyContent: 'center' },
    clock: { width: 22, height: 22, borderRadius: 11, borderWidth: 2 },
    hour: { position: 'absolute', width: 2, height: 7, top: 3, left: 8 },
    minute: { position: 'absolute', width: 6, height: 2, top: 8, left: 8 },
    lens: { width: 16, height: 16, borderRadius: 8, borderWidth: 2, position: 'absolute', left: 2, top: 2 },
    handle: {
      width: 10,
      height: 2,
      position: 'absolute',
      left: 15,
      top: 19,
      transform: [{ rotate: '45deg' }],
    },
    grid: { width: 22, height: 22, flexDirection: 'row', flexWrap: 'wrap', gap: 4 },
    square: { width: 9, height: 9, borderWidth: 2, borderRadius: 2 },
    roof: {
      width: 15,
      height: 15,
      borderLeftWidth: 2,
      borderTopWidth: 2,
      transform: [{ rotate: '45deg' }],
      top: -1,
    },
    house: {
      width: 17,
      height: 14,
      borderWidth: 2,
      borderTopWidth: 0,
      borderRadius: 2,
      position: 'absolute',
      bottom: 1,
    },
  }),
);
