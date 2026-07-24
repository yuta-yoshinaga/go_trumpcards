import type { BridgeBidEntry } from '../types/card';

/** Number of seats (players) in a Bridge auction. */
export const BRIDGE_SEAT_COUNT = 4;

/** A Bridge auction shaped into the traditional one-column-per-seat grid. */
export interface BridgeAuctionGrid {
  /**
   * Player index for each column, left-to-right. Column 0 is the opening
   * bidder's seat, and subsequent columns follow the bidding order.
   */
  columns: number[];
  /**
   * Auction rows (bidding rounds). Each row has {@link BRIDGE_SEAT_COUNT}
   * cells; a cell is the bid made by that column's seat, or `null` when the
   * seat has not yet bid in the (final, in-progress) round.
   */
  rows: (BridgeBidEntry | null)[][];
}

/**
 * Shapes a flat Bridge bid history into a 4-column auction grid aligned to the
 * opening bidder's seat. Because bidding always advances one seat at a time,
 * each column maps to a fixed seat and rows fill left-to-right in bid order.
 * Returns `null` for an empty auction so callers can hide the table.
 */
export function buildBridgeAuctionGrid(bidHistory: BridgeBidEntry[]): BridgeAuctionGrid | null {
  if (!bidHistory || bidHistory.length === 0) return null;
  const firstSeat = bidHistory[0].playerIdx;
  const columns = Array.from({ length: BRIDGE_SEAT_COUNT }, (_, i) => (firstSeat + i) % BRIDGE_SEAT_COUNT);
  const rows: (BridgeBidEntry | null)[][] = [];
  for (const entry of bidHistory) {
    const col = (entry.playerIdx - firstSeat + BRIDGE_SEAT_COUNT) % BRIDGE_SEAT_COUNT;
    if (col === 0) rows.push(Array<BridgeBidEntry | null>(BRIDGE_SEAT_COUNT).fill(null));
    rows[rows.length - 1][col] = entry;
  }
  return { columns, rows };
}

/**
 * Returns the final contract bid (the last level bid, `bidType === 1`) in the
 * auction, or `null` when the auction contains no contract bid. The returned
 * reference can be compared by identity to highlight the winning cell.
 */
export function finalContractBid(bidHistory: BridgeBidEntry[]): BridgeBidEntry | null {
  for (let i = bidHistory.length - 1; i >= 0; i--) {
    if (bidHistory[i].bidType === 1) return bidHistory[i];
  }
  return null;
}
