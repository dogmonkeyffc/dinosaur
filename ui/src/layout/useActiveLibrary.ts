import { useEffect } from 'react';
import { useOutletContext } from 'react-router-dom';

export function useActiveLibrary(libraryID?: number) {
  const { setActiveLibrary } = useOutletContext<{ setActiveLibrary: (id: number) => void }>();
  useEffect(() => {
    if (libraryID != null) setActiveLibrary(libraryID);
  }, [libraryID, setActiveLibrary]);
}
