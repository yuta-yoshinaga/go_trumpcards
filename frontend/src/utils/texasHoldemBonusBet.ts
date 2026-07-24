/**
 * Ante multiplier for the pre-flop Play bet. Mirrors `TexasHoldemBonus.Play`
 * in `internal/domain/TexasHoldemBonus.go`, which places `anteBet * 2`.
 */
export const TEXASHOLDEMBONUS_FLOP_MULTIPLIER = 2;

/**
 * Ante multiplier for a flop/turn Raise bet. Mirrors `TexasHoldemBonus.Raise`
 * in `internal/domain/TexasHoldemBonus.go`, which places `anteBet` (1×) on both
 * the flop and turn streets.
 */
export const TEXASHOLDEMBONUS_RAISE_MULTIPLIER = 1;

/**
 * Computes the chips a Texas Hold'em Bonus street bet costs, so the player can
 * see the cost before committing. The domain's Play bet is `ante * 2` and each
 * Raise is `ante * 1`; this helper multiplies the ante by the given street
 * multiplier. A negative ante is treated as 0.
 *
 * @param anteBet - The ante placed during the bet phase.
 * @param multiplier - The street multiplier (`2` for Play, `1` for Raise).
 * @returns The chips the street bet will cost (never below 0).
 */
export function texasHoldemBonusBetCost(anteBet: number, multiplier: number): number {
  return Math.max(0, anteBet) * multiplier;
}
