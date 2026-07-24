import { useCallback, useState } from 'react';

/** localStorage key for per-variant Klondike play statistics. */
export const KLONDIKE_STATS_KEY = 'klondike_stats';

/** Aggregated play statistics for a single Klondike variant (drawCount × scoringMode). */
export interface KlondikeVariantStat {
  /** Total finished games (wins + losses). */
  plays: number;
  /** Games cleared. */
  wins: number;
  /** Fastest clear time in seconds, or null if never won. */
  bestTimeSeconds: number | null;
  /** Fewest moves across cleared games, or null if never won. */
  fewestMoves: number | null;
}

/** Outcome of a single finished Klondike game. */
export interface KlondikeResult {
  drawCount: number;
  scoringMode: number;
  won: boolean;
  timeSeconds: number;
  moves: number;
}

/** Map of variant key (`drawCount:scoringMode`) to its aggregated stats. */
export type KlondikeStats = Record<string, KlondikeVariantStat>;

/** Builds the variant key from a draw count and scoring mode. */
export function klondikeVariantKey(drawCount: number, scoringMode: number): string {
  return `${drawCount}:${scoringMode}`;
}

/** Returns a zeroed stat record. */
export function emptyKlondikeStat(): KlondikeVariantStat {
  return { plays: 0, wins: 0, bestTimeSeconds: null, fewestMoves: null };
}

function isValidStat(value: unknown): value is KlondikeVariantStat {
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
export function readKlondikeStats(): KlondikeStats {
  try {
    const raw = localStorage.getItem(KLONDIKE_STATS_KEY);
    if (!raw) return {};
    const parsed: unknown = JSON.parse(raw);
    if (typeof parsed !== 'object' || parsed === null) return {};
    const out: KlondikeStats = {};
    for (const [key, val] of Object.entries(parsed as Record<string, unknown>)) {
      if (isValidStat(val)) out[key] = val;
    }
    return out;
  } catch {
    return {};
  }
}

/** Whether a recorded result set a new personal best (for celebration feedback). */
export interface KlondikeBestUpdate {
  newBestTime: boolean;
  newFewestMoves: boolean;
}

/**
 * Pure reducer: folds a finished-game result into the stats map and reports whether
 * it set a new personal best. Best time / fewest moves only advance on a win.
 */
export function applyKlondikeResult(
  stats: KlondikeStats,
  result: KlondikeResult,
): { stats: KlondikeStats; update: KlondikeBestUpdate } {
  const key = klondikeVariantKey(result.drawCount, result.scoringMode);
  const prev = stats[key] ?? emptyKlondikeStat();
  const next: KlondikeVariantStat = {
    plays: prev.plays + 1,
    wins: prev.wins + (result.won ? 1 : 0),
    bestTimeSeconds: prev.bestTimeSeconds,
    fewestMoves: prev.fewestMoves,
  };
  const update: KlondikeBestUpdate = { newBestTime: false, newFewestMoves: false };
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
export function klondikeWinRate(stat: KlondikeVariantStat): number {
  if (stat.plays === 0) return 0;
  return Math.round((stat.wins / stat.plays) * 100);
}

/**
 * Hook that persists per-variant Klondike statistics in localStorage and records
 * finished games. `recordResult` returns which personal bests were beaten so the
 * page can surface a badge.
 */
export function useKlondikeStats() {
  const [stats, setStats] = useState<KlondikeStats>(readKlondikeStats);

  const recordResult = useCallback((result: KlondikeResult): KlondikeBestUpdate => {
    const { stats: next, update } = applyKlondikeResult(readKlondikeStats(), result);
    setStats(next);
    try {
      localStorage.setItem(KLONDIKE_STATS_KEY, JSON.stringify(next));
    } catch {
      /* storage unavailable / quota exceeded */
    }
    return update;
  }, []);

  const getStat = useCallback(
    (drawCount: number, scoringMode: number): KlondikeVariantStat =>
      stats[klondikeVariantKey(drawCount, scoringMode)] ?? emptyKlondikeStat(),
    [stats],
  );

  return { stats, getStat, recordResult };
}
