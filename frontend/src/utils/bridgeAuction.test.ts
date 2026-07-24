import { describe, expect, it } from 'vitest';
import type { BridgeBidEntry } from '../types/card';
import { BRIDGE_SEAT_COUNT, buildBridgeAuctionGrid, finalContractBid } from './bridgeAuction';

const bid = (playerIdx: number, bidType: number, level = 0, suit = 0): BridgeBidEntry => ({
  playerIdx,
  bidType,
  level,
  suit,
});

describe('buildBridgeAuctionGrid', () => {
  it('returns null for an empty auction', () => {
    expect(buildBridgeAuctionGrid([])).toBeNull();
  });

  it('aligns column 0 to the opening bidder and orders seats by bid order', () => {
    // Opener is seat 2, so columns cycle 2 -> 3 -> 0 -> 1.
    const grid = buildBridgeAuctionGrid([bid(2, 1, 1, 5)]);
    expect(grid?.columns).toEqual([2, 3, 0, 1]);
  });

  it('places a full round of bids into the matching seat cells', () => {
    const history = [bid(0, 1, 1, 5), bid(1, 0), bid(2, 2), bid(3, 0)];
    const grid = buildBridgeAuctionGrid(history);
    expect(grid?.columns).toEqual([0, 1, 2, 3]);
    expect(grid?.rows).toHaveLength(1);
    expect(grid?.rows[0]).toEqual(history);
  });

  it('wraps into a new row after every four bids and pads the trailing round with nulls', () => {
    const history = [bid(0, 1, 1, 5), bid(1, 0), bid(2, 0), bid(3, 0), bid(0, 1, 2, 4)];
    const grid = buildBridgeAuctionGrid(history);
    expect(grid?.rows).toHaveLength(2);
    expect(grid?.rows[0]).toEqual(history.slice(0, 4));
    // Second round: only the opener has bid; remaining three seats are null.
    expect(grid?.rows[1]).toEqual([history[4], null, null, null]);
    expect(grid?.rows[1]).toHaveLength(BRIDGE_SEAT_COUNT);
  });
});

describe('finalContractBid', () => {
  it('returns null when there is no level bid', () => {
    expect(finalContractBid([bid(0, 0), bid(1, 0)])).toBeNull();
  });

  it('returns the last level bid by reference', () => {
    const first = bid(0, 1, 1, 5);
    const last = bid(2, 1, 3, 4);
    const result = finalContractBid([first, bid(1, 0), last, bid(3, 2)]);
    expect(result).toBe(last);
  });
});
