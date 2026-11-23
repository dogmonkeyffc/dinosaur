import type { MediaLibrary, Movie } from '@dinosaur/contracts';

export function parseHiddenLibraries(raw: string | null): number[] {
  if (raw === null) return [];
  const value: unknown = JSON.parse(raw);
  if (!Array.isArray(value) || value.some((id) => !Number.isSafeInteger(id) || id <= 0)) {
    throw new Error('首页显示设置无效。');
  }
  return [...new Set(value as number[])];
}

export function homeLibraries(libraries: MediaLibrary[], hidden: number[]) {
  return libraries.filter((library) => !hidden.includes(library.id));
}

export function homeRecent(items: Movie[], libraries: MediaLibrary[], hidden: number[]) {
  const visible = new Set(homeLibraries(libraries, hidden).map((library) => library.id));
  return items.filter((item: Movie & { library_id?: number }) =>
    item.library_id === undefined ? hidden.length === 0 : visible.has(item.library_id),
  );
}
