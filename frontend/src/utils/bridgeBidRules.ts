/**
 * Bridge auction legality helpers.
 *
 * These mirror the server-side rules in `internal/domain/Bridge.go`
 * (`doBidDouble`, `doBidRedouble`, `isHigherBid`) so the UI can disable
 * illegal bid controls before a request is sent.
 */

/** Bid types, matching the Go domain `BridgeBidType` enum. */
export const BRIDGE_BID_PASS = 0;
/** A normal contract bid (level + denomination). */
export const BRIDGE_BID_NORMAL = 1;
/** A double bid. */
export const BRIDGE_BID_DOUBLE = 2;
/** A redouble bid. */
export const BRIDGE_BID_REDOUBLE = 3;

/** Minimal bid-history entry shape needed for legality checks. */
export interface BridgeBidLike {
  playerIdx: number;
  bidType: number;
}

/** Minimal player shape needed for legality checks. */
export interface BridgePlayerLike {
  team: number;
}

/** Minimal auction state needed to derive bid legality. */
export interface BridgeAuctionLike {
  bidHistory: readonly BridgeBidLike[];
  players: readonly BridgePlayerLike[];
  bidPlayerIdx: number;
  doubled: number;
  contractLevel: number;
  contractSuit: number;
}

/**
 * Returns the team that made the last normal (contract) bid, or -1 if no
 * contract bid has been made yet. Mirrors the domain's `lastBidTeam`, which is
 * only updated on normal bids (not passes/doubles/redoubles).
 */
export function lastBidTeam(state: BridgeAuctionLike): number {
  for (let i = state.bidHistory.length - 1; i >= 0; i--) {
    const entry = state.bidHistory[i];
    if (entry.bidType === BRIDGE_BID_NORMAL) {
      return state.players[entry.playerIdx]?.team ?? -1;
    }
  }
  return -1;
}

/**
 * Returns true when the current bidder may Double. Legal only when a contract
 * bid exists, it is not already doubled/redoubled, and the last contract bid
 * was made by the opposing team.
 */
export function canDouble(state: BridgeAuctionLike): boolean {
  if (state.contractLevel <= 0) return false;
  if (state.doubled !== 0) return false;
  const team = state.players[state.bidPlayerIdx]?.team ?? -1;
  const bidTeam = lastBidTeam(state);
  return bidTeam !== -1 && team !== -1 && bidTeam !== team;
}

/**
 * Returns true when the current bidder may Redouble. Legal only when the
 * contract is currently doubled (not redoubled) and the current bidder is on
 * the team whose bid was doubled.
 */
export function canRedouble(state: BridgeAuctionLike): boolean {
  if (state.doubled !== 1) return false;
  const team = state.players[state.bidPlayerIdx]?.team ?? -1;
  return team !== -1 && team === lastBidTeam(state);
}

/**
 * Returns true when a normal bid at the given level/denomination outranks the
 * current contract. Mirrors the domain's `isHigherBid`.
 */
export function canBid(contractLevel: number, contractSuit: number, level: number, suit: number): boolean {
  if (contractLevel <= 0) return true;
  if (level > contractLevel) return true;
  return level === contractLevel && suit > contractSuit;
}
