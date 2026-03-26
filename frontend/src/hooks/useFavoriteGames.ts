import { useCallback, useState } from 'react';

/** localStorage key for favorite games. */
export const FAVORITE_GAMES_KEY = 'trumpcards-favorite-games';

/** Reads the favorites list from localStorage. */
function readFavorites(): string[] {
  try {
    const raw = localStorage.getItem(FAVORITE_GAMES_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((p): p is string => typeof p === 'string');
  } catch {
    return [];
  }
}

/** Hook that manages favorite game paths in localStorage. */
export function useFavoriteGames() {
  const [favorites, setFavorites] = useState<string[]>(readFavorites);

  const isFavorite = useCallback((path: string) => favorites.includes(path), [favorites]);

  const toggleFavorite = useCallback((path: string) => {
    setFavorites((prev) => {
      const next = prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path];
      try {
        localStorage.setItem(FAVORITE_GAMES_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable */
      }
      return next;
    });
  }, []);

  return { favorites, isFavorite, toggleFavorite };
}
