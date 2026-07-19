import { useCallback, useState } from 'react';

/** localStorage key for per-difficulty Spider Solitaire play statistics. */
export const SPIDER_STATS_KEY = 'trumpcards-spider-stats';

/** Aggregated play statistics for a single Spider difficulty. */
export interface SpiderDifficultyStat {
  /** Total finished games (wins + losses). */
  plays: number;
  /** Games cleared. */
  wins: number;
  /** Highest score across cleared games, or null if never won. */
  bestScore: number | null;
  /** Fewest moves across cleared games, or null if never won. */
  fewestMoves: number | null;
}

/** Outcome of a single finished Spider game. */
export interface SpiderResult {
  difficulty: number;
  won: boolean;
  score: number;
  moves: number;
}

/** Map of difficulty (as string key) to its aggregated stats. */
export type SpiderStats = Record<string, SpiderDifficultyStat>;

/** Returns a zeroed stat record. */
export function emptySpiderStat(): SpiderDifficultyStat {
  return { plays: 0, wins: 0, bestScore: null, fewestMoves: null };
}

function isValidStat(value: unknown): value is SpiderDifficultyStat {
  if (typeof value !== 'object' || value === null) return false;
  const s = value as Record<string, unknown>;
  const numOrNull = (v: unknown) => v === null || typeof v === 'number';
  return (
    typeof s.plays === 'number' && typeof s.wins === 'number' && numOrNull(s.bestScore) && numOrNull(s.fewestMoves)
  );
}

/** Reads and validates the stats map from localStorage; returns {} on any error. */
export function readSpiderStats(): SpiderStats {
  try {
    const raw = localStorage.getItem(SPIDER_STATS_KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== 'object' || parsed === null) return {};
    const out: SpiderStats = {};
    for (const [key, val] of Object.entries(parsed as Record<string, unknown>)) {
      if (isValidStat(val)) out[key] = val;
    }
    return out;
  } catch {
    return {};
  }
}

/** Whether a recorded result set a new personal best (for celebration feedback). */
export interface SpiderBestUpdate {
  newBestScore: boolean;
  newFewestMoves: boolean;
}

/**
 * Pure reducer: folds a finished-game result into the stats map and reports whether
 * it set a new personal best. Best score / fewest moves only advance on a win.
 */
export function applySpiderResult(
  stats: SpiderStats,
  result: SpiderResult,
): { stats: SpiderStats; update: SpiderBestUpdate } {
  const key = String(result.difficulty);
  const prev = stats[key] ?? emptySpiderStat();
  const next: SpiderDifficultyStat = {
    plays: prev.plays + 1,
    wins: prev.wins + (result.won ? 1 : 0),
    bestScore: prev.bestScore,
    fewestMoves: prev.fewestMoves,
  };
  const update: SpiderBestUpdate = { newBestScore: false, newFewestMoves: false };
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
export function spiderWinRate(stat: SpiderDifficultyStat): number {
  if (stat.plays === 0) return 0;
  return Math.round((stat.wins / stat.plays) * 100);
}

/**
 * Hook that persists per-difficulty Spider statistics in localStorage and records
 * finished games. `recordResult` returns which personal bests were beaten so the
 * page can surface a badge.
 */
export function useSpiderStats() {
  const [stats, setStats] = useState<SpiderStats>(readSpiderStats);

  const recordResult = useCallback((result: SpiderResult): SpiderBestUpdate => {
    const { stats: next, update } = applySpiderResult(readSpiderStats(), result);
    setStats(next);
    try {
      localStorage.setItem(SPIDER_STATS_KEY, JSON.stringify(next));
    } catch {
      /* storage unavailable / quota exceeded */
    }
    return update;
  }, []);

  const getStat = useCallback(
    (difficulty: number): SpiderDifficultyStat => stats[String(difficulty)] ?? emptySpiderStat(),
    [stats],
  );

  return { stats, getStat, recordResult };
}
