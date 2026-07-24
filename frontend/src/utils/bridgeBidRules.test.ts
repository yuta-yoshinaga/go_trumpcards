import { describe, expect, it } from 'vitest';
import { type BridgeAuctionLike, canBid, canDouble, canRedouble, lastBidTeam } from './bridgeBidRules';

/** Four players: seats 0/2 = team 0, seats 1/3 = team 1. */
const PLAYERS = [{ team: 0 }, { team: 1 }, { team: 0 }, { team: 1 }];

function auction(overrides: Partial<BridgeAuctionLike>): BridgeAuctionLike {
  return {
    bidHistory: [],
    players: PLAYERS,
    bidPlayerIdx: 0,
    doubled: 0,
    contractLevel: 0,
    contractSuit: 0,
    ...overrides,
  };
}

describe('lastBidTeam', () => {
  it('returns -1 when no contract bid has been made', () => {
    expect(lastBidTeam(auction({ bidHistory: [{ playerIdx: 0, bidType: 0 }] }))).toBe(-1);
  });

  it('returns the team of the most recent normal bid, ignoring later passes/doubles', () => {
    const state = auction({
      bidHistory: [
        { playerIdx: 1, bidType: 1 }, // team 1 bids
        { playerIdx: 2, bidType: 0 }, // team 0 passes
      ],
    });
    expect(lastBidTeam(state)).toBe(1);
  });
});

describe('canDouble', () => {
  it('is false with no contract bid', () => {
    expect(canDouble(auction({ contractLevel: 0 }))).toBe(false);
  });

  it('is true when the last bid was made by the opponent and not yet doubled', () => {
    // seat 0 (team 0) to bid; last bid by seat 1 (team 1)
    const state = auction({
      bidPlayerIdx: 0,
      contractLevel: 1,
      doubled: 0,
      bidHistory: [{ playerIdx: 1, bidType: 1 }],
    });
    expect(canDouble(state)).toBe(true);
  });

  it("is false when the last bid was made by the bidder's own team", () => {
    // seat 0 (team 0) to bid; last bid by seat 2 (team 0)
    const state = auction({
      bidPlayerIdx: 0,
      contractLevel: 1,
      doubled: 0,
      bidHistory: [{ playerIdx: 2, bidType: 1 }],
    });
    expect(canDouble(state)).toBe(false);
  });

  it('is false when already doubled', () => {
    const state = auction({
      bidPlayerIdx: 0,
      contractLevel: 1,
      doubled: 1,
      bidHistory: [{ playerIdx: 1, bidType: 1 }],
    });
    expect(canDouble(state)).toBe(false);
  });
});

describe('canRedouble', () => {
  it('is false when not doubled', () => {
    expect(canRedouble(auction({ doubled: 0 }))).toBe(false);
  });

  it("is true when the bidder's own bid has been doubled by the opponent", () => {
    // seat 0 (team 0) to bid; own team made the last bid, opponent doubled
    const state = auction({
      bidPlayerIdx: 0,
      contractLevel: 1,
      doubled: 1,
      bidHistory: [
        { playerIdx: 0, bidType: 1 }, // team 0 bids
        { playerIdx: 1, bidType: 2 }, // team 1 doubles
      ],
    });
    expect(canRedouble(state)).toBe(true);
  });

  it('is false when the opposing team doubled our... i.e. bidder is not on the doubled team', () => {
    // seat 1 (team 1) to bid; last bid by team 0, so team 1 cannot redouble
    const state = auction({
      bidPlayerIdx: 1,
      contractLevel: 1,
      doubled: 1,
      bidHistory: [
        { playerIdx: 0, bidType: 1 }, // team 0 bids
        { playerIdx: 1, bidType: 2 }, // team 1 doubles
      ],
    });
    expect(canRedouble(state)).toBe(false);
  });

  it('is false when already redoubled', () => {
    const state = auction({
      bidPlayerIdx: 0,
      contractLevel: 1,
      doubled: 2,
      bidHistory: [{ playerIdx: 0, bidType: 1 }],
    });
    expect(canRedouble(state)).toBe(false);
  });
});

describe('canBid', () => {
  it('allows any bid when there is no contract yet', () => {
    expect(canBid(0, 0, 1, 1)).toBe(true);
  });

  it('allows a higher level', () => {
    expect(canBid(1, 5, 2, 1)).toBe(true);
  });

  it('allows the same level with a higher denomination', () => {
    expect(canBid(2, 3, 2, 4)).toBe(true);
  });

  it('rejects the same level with an equal or lower denomination', () => {
    expect(canBid(2, 3, 2, 3)).toBe(false);
    expect(canBid(2, 3, 2, 1)).toBe(false);
  });

  it('rejects a lower level', () => {
    expect(canBid(3, 1, 2, 5)).toBe(false);
  });
});
