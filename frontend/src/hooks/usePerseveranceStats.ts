import { useCallback } from 'react';
import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Perseverance session statistics. */
export const PERSEVERANCE_STATS_KEY = 'trumpcards-perseverance-session-stats';

/** Aggregated clear statistics for Perseverance. */
export interface PerseveranceStats {
  starts: number;
  games: number;
  moves: number;
  redeals: number;
  /** Lowest move count among cleared games, or null before a recorded clear. */
  minMoves: number | null;
  /** Guards against recording the same cleared game more than once. */
  clearRecorded: boolean;
  /** Guards against counting repeated playing renders as new games. */
  gameStarted: boolean;
}

type PerseveranceStatsAction =
  | { type: 'clear'; moveCount: number; redealsLeft: number }
  | { type: 'playing' }
  | { type: 'reset' };

/** Returns empty Perseverance statistics. */
export function emptyPerseveranceStats(): PerseveranceStats {
  return { starts: 0, games: 0, moves: 0, redeals: 0, minMoves: null, clearRecorded: false, gameStarted: false };
}

/** Returns the whole-number clear rate, or zero before the first game starts. */
export function perseveranceClearRate(stats: PerseveranceStats): number {
  return stats.starts === 0 ? 0 : Math.round((stats.games / stats.starts) * 100);
}

function isValidStats(value: unknown): value is PerseveranceStats {
  if (typeof value !== 'object' || value === null) return false;
  const stats = value as Record<string, unknown>;
  return (
    typeof stats.games === 'number' &&
    typeof stats.moves === 'number' &&
    typeof stats.redeals === 'number' &&
    (stats.minMoves === undefined || stats.minMoves === null || typeof stats.minMoves === 'number') &&
    typeof stats.clearRecorded === 'boolean' &&
    (stats.starts === undefined || typeof stats.starts === 'number') &&
    (stats.gameStarted === undefined || typeof stats.gameStarted === 'boolean')
  );
}

/**
 * Reads persisted Perseverance statistics, migrating records from before starts
 * were tracked. Since old records did not count starts, their clear rate begins
 * at 100% by using the clear count as the starting count.
 */
const readValidPerseveranceStats = createFlatStatsReader(PERSEVERANCE_STATS_KEY, emptyPerseveranceStats, isValidStats);

/** Reads persisted Perseverance statistics, migrating older records as needed. */
export function readPerseveranceStats(): PerseveranceStats {
  const stats = readValidPerseveranceStats();
  return {
    ...stats,
    starts: stats.starts ?? stats.games,
    gameStarted: stats.gameStarted ?? false,
    minMoves: stats.minMoves ?? null,
  };
}

/** Folds a clear, new-playing-game marker, or stats reset into Perseverance statistics. */
export function applyPerseveranceStatsAction(
  prev: PerseveranceStats,
  action: PerseveranceStatsAction,
): { stats: PerseveranceStats; update: undefined } {
  if (action.type === 'reset') {
    return {
      stats: { ...emptyPerseveranceStats(), clearRecorded: prev.clearRecorded },
      update: undefined,
    };
  }
  if (action.type === 'playing') {
    if (prev.gameStarted) return { stats: prev, update: undefined };
    return {
      stats: { ...prev, starts: prev.starts + 1, clearRecorded: false, gameStarted: true },
      update: undefined,
    };
  }
  if (prev.clearRecorded) return { stats: prev, update: undefined };
  return {
    stats: {
      starts: prev.starts + (prev.gameStarted ? 0 : 1),
      games: prev.games + 1,
      moves: prev.moves + action.moveCount,
      redeals: prev.redeals + action.redealsLeft,
      minMoves: prev.minMoves == null ? action.moveCount : Math.min(prev.minMoves, action.moveCount),
      clearRecorded: true,
      gameStarted: false,
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
