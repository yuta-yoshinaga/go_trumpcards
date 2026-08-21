import { useCallback } from 'react';
import { createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for per-variant Whitehead play statistics. */
export const WHITEHEAD_STATS_KEY = 'whitehead_stats';

/** Aggregated play statistics for a single Whitehead variant (drawCount × scoringMode). */
export interface WhiteheadVariantStat {
  /** Total finished games (wins + losses). */
  plays: number;
  /** Games cleared. */
  wins: number;
  /** Fastest clear time in seconds, or null if never won. */
  bestTimeSeconds: number | null;
  /** Fewest moves across cleared games, or null if never won. */
  fewestMoves: number | null;
}

/** Outcome of a single finished Whitehead game. */
export interface WhiteheadResult {
  drawCount: number;
  scoringMode: number;
  won: boolean;
  timeSeconds: number;
  moves: number;
}

/** Map of variant key (`drawCount:scoringMode`) to its aggregated stats. */
export type WhiteheadStats = Record<string, WhiteheadVariantStat>;

/** Builds the variant key from a draw count and scoring mode. */
export function whiteheadVariantKey(drawCount: number, scoringMode: number): string {
  return `${drawCount}:${scoringMode}`;
}

/** Returns a zeroed stat record. */
export function emptyWhiteheadStat(): WhiteheadVariantStat {
  return { plays: 0, wins: 0, bestTimeSeconds: null, fewestMoves: null };
}

function isValidStat(value: unknown): value is WhiteheadVariantStat {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  const numOrNull = (v: unknown) => v === null || typeof v === 'number';
  return (
    typeof s.plays === 'number' &&
    typeof s.wins === 'number' &&
    numOrNull(s.bestTimeSeconds) &&
    numOrNull(s.fewestMoves)
  );
}

/** Reads and validates the stats map from localStorage; returns {} on any error. */
export function readWhiteheadStats(): WhiteheadStats {
  try {
    const raw = localStorage.getItem(WHITEHEAD_STATS_KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== 'object' || parsed === null) return {};
    const out: WhiteheadStats = {};
    for (const [key, val] of Object.entries(parsed as Record<string, unknown>)) {
      if (isValidStat(val)) out[key] = val;
    }
    return out;
  } catch {
    return {};
  }
}

/** Whether a recorded result set a new personal best (for celebration feedback). */
export interface WhiteheadBestUpdate {
  newBestTime: boolean;
  newFewestMoves: boolean;
}

/**
 * Pure reducer: folds a finished-game result into the stats map and reports whether
 * it set a new personal best. Best time / fewest moves only advance on a win.
 */
export function applyWhiteheadResult(
  stats: WhiteheadStats,
  result: WhiteheadResult,
): { stats: WhiteheadStats; update: WhiteheadBestUpdate } {
  const key = whiteheadVariantKey(result.drawCount, result.scoringMode);
  const prev = stats[key] ?? emptyWhiteheadStat();
  const next: WhiteheadVariantStat = {
    plays: prev.plays + 1,
    wins: prev.wins + (result.won ? 1 : 0),
    bestTimeSeconds: prev.bestTimeSeconds,
    fewestMoves: prev.fewestMoves,
  };
  const update: WhiteheadBestUpdate = { newBestTime: false, newFewestMoves: false };
  if (result.won) {
    if (result.timeSeconds > 0 && (prev.bestTimeSeconds === null || result.timeSeconds < prev.bestTimeSeconds)) {
      next.bestTimeSeconds = result.timeSeconds;
      update.newBestTime = true;
    }
    if (prev.fewestMoves === null || result.moves < prev.fewestMoves) {
      next.fewestMoves = result.moves;
      update.newFewestMoves = true;
    }
  }
  return { stats: { ...stats, [key]: next }, update };
}

/** Win rate as a whole-number percentage (0 when no games played). */
export function whiteheadWinRate(stat: WhiteheadVariantStat): number {
  if (stat.plays === 0) return 0;
  return Math.round((stat.wins / stat.plays) * 100);
}

/**
 * Hook that persists per-variant Whitehead statistics in localStorage and records
 * finished games. `recordResult` returns which personal bests were beaten so the
 * page can surface a badge.
 */
const store = createLocalStorageStats<WhiteheadStats, WhiteheadResult, WhiteheadBestUpdate>({
  key: WHITEHEAD_STATS_KEY,
  read: readWhiteheadStats,
  reduce: applyWhiteheadResult,
});

export function useWhiteheadStats() {
  const { stats, recordResult } = store.useStats();

  const getStat = useCallback(
    (drawCount: number, scoringMode: number): WhiteheadVariantStat =>
      stats[whiteheadVariantKey(drawCount, scoringMode)] ?? emptyWhiteheadStat(),
    [stats],
  );

  return { stats, getStat, recordResult };
}
