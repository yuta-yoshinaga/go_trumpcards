import { createFlatStatsReader, createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for Curds and Whey play statistics. */
export const CURDS_AND_WHEY_STATS_KEY = 'trumpcards-curdsandwhey-stats';

/** Aggregated Curds and Whey statistics. */
export interface CurdsAndWheyStats {
  plays: number;
  wins: number;
  fewestMoves: number | null;
}

/** Outcome of one finished Curds and Whey game. */
export interface CurdsAndWheyResult {
  won: boolean;
  moves: number;
}

/** Returns the initial Curds and Whey statistics. */
export function emptyCurdsAndWheyStats(): CurdsAndWheyStats {
  return { plays: 0, wins: 0, fewestMoves: null };
}

/** Calculates the rounded win percentage for Curds and Whey statistics. */
export function curdsAndWheyWinRate(stats: CurdsAndWheyStats): number {
  return stats.plays === 0 ? 0 : Math.round((stats.wins / stats.plays) * 100);
}

function isValidStats(value: unknown): value is CurdsAndWheyStats {
  if (typeof value !== 'object' || value === null) return false;
  const stats = value as Record<string, unknown>;
  return (
    typeof stats.plays === 'number' &&
    typeof stats.wins === 'number' &&
    (stats.fewestMoves === null || typeof stats.fewestMoves === 'number')
  );
}

/** Applies a finished game result to the accumulated Curds and Whey statistics. */
export function applyCurdsAndWheyResult(stats: CurdsAndWheyStats, result: CurdsAndWheyResult): CurdsAndWheyStats {
  return {
    plays: stats.plays + 1,
    wins: stats.wins + Number(result.won),
    fewestMoves:
      result.won && result.moves > 0 && (stats.fewestMoves === null || result.moves < stats.fewestMoves)
        ? result.moves
        : stats.fewestMoves,
  };
}

/** Reads persisted Curds and Whey statistics, falling back to an empty record when invalid. */
export const readCurdsAndWheyStats = createFlatStatsReader(
  CURDS_AND_WHEY_STATS_KEY,
  emptyCurdsAndWheyStats,
  isValidStats,
);

const store = createLocalStorageStats<CurdsAndWheyStats, CurdsAndWheyResult, void>({
  key: CURDS_AND_WHEY_STATS_KEY,
  read: readCurdsAndWheyStats,
  reduce: (prev, result) => ({ stats: applyCurdsAndWheyResult(prev, result), update: undefined }),
});

/** Provides persisted Curds and Whey statistics and a function to record game results. */
export const useCurdsAndWheyStats = store.useStats;
