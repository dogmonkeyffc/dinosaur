import { Pressable, ScrollView, Text, View } from 'react-native';
import type { PlaybackIconName } from './PlaybackIcon';
import { phoneStyles as css } from './phoneStyles';

export interface PlayerSetting {
  icon: PlaybackIconName;
  label: string;
  disabled?: boolean;
  options: { label: string; selected: boolean; press: () => void }[];
}

export function PlayerMenu({ setting, close }: { setting: PlayerSetting; close: () => void }) {
  return (
    <View style={css.menu} accessibilityViewIsModal>
      <ScrollView style={{ maxHeight: 210 }}>
        {setting.options.map((option, index) => (
          <Pressable
            key={index}
            accessibilityRole="radio"
            accessibilityState={{ checked: option.selected }}
            accessibilityLabel={option.label}
            onPress={() => {
              if (!option.selected) option.press();
              close();
            }}
            style={[css.menuOption, option.selected && css.selectedOption]}
          >
            <Text style={[css.menuLabel, option.selected && css.selectedLabel]}>{option.label}</Text>
            {option.selected && <Text style={css.selectedLabel}>✓</Text>}
          </Pressable>
        ))}
      </ScrollView>
    </View>
  );
}
