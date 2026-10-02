import type { HandSortMode } from './handDisplaySort';

/** Hand sort modes offered in the Burraco footer. */
export type BurracoSortMode = HandSortMode;

/** localStorage key persisting the player's chosen Burraco hand-sort mode. */
export const BURRACO_SORT_STORAGE_KEY = 'burraco-sort-mode';

/** Reads the persisted sort mode from localStorage, defaulting to `original`. */
export function loadBurracoSortMode(): BurracoSortMode {
  try {
    const v = localStorage.getItem(BURRACO_SORT_STORAGE_KEY);
    if (v === 'rank' || v === 'suit' || v === 'original') return v;
  } catch {
    // Ignore storage access errors (e.g. privacy mode) and fall back to default.
  }
  return 'original';
}

/** Persists the chosen sort mode to localStorage, ignoring storage failures. */
export function saveBurracoSortMode(mode: BurracoSortMode): void {
  try {
    localStorage.setItem(BURRACO_SORT_STORAGE_KEY, mode);
  } catch {
    // Ignore storage access errors (e.g. privacy mode).
  }
}
