import type { RookPlayerData } from '../types/card';

/** A lightweight identifier for a player who has dropped out of the bidding. */
export interface RookPassedPlayer {
  id: number;
  isHuman: boolean;
}

/** Bidding status derived from the Rook players, for the bid-turn readout. */
export interface RookBidStatus {
  /** Number of players still in the bidding (have not passed). */
  activeBidders: number;
  /** Players who have passed (dropped out of the bidding), in seat order. */
  passed: RookPassedPlayer[];
}

/**
 * Derive the current bidding status from the Rook players: how many are still
 * in the auction and which ones have passed. Pure — formatting/i18n is left to
 * the caller.
 */
export function rookBidStatus(players: RookPlayerData[]): RookBidStatus {
  const passed = players.filter((p) => p.passed).map((p) => ({ id: p.id, isHuman: p.isHuman }));
  return { activeBidders: players.length - passed.length, passed };
}
