import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for TriPeaks Solitaire best-record statistics. */
export const TRIPEAKS_STATS_KEY = 'trumpcards-tripeaks-stats';

/** Persisted best-record statistics for TriPeaks. */
export interface TriPeaksStats {
  /** Games finished (cleared or given up / stuck). */
  plays: number;
  /** Games cleared. */
  wins: number;
  /** Highest score ever recorded, or null if never scored. */
  bestScore: number | null;
}

/** Outcome of a single finished TriPeaks game. */
export interface TriPeaksResult {
  won: boolean;
  score: number;
}

/** Returns a zeroed stats record. */
export function emptyTriPeaksStats(): TriPeaksStats {
  return { plays: 0, wins: 0, bestScore: null };
}

function isValidStats(value: unknown): value is TriPeaksStats {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  return (
    typeof s.plays === 'number' &&
    typeof s.wins === 'number' &&
    (s.bestScore === null || typeof s.bestScore === 'number')
  );
}

/**
 * Pure reducer: folds a finished-game result into the stats and reports whether it
 * set a new best score. Best score only advances on a positive score.
 */
export function applyTriPeaksResult(
  stats: TriPeaksStats,
  result: TriPeaksResult,
): { stats: TriPeaksStats; newBest: boolean } {
  const next: TriPeaksStats = {
    plays: stats.plays + 1,
    wins: stats.wins + (result.won ? 1 : 0),
    bestScore: stats.bestScore,
  };
  let newBest = false;
  if (result.score > 0 && (stats.bestScore === null || result.score > stats.bestScore)) {
    next.bestScore = result.score;
    newBest = true;
  }
  return { stats: next, newBest };
}

/** Reads and validates the stats from localStorage; returns a zeroed record on any error. */
export const readTriPeaksStats = createFlatStatsReader(TRIPEAKS_STATS_KEY, emptyTriPeaksStats, isValidStats);

const store = createLocalStorageStats<TriPeaksStats, TriPeaksResult, boolean>({
  key: TRIPEAKS_STATS_KEY,
  read: readTriPeaksStats,
  reduce: (prev, result) => {
    const { stats, newBest } = applyTriPeaksResult(prev, result);
    return { stats, update: newBest };
  },
});

/**
 * Hook that persists TriPeaks best-record statistics in localStorage and records
 * finished games. `recordResult` returns whether the game set a new best score so
 * the page can surface a badge.
 */
export const useTriPeaksStats = store.useStats;
