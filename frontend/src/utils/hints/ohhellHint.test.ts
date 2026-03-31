import { describe, expect, it } from 'vitest';
import type { Card, OhHellResponse } from '../../types/card';
import { OhHellPhase } from '../../types/phases';
import { getOhHellHint } from './ohhellHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<OhHellResponse> = {}): OhHellResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 5,
        cards: [card('SPADE', 14), card('HEART', 13), card('CLOVER', 5)],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 5,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 5,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 5,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
    ],
    phase: OhHellPhase.BID,
    roundNumber: 1,
    totalRounds: 7,
    handSize: 5,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 3,
    currentTrick: [],
    trumpCard: card('HEART', 10),
    trumpSuit: 3,
    restrictedBid: -1,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 0, maxHandSize: 7, scoringVariant: 0, roundDirection: 0 },
    message: '',
    ...overrides,
  };
}

describe('getOhHellHint', () => {
  // Null/guard conditions
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getOhHellHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getOhHellHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getOhHellHint(makeState({ phase: OhHellPhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getOhHellHint(makeState({ phase: OhHellPhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getOhHellHint(makeState({ phase: OhHellPhase.GAME_END }))).toBeNull();
  });

  // Bid phase
  it('returns null in BID phase when not human bid turn', () => {
    expect(getOhHellHint(makeState({ phase: OhHellPhase.BID, bidPlayerIdx: 1 }))).toBeNull();
  });

  it('returns bid hint with strong confidence for many high cards', () => {
    const state = makeState({ phase: OhHellPhase.BID, bidPlayerIdx: 0, trumpSuit: 3 });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 13), card('HEART', 12), card('DIAMOND', 11)];
    const result = getOhHellHint(state);
    expect(result?.targetAction).toMatch(/^bid:\d+$/);
    expect(result?.reason).toBe('hint.bidEstimate');
    expect(result?.confidence).toBe('strong');
  });

  it('returns bid hint with moderate confidence for few high cards', () => {
    const state = makeState({ phase: OhHellPhase.BID, bidPlayerIdx: 0 });
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    const result = getOhHellHint(state);
    expect(result?.reason).toBe('hint.bidEstimate');
    expect(result?.confidence).toBe('moderate');
  });

  it('adjusts bid when estimated equals restrictedBid', () => {
    const state = makeState({ phase: OhHellPhase.BID, bidPlayerIdx: 0, restrictedBid: 1 });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 3), card('DIAMOND', 5)];
    const result = getOhHellHint(state);
    expect(result?.targetAction).toMatch(/^bid:\d+$/);
    const bidValue = Number(result?.targetAction.split(':')[1]);
    expect(bidValue).not.toBe(1);
  });

  // Play phase
  it('returns null in PLAY phase when not human turn', () => {
    expect(getOhHellHint(makeState({ phase: OhHellPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests strategic lead when leading', () => {
    const state = makeState({ phase: OhHellPhase.PLAY, currentPlayerIdx: 0 });
    const result = getOhHellHint(state);
    expect(result?.reason).toBe('hint.leadStrategic');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests following suit', () => {
    const state = makeState({
      phase: OhHellPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 14)];
    const result = getOhHellHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests trumping when void in led suit and have trump', () => {
    const state = makeState({
      phase: OhHellPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 5)];
    const result = getOhHellHint(state);
    expect(result?.reason).toBe('hint.trumpWithCard');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests discarding lowest when void with no trump', () => {
    const state = makeState({
      phase: OhHellPhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getOhHellHint(state);
    expect(result?.reason).toBe('hint.discardLowest');
    expect(result?.confidence).toBe('moderate');
  });
});
