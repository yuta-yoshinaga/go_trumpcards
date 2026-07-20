/**
 * Combined-bet bounds for Ultimate Texas Hold'em. Setting the Ante also commits
 * an equal Blind, so the chips committed up front are `ante * 2 + trips`. The
 * Ante+Blind and the optional Trips side bet must together fit within the
 * player's chip balance.
 */
export interface UtHoldemBetBounds {
  /** Total chips committed by the bet: `ante * 2` (Ante + Blind) plus Trips. */
  total: number;
  /** Largest Trips side bet that still fits within chips at the current Ante (never below 0). */
  maxTrips: number;
  /** Whether the combined bet fits within the player's chip balance. */
  valid: boolean;
}

/**
 * Computes the combined Ante+Blind+Trips bounds against the player's chips.
 *
 * @param ante - The Ante bet; the Blind matches it, so it commits `ante * 2`.
 * @param trips - The optional Trips side bet.
 * @param chips - The player's current chip balance.
 * @returns The committed total, the maximum Trips that still fits, and whether the bet is valid.
 */
export function utHoldemBetBounds(ante: number, trips: number, chips: number): UtHoldemBetBounds {
  const maxTrips = Math.max(0, chips - ante * 2);
  const total = ante * 2 + trips;
  return { total, maxTrips, valid: total <= chips };
}
