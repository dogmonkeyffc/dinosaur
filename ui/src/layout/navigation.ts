import type { Library } from '../lib/types';

export function libraryPath(library: Pick<Library, 'id' | 'type'>) {
  return `/${library.type === 'tv' ? 'series' : library.type === 'video' ? 'videos' : 'movies'}?library_id=${library.id}`;
}
