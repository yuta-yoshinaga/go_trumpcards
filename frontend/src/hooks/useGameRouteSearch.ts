import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { type GameRoute, gameRoutes } from '../constants/gameRoutes';

/** Result of a bilingual game route search. */
export interface UseGameRouteSearchResult {
  /** Matching routes when `searchTerm` is non-empty, otherwise `null` (caller should fall back to its category view). */
  filteredRoutes: GameRoute[] | null;
  /** Set of paths that match — convenience for callers like DesktopSidebar that already iterate `gameRoutes`. */
  filteredPaths: Set<string> | null;
}

/**
 * Filters {@link gameRoutes} by a bilingual (ja/en) substring match against
 * each route's translated label. NavBar and DesktopSidebar both call this so
 * the search behavior stays in lockstep when new games are added.
 */
export function useGameRouteSearch(searchTerm: string): UseGameRouteSearchResult {
  const { i18n } = useTranslation('common');

  // Both i18n.t calls pass explicit lng overrides, so the index is language-
  // independent. Depending on i18n.t (not the whole i18n instance) keeps the
  // index from rebuilding on every JA ↔ EN toggle (PR #1572 review).
  const searchableRoutes = useMemo(
    () =>
      gameRoutes.map((route) => ({
        route,
        ja: i18n.t(route.labelKey, { lng: 'ja', ns: 'common' }).toLowerCase(),
        en: i18n.t(route.labelKey, { lng: 'en', ns: 'common' }).toLowerCase(),
      })),
    [i18n.t],
  );

  return useMemo(() => {
    if (!searchTerm) return { filteredRoutes: null, filteredPaths: null };
    const lower = searchTerm.toLowerCase();
    const matches = searchableRoutes
      .filter(({ ja, en }) => ja.includes(lower) || en.includes(lower))
      .map(({ route }) => route);
    return {
      filteredRoutes: matches,
      filteredPaths: new Set(matches.map((r) => r.path)),
    };
  }, [searchTerm, searchableRoutes]);
}
