import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Cruel play statistics. */
export const CRUEL_STATS_KEY = 'trumpcards-cruel-stats';

/** Aggregated statistics for Cruel games. */
export interface CruelStats {
  plays: number;
  wins: number;
  fewestMoves: number | null;
}

/** Outcome of one finished Cruel game. */
export interface CruelResult {
  won: boolean;
  moves: number;
}

/** Returns an empty Cruel statistics record. */
export function emptyCruelStats(): CruelStats {
  return { plays: 0, wins: 0, fewestMoves: null };
}

/** Returns the whole-number Cruel win rate. */
export function cruelWinRate(stats: CruelStats): number {
  if (stats.plays === 0) return 0;
  return Math.round((stats.wins / stats.plays) * 100);
}

function isValidStats(value: unknown): value is CruelStats {
  if (typeof value !== 'object' || value === null) return false;
  const stats = value as Record<string, unknown>;
  return (
    typeof stats.plays === 'number' &&
    typeof stats.wins === 'number' &&
    (stats.fewestMoves === null || typeof stats.fewestMoves === 'number')
  );
}

/** Adds one finished Cruel game and reports whether it set a new move record. */
export function applyCruelResult(stats: CruelStats, result: CruelResult): { stats: CruelStats; newBest: boolean } {
  const next: CruelStats = {
    plays: stats.plays + 1,
    wins: stats.wins + (result.won ? 1 : 0),
    fewestMoves: stats.fewestMoves,
  };
  const newBest = result.won && result.moves > 0 && (stats.fewestMoves === null || result.moves < stats.fewestMoves);
  if (newBest) next.fewestMoves = result.moves;
  return { stats: next, newBest };
}

/** Reads the persisted Cruel statistics, falling back to an empty record. */
export const readCruelStats = createFlatStatsReader(CRUEL_STATS_KEY, emptyCruelStats, isValidStats);

const store = createLocalStorageStats<CruelStats, CruelResult, boolean>({
  key: CRUEL_STATS_KEY,
  read: readCruelStats,
  reduce: (stats, result) => {
    const { stats: next, newBest } = applyCruelResult(stats, result);
    return { stats: next, update: newBest };
  },
});

/** Provides persisted Cruel statistics and a one-game result recorder. */
export const useCruelStats = store.useStats;
