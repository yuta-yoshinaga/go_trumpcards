import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Stalactites play statistics. */
export const STALACTITE_STATS_KEY = 'trumpcards-stalactite-stats';

/** Aggregated statistics for Stalactites games. */
export interface StalactiteStats {
  plays: number;
  wins: number;
  fewestMoves: number | null;
}

/** Outcome of one finished Stalactites game. */
export interface StalactiteResult {
  won: boolean;
  moves: number;
}

/** Returns an empty Stalactites statistics record. */
export function emptyStalactiteStats(): StalactiteStats {
  return { plays: 0, wins: 0, fewestMoves: null };
}

/** Returns the whole-number Stalactites win rate. */
export function stalactiteWinRate(stats: StalactiteStats): number {
  if (stats.plays === 0) return 0;
  return Math.round((stats.wins / stats.plays) * 100);
}

function isValidStats(value: unknown): value is StalactiteStats {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  return (
    typeof s.plays === 'number' &&
    typeof s.wins === 'number' &&
    (s.fewestMoves === null || typeof s.fewestMoves === 'number')
  );
}

/** Adds one finished Stalactites game and reports whether it set a new move record. */
export function applyStalactiteResult(
  stats: StalactiteStats,
  result: StalactiteResult,
): { stats: StalactiteStats; newBest: boolean } {
  const next: StalactiteStats = {
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

/** Reads persisted Stalactites statistics, falling back to an empty record. */
export const readStalactiteStats = createFlatStatsReader(STALACTITE_STATS_KEY, emptyStalactiteStats, isValidStats);

const store = createLocalStorageStats<StalactiteStats, StalactiteResult, boolean>({
  key: STALACTITE_STATS_KEY,
  read: readStalactiteStats,
  reduce: (prev, result) => {
    const { stats, newBest } = applyStalactiteResult(prev, result);
    return { stats, update: newBest };
  },
});

/** Provides persisted Stalactites statistics and a one-game result recorder. */
export const useStalactiteStats = store.useStats;
