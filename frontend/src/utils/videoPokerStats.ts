/**
 * Video Poker session statistics: a pure reducer plus localStorage persistence
 * helpers. Stats are aggregated per hand (one RESULT phase = one hand) and kept
 * per variant under the `vp_stats_<gameName>` key so each game tallies
 * independently. All persistence is try/catch-guarded so a full or unavailable
 * localStorage never breaks the game.
 */

/** Aggregated session statistics for one Video Poker variant. */
export interface VideoPokerStats {
  /** Total hands played (RESULT phases reached). */
  hands: number;
  /** Hands that paid out (payout > 0). */
  wins: number;
  /** Total coins wagered across all hands. */
  totalBet: number;
  /** Total coins paid out across all hands. */
  totalPayout: number;
  /** Count of each winning hand, keyed by paytable row key. */
  handCounts: Record<string, number>;
}

/** Outcome of a single hand, extracted from the RESULT-phase response. */
export interface VideoPokerResult {
  /** Coins wagered on this hand. */
  bet: number;
  /** Coins paid out (0 on a loss). */
  payout: number;
  /** Paytable row key of the winning hand, or null when the hand paid nothing. */
  rowKey: string | null;
}

/** Returns a fresh, empty stats object (with its own `handCounts` map). */
export function emptyVideoPokerStats(): VideoPokerStats {
  return { hands: 0, wins: 0, totalBet: 0, totalPayout: 0, handCounts: {} };
}

/** Pure reducer: folds one hand result into the running stats, returning a new object. */
export function recordVideoPokerResult(stats: VideoPokerStats, result: VideoPokerResult): VideoPokerStats {
  const won = result.payout > 0;
  const handCounts = { ...stats.handCounts };
  if (won && result.rowKey) {
    handCounts[result.rowKey] = (handCounts[result.rowKey] ?? 0) + 1;
  }
  return {
    hands: stats.hands + 1,
    wins: stats.wins + (won ? 1 : 0),
    totalBet: stats.totalBet + result.bet,
    totalPayout: stats.totalPayout + result.payout,
    handCounts,
  };
}

/** Net coins won or lost this session (payouts minus wagers). */
export function videoPokerNet(stats: VideoPokerStats): number {
  return stats.totalPayout - stats.totalBet;
}

/** Win rate as a 0..1 fraction (0 when no hands played). */
export function videoPokerWinRate(stats: VideoPokerStats): number {
  return stats.hands > 0 ? stats.wins / stats.hands : 0;
}

/** localStorage key for a variant's stats. */
export function videoPokerStatsKey(gameName: string): string {
  return `vp_stats_${gameName}`;
}

/** Coerces a possibly-malformed parsed value into a valid VideoPokerStats. */
function normalizeStats(value: unknown): VideoPokerStats {
  const base = emptyVideoPokerStats();
  if (typeof value !== 'object' || value === null) return base;
  const raw = value as Record<string, unknown>;
  const num = (v: unknown): number => (typeof v === 'number' && Number.isFinite(v) ? v : 0);
  const handCounts: Record<string, number> = {};
  if (typeof raw.handCounts === 'object' && raw.handCounts !== null) {
    for (const [key, v] of Object.entries(raw.handCounts as Record<string, unknown>)) {
      const n = num(v);
      if (n > 0) handCounts[key] = n;
    }
  }
  return {
    hands: num(raw.hands),
    wins: num(raw.wins),
    totalBet: num(raw.totalBet),
    totalPayout: num(raw.totalPayout),
    handCounts,
  };
}

/** Reads a variant's stats from localStorage, returning empty stats if absent, invalid, or unavailable. */
export function readVideoPokerStats(gameName: string): VideoPokerStats {
  try {
    const raw = localStorage.getItem(videoPokerStatsKey(gameName));
    if (raw === null) return emptyVideoPokerStats();
    return normalizeStats(JSON.parse(raw));
  } catch {
    return emptyVideoPokerStats();
  }
}

/** Persists a variant's stats to localStorage; silently ignores quota/availability errors. */
export function writeVideoPokerStats(gameName: string, stats: VideoPokerStats): void {
  try {
    localStorage.setItem(videoPokerStatsKey(gameName), JSON.stringify(stats));
  } catch {
    // localStorage may be full or unavailable (private mode); stats are best-effort.
  }
}
