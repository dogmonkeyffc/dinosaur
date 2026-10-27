import { StyleSheet } from 'react-native';

export const colors = {
  background: '#111a17',
  surface: '#1e2c25',
  text: '#f0f4e9',
  muted: '#a9b5ac',
  accent: '#c2df9e',
  border: '#3c5144',
  error: '#ffb6a7',
};
export const styles = StyleSheet.create({
  screen: { flex: 1, backgroundColor: colors.background },
  content: { padding: 24, gap: 18, flexGrow: 1 },
  title: { color: colors.text, fontSize: 28, fontWeight: '700' },
  subtitle: { color: colors.text, fontSize: 20, fontWeight: '600' },
  text: { color: colors.text, fontSize: 16, lineHeight: 25 },
  muted: { color: colors.muted, fontSize: 14, lineHeight: 22 },
  error: { color: colors.error, fontSize: 15, lineHeight: 24 },
  card: {
    backgroundColor: colors.surface,
    borderRadius: 16,
    borderWidth: 1,
    borderColor: colors.border,
    padding: 20,
    gap: 14,
  },
  row: { flexDirection: 'row', flexWrap: 'wrap', gap: 12, alignItems: 'center' },
  input: {
    color: colors.text,
    backgroundColor: colors.surface,
    borderColor: colors.border,
    borderWidth: 1,
    borderRadius: 12,
    padding: 16,
    fontSize: 17,
    minHeight: 54,
  },
  centered: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    padding: 32,
    gap: 20,
    backgroundColor: colors.background,
  },
});
