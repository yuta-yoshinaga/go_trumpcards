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

/** Minimum Ante accepted by the game. */
export const UTH_ANTE_MIN = 10;
/** Minimum Trips side bet. */
export const UTH_TRIPS_MIN = 0;
/** Ante increment accepted by the game. */
export const UTH_ANTE_STEP = 10;
/** Trips increment accepted by the game. */
export const UTH_TRIPS_STEP = 10;

/** Returns the maximum Ante allowed by the chip balance. */
export function anteMax(chips: number): number {
  return Math.floor(chips / 2);
}

/** Returns the maximum Trips bet after committing the Ante and matching Blind. */
export function maxTrips(anteBet: number, chips: number): number {
  return Math.max(0, chips - anteBet * 2);
}

/** Returns whether a Play multiplier is affordable with the current balance. */
export function canPlay(multiplier: number, chips: number, anteBet: number): boolean {
  return chips >= anteBet * multiplier;
}

/** Translation keys for the page's poker hand rank labels (rank 0 through 9). */
export const UTH_HAND_RANK_KEYS: Readonly<Record<number, string>> = {
  0: 'handRank.0',
  1: 'handRank.1',
  2: 'handRank.2',
  3: 'handRank.3',
  4: 'handRank.4',
  5: 'handRank.5',
  6: 'handRank.6',
  7: 'handRank.7',
  8: 'handRank.8',
  9: 'handRank.9',
};

/** English names corresponding to the Ultimate Texas Hold'em hand rank keys. */
export const UTH_HAND_RANK_NAMES: Readonly<Record<number, string>> = {
  0: 'High Card',
  1: 'One Pair',
  2: 'Two Pair',
  3: 'Three of a Kind',
  4: 'Straight',
  5: 'Flush',
  6: 'Full House',
  7: 'Four of a Kind',
  8: 'Straight Flush',
  9: 'Royal Flush',
};

/**
 * Computes the combined Ante+Blind+Trips bounds against the player's chips.
 *
 * @param ante - The Ante bet; the Blind matches it, so it commits `ante * 2`.
 * @param trips - The optional Trips side bet.
 * @param chips - The player's current chip balance.
 * @returns The committed total, the maximum Trips that still fits, and whether the bet is valid.
 */
export function utHoldemBetBounds(ante: number, trips: number, chips: number): UtHoldemBetBounds {
  const tripsLimit = maxTrips(ante, chips);
  const total = ante * 2 + trips;
  return { total, maxTrips: tripsLimit, valid: total <= chips };
}
