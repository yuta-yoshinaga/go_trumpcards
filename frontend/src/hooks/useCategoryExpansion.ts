import { useCallback, useState } from 'react';

/** localStorage key for sidebar / nav category accordion state. */
export const CATEGORY_EXPANSION_KEY = 'trumpcards-category-expansion';

/**
 * Per-category expansion state. Missing keys are treated as `true`
 * (expanded) so newly-added categories surface to existing visitors,
 * and so the very first visit shows everything.
 */
type ExpansionMap = Record<string, boolean>;

/** Reads the saved expansion map, falling back to an empty (= all open) map. */
function readMap(): ExpansionMap {
  try {
    const raw = localStorage.getItem(CATEGORY_EXPANSION_KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (!parsed || typeof parsed !== 'object' || Array.isArray(parsed)) return {};
    const result: ExpansionMap = {};
    for (const [k, v] of Object.entries(parsed as Record<string, unknown>)) {
      if (typeof v === 'boolean') result[k] = v;
    }
    return result;
  } catch {
    return {};
  }
}

/**
 * Hook that tracks whether each navigation category accordion is
 * expanded. Defaults every category to `true` on first visit (so all
 * 77 games are discoverable), and persists user-toggled collapses to
 * localStorage so repeat visitors see their preferred layout.
 */
export function useCategoryExpansion() {
  const [map, setMap] = useState<ExpansionMap>(readMap);

  const isExpanded = useCallback((labelKey: string) => map[labelKey] !== false, [map]);

  const setExpanded = useCallback(
    (labelKey: string, open: boolean) => {
      // Missing entries are treated as "open" (the implicit default), so a
      // toggle that lands on the same value is a no-op — including the very
      // first interaction where map[labelKey] is undefined and open is true.
      const current = map[labelKey] ?? true;
      if (current === open) return;
      const next = { ...map, [labelKey]: open };
      // Side effect outside the setState updater: React may invoke updater
      // functions more than once (e.g. StrictMode in development), so calling
      // localStorage.setItem inside one would write twice for the same
      // transition. Computing next state up here keeps the updater pure.
      try {
        localStorage.setItem(CATEGORY_EXPANSION_KEY, JSON.stringify(next));
      } catch {
        /* private mode / storage full / etc. */
      }
      setMap(next);
    },
    [map],
  );

  return { isExpanded, setExpanded };
}
