import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

export const STALACTITE_STATS_KEY = 'trumpcards-stalactite-stats';

export interface StalactiteStats {
  plays: number;
  wins: number;
  fewestMoves: number | null;
}

export interface StalactiteResult {
  won: boolean;
  moves: number;
}

export function emptyStalactiteStats(): StalactiteStats {
  return { plays: 0, wins: 0, fewestMoves: null };
}

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

export const readStalactiteStats = createFlatStatsReader(STALACTITE_STATS_KEY, emptyStalactiteStats, isValidStats);

const store = createLocalStorageStats<StalactiteStats, StalactiteResult, boolean>({
  key: STALACTITE_STATS_KEY,
  read: readStalactiteStats,
  reduce: (prev, result) => {
    const { stats, newBest } = applyStalactiteResult(prev, result);
    return { stats, update: newBest };
  },
});

export const useStalactiteStats = store.useStats;
