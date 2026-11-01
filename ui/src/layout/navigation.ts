import type { Library } from '../lib/types';

export function libraryPath(library: Pick<Library, 'id' | 'type'>) {
  return `/${library.type === 'movie' ? 'movies' : 'series'}?library_id=${library.id}`;
}
