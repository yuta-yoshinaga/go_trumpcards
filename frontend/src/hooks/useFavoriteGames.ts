import { useCallback, useState } from 'react';
import { gameRoutes } from '../constants/gameRoutes';

/** localStorage key for favorite games. */
export const FAVORITE_GAMES_KEY = 'trumpcards-favorite-games';

const knownPaths = new Set(gameRoutes.map((r) => r.path));

/** Maximum number of favorite games to store. */
const MAX_FAVORITES = 10;

/** Reads the favorites list from localStorage, filtering out stale paths. */
function readFavorites(): string[] {
  try {
    const raw = localStorage.getItem(FAVORITE_GAMES_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((p): p is string => typeof p === 'string' && knownPaths.has(p));
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
      const next = prev.includes(path) ? prev.filter((p) => p !== path) : [...prev, path].slice(0, MAX_FAVORITES);
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
