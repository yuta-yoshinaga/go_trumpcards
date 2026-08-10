/** Why raising is unavailable, or `open` when it is still allowed. */
export type RaiseAvailability = 'open' | 'cap' | 'chips';

/** The inputs the backend's `canRaise` is computed from. */
export interface RaiseAvailabilityInput {
  /** Raises already made this round. */
  raiseCount: number;
  /** Raises allowed per round. */
  maxRaises: number;
  /** The human's remaining chips. */
  chips: number;
  /** The current bet to match. */
  currentBet: number;
  /** What the human has already put in this round. */
  roundBet: number;
  /** The ante, which a raise adds on top of the call. */
  ante: number;
}

/**
 * Why the raise button is missing, mirroring the backend's
 *
 * ```go
 * canRaise := raiseCount < MaxRaises && chips >= need+ante
 * ```
 *
 * in `Primero.go` / `Bouillotte.go`. **Both games hide the button rather than
 * disabling it**, so without a reason the player cannot tell a spent raise
 * allowance from an empty stack (#4924 / #4925).
 *
 * The cap is reported first when both hold: it is the condition that cannot be
 * recovered from within the round.
 * @param input - The round's raise state and the human's chips.
 * @returns `cap` when the per-round limit is spent, `chips` when the stack
 *   cannot cover call + ante, otherwise `open`.
 */
export function raiseAvailability(input: RaiseAvailabilityInput): RaiseAvailability {
  if (input.raiseCount >= input.maxRaises) return 'cap';
  if (input.chips < raiseCost(input)) return 'chips';
  return 'open';
}

/**
 * Chips a raise costs: what is still owed on the current bet, plus the ante.
 * Mirrors `need + ante`, with `need` floored at zero.
 * @param input - The current bet, what the human already staked, and the ante.
 * @returns The chips needed to raise.
 */
export function raiseCost(input: Pick<RaiseAvailabilityInput, 'currentBet' | 'roundBet' | 'ante'>): number {
  return Math.max(0, input.currentBet - input.roundBet) + input.ante;
}
