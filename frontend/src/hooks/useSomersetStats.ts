import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Somerset play statistics. */
export const SOMERSET_STATS_KEY = 'trumpcards-somerset-stats';

/** Aggregated statistics for Somerset games. */
export interface SomersetStats {
  plays: number;
  wins: number;
  fewestMoves: number | null;
}

/** Outcome of one finished Somerset game. */
export interface SomersetResult {
  won: boolean;
  moves: number;
}

/** Returns an empty Somerset statistics record. */
export function emptySomersetStats(): SomersetStats {
  return { plays: 0, wins: 0, fewestMoves: null };
}

/** Returns the whole-number Somerset win rate. */
export function somersetWinRate(stats: SomersetStats): number {
  if (stats.plays === 0) return 0;
  return Math.round((stats.wins / stats.plays) * 100);
}

function isValidStats(value: unknown): value is SomersetStats {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  return (
    typeof s.plays === 'number' &&
    typeof s.wins === 'number' &&
    (s.fewestMoves === null || typeof s.fewestMoves === 'number')
  );
}

/** Adds one finished Somerset game and reports whether it set a new move record. */
export function applySomersetResult(
  stats: SomersetStats,
  result: SomersetResult,
): { stats: SomersetStats; newBest: boolean } {
  const next: SomersetStats = {
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

/** Reads persisted Somerset statistics, falling back to an empty record. */
export const readSomersetStats = createFlatStatsReader(SOMERSET_STATS_KEY, emptySomersetStats, isValidStats);

const store = createLocalStorageStats<SomersetStats, SomersetResult, boolean>({
  key: SOMERSET_STATS_KEY,
  read: readSomersetStats,
  reduce: (prev, result) => {
    const { stats, newBest } = applySomersetResult(prev, result);
    return { stats, update: newBest };
  },
});

/** Provides persisted Somerset statistics and a one-game result recorder. */
export const useSomersetStats = store.useStats;
