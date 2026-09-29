import { useCallback } from 'react';
import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Perseverance session statistics. */
export const PERSEVERANCE_STATS_KEY = 'trumpcards-perseverance-session-stats';

/** Aggregated clear statistics for Perseverance. */
export interface PerseveranceStats {
  games: number;
  moves: number;
  redeals: number;
  /** Guards against recording the same cleared game more than once. */
  clearRecorded: boolean;
}

type PerseveranceStatsAction =
  | { type: 'clear'; moveCount: number; redealsLeft: number }
  | { type: 'playing' }
  | { type: 'reset' };

/** Returns empty Perseverance statistics. */
export function emptyPerseveranceStats(): PerseveranceStats {
  return { games: 0, moves: 0, redeals: 0, clearRecorded: false };
}

function isValidStats(value: unknown): value is PerseveranceStats {
  if (typeof value !== 'object' || value === null) return false;
  const stats = value as Record<string, unknown>;
  return (
    typeof stats.games === 'number' &&
    typeof stats.moves === 'number' &&
    typeof stats.redeals === 'number' &&
    typeof stats.clearRecorded === 'boolean'
  );
}

/** Reads persisted Perseverance statistics, falling back to an empty record. */
export const readPerseveranceStats = createFlatStatsReader(
  PERSEVERANCE_STATS_KEY,
  emptyPerseveranceStats,
  isValidStats,
);

/** Folds a clear, new-playing-game marker, or stats reset into Perseverance statistics. */
export function applyPerseveranceStatsAction(
  prev: PerseveranceStats,
  action: PerseveranceStatsAction,
): { stats: PerseveranceStats; update: undefined } {
  if (action.type === 'reset') {
    return { stats: { ...emptyPerseveranceStats(), clearRecorded: prev.clearRecorded }, update: undefined };
  }
  if (action.type === 'playing') {
    return { stats: prev.clearRecorded ? { ...prev, clearRecorded: false } : prev, update: undefined };
  }
  if (prev.clearRecorded) return { stats: prev, update: undefined };
  return {
    stats: {
      games: prev.games + 1,
      moves: prev.moves + action.moveCount,
      redeals: prev.redeals + action.redealsLeft,
      clearRecorded: true,
    },
    update: undefined,
  };
}

const store = createLocalStorageStats<PerseveranceStats, PerseveranceStatsAction, void>({
  key: PERSEVERANCE_STATS_KEY,
  read: readPerseveranceStats,
  reduce: applyPerseveranceStatsAction,
});

/** Provides persisted Perseverance statistics and clear, playing, and reset actions. */
export function usePerseveranceStats() {
  const { stats, recordResult } = store.useStats();
  const recordClear = useCallback(
    (moveCount: number, redealsLeft: number) => recordResult({ type: 'clear', moveCount, redealsLeft }),
    [recordResult],
  );
  const markPlaying = useCallback(() => recordResult({ type: 'playing' }), [recordResult]);
  const clearStats = useCallback(() => recordResult({ type: 'reset' }), [recordResult]);
  return {
    stats,
    recordClear,
    markPlaying,
    clearStats,
  };
}
