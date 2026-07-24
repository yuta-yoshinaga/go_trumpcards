import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Pyramid Solitaire best-record statistics. */
export const PYRAMID_STATS_KEY = 'trumpcards-pyramid-stats';

/** Persisted best-record statistics for Pyramid. */
export interface PyramidStats {
  /** Games finished (cleared or given up / stuck). */
  plays: number;
  /** Games cleared. */
  wins: number;
  /** Fewest moves across cleared games, or null if never cleared. */
  fewestMoves: number | null;
}

/** Outcome of a single finished Pyramid game. */
export interface PyramidResult {
  won: boolean;
  moves: number;
}

/** Returns a zeroed stats record. */
export function emptyPyramidStats(): PyramidStats {
  return { plays: 0, wins: 0, fewestMoves: null };
}

function isValidStats(value: unknown): value is PyramidStats {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  return (
    typeof s.plays === 'number' &&
    typeof s.wins === 'number' &&
    (s.fewestMoves === null || typeof s.fewestMoves === 'number')
  );
}

/**
 * Pure reducer: folds a finished-game result into the stats and reports whether it
 * set a new fewest-moves record. Fewest moves only advances on a positive-move clear.
 */
export function applyPyramidResult(
  stats: PyramidStats,
  result: PyramidResult,
): { stats: PyramidStats; newBest: boolean } {
  const next: PyramidStats = {
    plays: stats.plays + 1,
    wins: stats.wins + (result.won ? 1 : 0),
    fewestMoves: stats.fewestMoves,
  };
  let newBest = false;
  if (result.won && result.moves > 0 && (stats.fewestMoves === null || result.moves < stats.fewestMoves)) {
    next.fewestMoves = result.moves;
    newBest = true;
  }
  return { stats: next, newBest };
}

/** Reads and validates the stats from localStorage; returns a zeroed record on any error. */
export const readPyramidStats = createFlatStatsReader(PYRAMID_STATS_KEY, emptyPyramidStats, isValidStats);

const store = createLocalStorageStats<PyramidStats, PyramidResult, boolean>({
  key: PYRAMID_STATS_KEY,
  read: readPyramidStats,
  reduce: (prev, result) => {
    const { stats, newBest } = applyPyramidResult(prev, result);
    return { stats, update: newBest };
  },
});

/**
 * Hook that persists Pyramid best-record statistics in localStorage and records
 * finished games. `recordResult` returns whether the game set a new fewest-moves
 * record so the page can surface a badge.
 */
export const usePyramidStats = store.useStats;
