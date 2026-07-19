/** Roadmap side code for a Player win. Mirrors `ROAD_PLAYER` in BaccaratPage. */
export const ROAD_PLAYER = 0;
/** Roadmap side code for a Banker win. Mirrors `ROAD_BANKER` in BaccaratPage. */
export const ROAD_BANKER = 1;

/**
 * Aggregate statistics for a baccarat shoe history. Counts and rates cover
 * Player / Banker / Tie outcomes; `streakSide` / `streakCount` describe the
 * trailing run of the same non-tie side (ties never break a streak).
 */
export interface BaccaratShoeStats {
  /** Number of Player wins in the history. */
  playerCount: number;
  /** Number of Banker wins in the history. */
  bankerCount: number;
  /** Number of Ties in the history. */
  tieCount: number;
  /** Total number of resolved rounds (Player + Banker + Tie). */
  total: number;
  /** Player appearance rate as an integer percentage (0-100). */
  playerPct: number;
  /** Banker appearance rate as an integer percentage (0-100). */
  bankerPct: number;
  /** Tie appearance rate as an integer percentage (0-100). */
  tiePct: number;
  /**
   * Side of the current trailing streak: `ROAD_PLAYER`, `ROAD_BANKER`, or
   * `null` when there is no non-tie outcome yet.
   */
  streakSide: number | null;
  /** Length of the current trailing streak (ties do not break it). */
  streakCount: number;
}

/**
 * Computes Player/Banker/Tie counts, integer appearance rates, and the current
 * side streak from a baccarat shoe `history`. Each element is a side code:
 * `ROAD_PLAYER`, `ROAD_BANKER`, or any other value (treated as a Tie). A Tie is
 * counted but does not break the trailing streak. Percentages are rounded to
 * the nearest integer and computed against the total resolved-round count.
 */
export function computeBaccaratShoeStats(history: readonly number[]): BaccaratShoeStats {
  let playerCount = 0;
  let bankerCount = 0;
  let tieCount = 0;
  for (const r of history) {
    if (r === ROAD_PLAYER) playerCount += 1;
    else if (r === ROAD_BANKER) bankerCount += 1;
    else tieCount += 1;
  }
  const total = playerCount + bankerCount + tieCount;
  const pct = (n: number) => (total === 0 ? 0 : Math.round((n / total) * 100));

  let streakSide: number | null = null;
  let streakCount = 0;
  for (let i = history.length - 1; i >= 0; i -= 1) {
    const r = history[i];
    if (r !== ROAD_PLAYER && r !== ROAD_BANKER) continue;
    if (streakSide === null) {
      streakSide = r;
      streakCount = 1;
    } else if (r === streakSide) {
      streakCount += 1;
    } else {
      break;
    }
  }

  return {
    playerCount,
    bankerCount,
    tieCount,
    total,
    playerPct: pct(playerCount),
    bankerPct: pct(bankerCount),
    tiePct: pct(tieCount),
    streakSide,
    streakCount,
  };
}
