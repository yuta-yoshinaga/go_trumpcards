/** Minimum BlackJack table bet (matches the bet input's min/step). */
export const BJ_MIN_BET = 10;

/** Quick-bet shortcut kinds offered in the bet phase. */
export type BjQuickBetKind = 'min' | 'half' | 'max';

/**
 * Additive quick-chip denominations offered in the bet phase. Each button adds
 * its value to the current bet (clamped to the player's chip balance) so the
 * player can build a bet quickly instead of typing or stepping the input.
 */
export const BJ_CHIP_VALUES = [10, 25, 100] as const;

/**
 * Adds a chip value to the current bet, clamped to the player's chip balance.
 * Never returns more than `playerChips`.
 *
 * @param current - The current bet amount.
 * @param delta - The chip value to add.
 * @param playerChips - The player's current chip balance (the upper bound).
 * @returns The new bet amount, clamped to `playerChips`.
 */
export function bjAddChip(current: number, delta: number, playerChips: number): number {
  return Math.min(playerChips, current + delta);
}

/** Round a chip amount down to the nearest valid (10-multiple) bet. */
function floorToStep(amount: number): number {
  return Math.floor(Math.max(0, amount) / BJ_MIN_BET) * BJ_MIN_BET;
}

/**
 * Computes the bet amount for a quick-bet shortcut from the player's chips.
 * `half` and `max` are rounded down to a 10-multiple and never exceed `playerChips`.
 *
 * @param kind - Which shortcut: table minimum, half the chips, or all chips.
 * @param playerChips - The player's current chip balance.
 * @returns The bet amount to apply.
 */
export function bjQuickBetAmount(kind: BjQuickBetKind, playerChips: number): number {
  switch (kind) {
    case 'min':
      return Math.min(playerChips, BJ_MIN_BET);
    case 'half':
      return Math.min(playerChips, Math.max(BJ_MIN_BET, floorToStep(playerChips / 2)));
    case 'max':
      return Math.min(playerChips, Math.max(BJ_MIN_BET, floorToStep(playerChips)));
  }
}
