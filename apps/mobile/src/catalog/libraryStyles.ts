import { StyleSheet } from 'react-native';
import { palette } from './homeStyles';

export const libraryStyles = StyleSheet.create({
  content: { flexGrow: 1, paddingBottom: 28, gap: 24 },
  hero: {
    backgroundColor: '#243c39',
    overflow: 'hidden',
    borderBottomLeftRadius: 26,
    borderBottomRightRadius: 26,
  },
  backdrop: { position: 'absolute', width: '100%', height: '100%', opacity: 0.65 },
  shade: { position: 'absolute', bottom: 0, height: '68%', width: '100%', backgroundColor: '#0814138c' },
  header: {
    flexDirection: 'row',
    alignItems: 'center',
    justifyContent: 'space-between',
    paddingHorizontal: 24,
  },
  logo: { color: '#d2eedb', fontSize: 30, fontWeight: '800' },
  heroBody: { flex: 1, justifyContent: 'flex-end', padding: 24, gap: 22 },
  title: { color: '#fff', fontSize: 32, lineHeight: 40, fontWeight: '600', maxWidth: '94%' },
  action: { flexDirection: 'row', alignItems: 'center', justifyContent: 'space-between', gap: 18 },
  thumbnails: { flexDirection: 'row', gap: 8 },
  thumbnail: { width: 36, height: 50, borderRadius: 5, overflow: 'hidden' },
  button: {
    minWidth: 90,
    minHeight: 48,
    borderRadius: 26,
    backgroundColor: '#e3e8fc',
    alignItems: 'center',
    justifyContent: 'center',
    paddingHorizontal: 20,
  },
  buttonText: { color: '#17213a', fontSize: 16, fontWeight: '600' },
  list: { gap: 8, paddingHorizontal: 24 },
  listTitle: { color: palette.text, fontSize: 21, fontWeight: '600', marginBottom: 12 },
  row: { flexDirection: 'row', gap: 18, alignItems: 'center', minHeight: 122, paddingVertical: 7 },
  cover: { width: 68, height: 102, borderRadius: 10, overflow: 'hidden' },
  rowTitle: { flex: 1, fontSize: 17, lineHeight: 25, color: palette.text },
  arrow: { color: palette.muted, fontSize: 25, paddingLeft: 6 },
  feedback: { paddingHorizontal: 24, gap: 12 },
  retry: { minHeight: 48, justifyContent: 'center' },
});
