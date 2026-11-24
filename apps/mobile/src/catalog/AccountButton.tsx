import { Pressable, Text, View } from 'react-native';
import { homeStyles as css } from './homeStyles';

export function AccountButton({ username, open }: { username: string; open: () => void }) {
  return (
    <Pressable
      accessibilityLabel="账号与设置"
      accessibilityRole="button"
      hitSlop={8}
      onPress={open}
      style={css.avatarRing}
    >
      <View style={css.avatar}>
        <Text style={css.avatarText}>{username.slice(0, 1).toUpperCase()}</Text>
      </View>
    </Pressable>
  );
}
