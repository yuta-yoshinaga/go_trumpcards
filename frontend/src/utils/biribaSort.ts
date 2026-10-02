import type { HandSortMode } from './handDisplaySort';

/** Hand sort modes offered in the Biriba footer. */
export type BiribaSortMode = HandSortMode;

/** localStorage key persisting the player's chosen Biriba hand-sort mode. */
export const BIRIBA_SORT_STORAGE_KEY = 'biriba-sort-mode';

/** Reads the persisted sort mode from localStorage, defaulting to `original`. */
export function loadBiribaSortMode(): BiribaSortMode {
  try {
    const v = localStorage.getItem(BIRIBA_SORT_STORAGE_KEY);
    if (v === 'rank' || v === 'suit' || v === 'original') return v;
  } catch {
    // Ignore storage access errors (e.g. privacy mode) and fall back to default.
  }
  return 'original';
}

/** Persists the chosen sort mode to localStorage, ignoring storage failures. */
export function saveBiribaSortMode(mode: BiribaSortMode): void {
  try {
    localStorage.setItem(BIRIBA_SORT_STORAGE_KEY, mode);
  } catch {
    // Ignore storage access errors (e.g. privacy mode).
  }
}
