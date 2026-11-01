import { useEffect, useState } from 'react';

export const scrapingStates = ['', 'matched', 'unmatched', 'needs_review', 'failed', 'warning'];

interface Filters {
  library: string;
  status: string;
}

function readFilters(key: string): Filters {
  try {
    const saved = JSON.parse(localStorage.getItem(key) || '{}');
    return {
      library: typeof saved?.library === 'string' && /^[1-9]\d*$/.test(saved.library) ? saved.library : '',
      status: typeof saved?.status === 'string' && scrapingStates.includes(saved.status) ? saved.status : '',
    };
  } catch {
    return { library: '', status: '' };
  }
}

export function useScrapingFilters(userID: number) {
  const key = `dinosaur.scraping.filters.${userID}`;
  const [filters, setFilters] = useState(() => readFilters(key));
  useEffect(() => {
    try {
      localStorage.setItem(key, JSON.stringify(filters));
    } catch {
      // Filtering still works when browser storage is unavailable.
    }
  }, [key, filters]);

  return {
    ...filters,
    setLibrary: (library: string) => setFilters((current) => ({ ...current, library })),
    setStatus: (status: string) => setFilters((current) => ({ ...current, status })),
  };
}
