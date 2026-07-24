/**
 * Pure scoring logic for TriPeaks Solitaire (issue #3087).
 *
 * The backend does not track a score, so the frontend derives one from the chain
 * of consecutive removals (the same signal the combo badge uses): the nth card
 * removed without an intervening stock draw is worth `n × POINTS_PER_CHAIN`, and
 * fully clearing a peak grants a flat bonus. Drawing from the stock (or undoing)
 * breaks the chain but keeps the accumulated score.
 */

/** Points awarded for the nth consecutive removal is `chain × this`. */
export const TRIPEAKS_POINTS_PER_CHAIN = 100;

/** Flat bonus granted each time a peak is fully cleared. */
export const TRIPEAKS_PEAK_BONUS = 500;

/** Running score state for a single TriPeaks game. */
export interface TriPeaksScoreState {
  /** Accumulated score so far this game. */
  score: number;
  /** Length of the current unbroken removal chain (0 when idle or just drew). */
  chain: number;
}

/** Zeroed score state for a fresh game. */
export const initialTriPeaksScore: TriPeaksScoreState = { score: 0, chain: 0 };

/** Board snapshot inputs the score transition depends on. */
export interface TriPeaksSnapshot {
  /** Server move counter (rises by one per removal). */
  moveCount: number;
  /** Cards left in the stock (changes when the player draws). */
  stockCount: number;
  /** Number of peaks (0-3) currently fully cleared. */
  peaksCleared: number;
}

/** Points earned by removing the nth consecutive card (1-indexed chain). */
export function chainRemovalPoints(chain: number): number {
  return chain * TRIPEAKS_POINTS_PER_CHAIN;
}

/**
 * Folds one board transition into the running score. Mirrors `useChainCombo`'s
 * chain rules so the displayed combo and the score stay in sync:
 *
 * - `moveCount === 0` — a fresh game: reset both score and chain to zero.
 * - stock changed (a draw) — chain resets, score is preserved.
 * - `moveCount` decreased (an undo) — chain resets, score is preserved.
 * - `moveCount` increased (a removal) — extend the chain and add
 *   `chain × POINTS_PER_CHAIN`, plus `PEAK_BONUS` for each newly cleared peak.
 * - otherwise — unchanged.
 *
 * `prev` is `null` before the first observed snapshot, which yields no change.
 */
export function applyTriPeaksScore(
  state: TriPeaksScoreState,
  prev: TriPeaksSnapshot | null,
  cur: TriPeaksSnapshot,
): TriPeaksScoreState {
  if (cur.moveCount === 0) return { ...initialTriPeaksScore };
  if (prev === null) return state;
  if (cur.stockCount !== prev.stockCount) return { score: state.score, chain: 0 };
  if (cur.moveCount < prev.moveCount) return { score: state.score, chain: 0 };
  if (cur.moveCount > prev.moveCount) {
    const chain = state.chain + 1;
    const clearedNow = Math.max(0, cur.peaksCleared - prev.peaksCleared);
    const gain = chainRemovalPoints(chain) + clearedNow * TRIPEAKS_PEAK_BONUS;
    return { score: state.score + gain, chain };
  }
  return state;
}
