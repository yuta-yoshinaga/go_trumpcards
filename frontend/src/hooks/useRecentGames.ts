import { useEffect, useState } from 'react';
import { gameRoutes } from '../constants/gameRoutes';

/** localStorage key for recently played games. */
export const RECENT_GAMES_KEY = 'trumpcards-recent-games';

/** Maximum number of recent games to store. */
const MAX_RECENT = 5;

const knownPaths = new Set(gameRoutes.map((r) => r.path));

/** Reads the recent games list from localStorage. */
function readRecent(): string[] {
  try {
    const raw = localStorage.getItem(RECENT_GAMES_KEY);
    if (!raw) return [];
    const parsed: unknown = JSON.parse(raw);
    if (!Array.isArray(parsed)) return [];
    return parsed.filter((p): p is string => typeof p === 'string' && knownPaths.has(p));
  } catch {
    return [];
  }
}

/** Hook that tracks recently played game paths in localStorage. */
export function useRecentGames(pathname: string): string[] {
  const [recent, setRecent] = useState<string[]>(readRecent);

  useEffect(() => {
    if (!knownPaths.has(pathname)) return;
    setRecent((prev) => {
      const filtered = prev.filter((p) => p !== pathname);
      const next = [pathname, ...filtered].slice(0, MAX_RECENT);
      try {
        localStorage.setItem(RECENT_GAMES_KEY, JSON.stringify(next));
      } catch {
        /* storage unavailable */
      }
      return next;
    });
  }, [pathname]);

  return recent;
}
