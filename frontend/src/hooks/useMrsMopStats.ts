import { useCallback } from 'react';
import { createLocalStorageStats } from './createLocalStorageStats';

/** localStorage key for per-difficulty MrsMop Solitaire play statistics. */
export const MRSMOP_STATS_KEY = 'trumpcards-mrsMop-stats';

/** Aggregated play statistics for a single MrsMop difficulty. */
export interface MrsMopDifficultyStat {
  /** Total finished games (wins + losses). */
  plays: number;
  /** Games cleared. */
  wins: number;
  /** Highest score across cleared games, or null if never won. */
  bestScore: number | null;
  /** Fewest moves across cleared games, or null if never won. */
  fewestMoves: number | null;
}

/** Outcome of a single finished MrsMop game. */
export interface MrsMopResult {
  difficulty: number;
  won: boolean;
  score: number;
  moves: number;
}

/** Map of difficulty (as string key) to its aggregated stats. */
export type MrsMopStats = Record<string, MrsMopDifficultyStat>;

/** Returns a zeroed stat record. */
export function emptyMrsMopStat(): MrsMopDifficultyStat {
  return { plays: 0, wins: 0, bestScore: null, fewestMoves: null };
}

function isValidStat(value: unknown): value is MrsMopDifficultyStat {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  const numOrNull = (v: unknown) => v === null || typeof v === 'number';
  return (
    typeof s.plays === 'number' && typeof s.wins === 'number' && numOrNull(s.bestScore) && numOrNull(s.fewestMoves)
  );
}

/** Reads and validates the stats map from localStorage; returns {} on any error. */
export function readMrsMopStats(): MrsMopStats {
  try {
    const raw = localStorage.getItem(MRSMOP_STATS_KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== 'object' || parsed === null) return {};
    const out: MrsMopStats = {};
    for (const [key, val] of Object.entries(parsed as Record<string, unknown>)) {
      if (isValidStat(val)) out[key] = val;
    }
    return out;
  } catch {
    return {};
  }
}

/** Whether a recorded result set a new personal best (for celebration feedback). */
export interface MrsMopBestUpdate {
  newBestScore: boolean;
  newFewestMoves: boolean;
}

/**
 * Pure reducer: folds a finished-game result into the stats map and reports whether
 * it set a new personal best. Best score / fewest moves only advance on a win.
 */
export function applyMrsMopResult(
  stats: MrsMopStats,
  result: MrsMopResult,
): { stats: MrsMopStats; update: MrsMopBestUpdate } {
  const key = String(result.difficulty);
  const prev = stats[key] ?? emptyMrsMopStat();
  const next: MrsMopDifficultyStat = {
    plays: prev.plays + 1,
    wins: prev.wins + (result.won ? 1 : 0),
    bestScore: prev.bestScore,
    fewestMoves: prev.fewestMoves,
  };
  const update: MrsMopBestUpdate = { newBestScore: false, newFewestMoves: false };
  if (result.won) {
    if (prev.bestScore === null || result.score > prev.bestScore) {
      next.bestScore = result.score;
      update.newBestScore = true;
    }
    if (prev.fewestMoves === null || result.moves < prev.fewestMoves) {
      next.fewestMoves = result.moves;
      update.newFewestMoves = true;
    }
  }
  return { stats: { ...stats, [key]: next }, update };
}

/** Win rate as a whole-number percentage (0 when no games played). */
export function mrsMopWinRate(stat: MrsMopDifficultyStat): number {
  if (stat.plays === 0) return 0;
  return Math.round((stat.wins / stat.plays) * 100);
}

/**
 * Hook that persists per-difficulty MrsMop statistics in localStorage and records
 * finished games. `recordResult` returns which personal bests were beaten so the
 * page can surface a badge.
 */
const store = createLocalStorageStats<MrsMopStats, MrsMopResult, MrsMopBestUpdate>({
  key: MRSMOP_STATS_KEY,
  read: readMrsMopStats,
  reduce: applyMrsMopResult,
});

export function useMrsMopStats() {
  const { stats, recordResult } = store.useStats();

  const getStat = useCallback(
    (difficulty: number): MrsMopDifficultyStat => stats[String(difficulty)] ?? emptyMrsMopStat(),
    [stats],
  );

  return { stats, getStat, recordResult };
}
