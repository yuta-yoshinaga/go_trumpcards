import { useCallback } from 'react';
import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Shamrocks clear statistics. */
export const SHAMROCKS_STATS_KEY = 'trumpcards-shamrocks-session-stats';

/** Aggregated Shamrocks clear statistics. */
export interface ShamrocksStats {
  games: number;
  moves: number;
  /** Guards against recording the same completed game more than once. */
  clearRecorded: boolean;
}

/** Returns an empty Shamrocks statistics record. */
export function emptyShamrocksStats(): ShamrocksStats {
  return { games: 0, moves: 0, clearRecorded: false };
}

function isValidStats(value: unknown): value is ShamrocksStats {
  if (typeof value !== 'object' || value === null) return false;
  const stats = value as Record<string, unknown>;
  return typeof stats.games === 'number' && typeof stats.moves === 'number' && typeof stats.clearRecorded === 'boolean';
}

/** Reads persisted Shamrocks statistics, falling back to an empty record. */
export const readShamrocksStats = createFlatStatsReader(SHAMROCKS_STATS_KEY, emptyShamrocksStats, isValidStats);

/** Adds one cleared game's move count, unless a clear is already recorded. */
export function applyShamrocksClear(
  stats: ShamrocksStats,
  result: { moves?: number; reset?: boolean; playing?: boolean },
): { stats: ShamrocksStats; update: undefined } {
  if (result.reset)
    return { stats: { ...emptyShamrocksStats(), clearRecorded: stats.clearRecorded }, update: undefined };
  if (result.playing) return { stats: { ...stats, clearRecorded: false }, update: undefined };
  if (stats.clearRecorded) return { stats, update: undefined };
  return {
    stats: { games: stats.games + 1, moves: stats.moves + (result.moves ?? 0), clearRecorded: true },
    update: undefined,
  };
}

const store = createLocalStorageStats<
  ShamrocksStats,
  { moves?: number; reset?: boolean; playing?: boolean },
  undefined
>({
  key: SHAMROCKS_STATS_KEY,
  read: readShamrocksStats,
  reduce: applyShamrocksClear,
});

/** Provides persisted clear statistics, duplicate-safe recording, and reset. */
export function useShamrocksStats() {
  const { stats, recordResult } = store.useStats();
  const recordClear = useCallback((moves: number) => recordResult({ moves }), [recordResult]);
  const markPlaying = useCallback(() => recordResult({ playing: true }), [recordResult]);
  const resetStats = useCallback(() => recordResult({ reset: true }), [recordResult]);
  return { stats, recordClear, markPlaying, resetStats };
}
