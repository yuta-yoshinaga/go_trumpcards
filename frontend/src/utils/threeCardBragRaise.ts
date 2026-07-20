/**
 * Three Card Brag raise-amount bounds, mirroring the CUI guidance in
 * `ThreeCardBragCuiPresenter.threeCardBragRaiseRangeStr`.
 */
export interface ThreeCardBragRaiseBounds {
  /** Smallest legal raise (must exceed the current stake). */
  min: number;
  /** Largest raise the player can afford. */
  max: number;
  /** Whether any legal raise exists (`max >= min`). */
  canRaise: boolean;
}

/**
 * Computes the legal raise range for the human at a Brag betting turn.
 *
 * The minimum is always `stake + 1` (a raise must exceed the current stake).
 * The maximum is the largest stake the player can afford to call: a Seen
 * player pays double the stake, halving the affordable ceiling, so the cap is
 * `floor(chips / 2)`; a Blind player's cap is their full chip count.
 *
 * @param stake - The current stake.
 * @param chips - The human player's remaining chips.
 * @param seen - Whether the human has looked at their hand (Seen vs Blind).
 * @returns The `{ min, max, canRaise }` bounds.
 */
export function threeCardBragRaiseBounds(stake: number, chips: number, seen: boolean): ThreeCardBragRaiseBounds {
  const min = stake + 1;
  const max = seen ? Math.floor(chips / 2) : chips;
  return { min, max, canRaise: max >= min };
}

/**
 * Computes the actual chips a player pays for a Brag bet or raise at a given
 * nominal stake, mirroring the domain's `callCost` rule
 * (`internal/domain/ThreeCardBrag.go`): a Seen player pays double the nominal
 * amount, a Blind player pays the nominal amount.
 *
 * @param nominal - The nominal stake to call or raise to.
 * @param seen - Whether the player has looked at their hand (Seen vs Blind).
 * @returns The real chips cost (`nominal * 2` when Seen, else `nominal`).
 */
export function threeCardBragActualCost(nominal: number, seen: boolean): number {
  return seen ? nominal * 2 : nominal;
}

/**
 * Clamps a raise amount into the `[min, max]` range. When no legal raise
 * exists (`max < min`), returns `min` so the value never drops below the stake.
 *
 * @param value - The candidate raise amount.
 * @param min - The minimum legal raise.
 * @param max - The maximum affordable raise.
 * @returns The clamped raise amount.
 */
export function clampThreeCardBragRaise(value: number, min: number, max: number): number {
  if (max < min) return min;
  return Math.min(Math.max(value, min), max);
}
