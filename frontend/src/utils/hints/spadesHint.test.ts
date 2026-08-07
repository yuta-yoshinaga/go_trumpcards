import { describe, expect, it } from 'vitest';
import type { Card, SpadesResponse } from '../../types/card';
import { SpadesPhase } from '../../types/phases';
import { getSpadesHint } from './spadesHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<SpadesResponse> = {}): SpadesResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [card('SPADE', 14), card('HEART', 13), card('CLOVER', 5)],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        bags: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        bags: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 13,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        bags: 0,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 13,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        bags: 0,
      },
    ],
    phase: SpadesPhase.BID,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    currentTrick: [],
    spadesBroken: false,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    validPlayIndices: [],
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 500, nilBonus: 100, bagPenaltyThreshold: 10 },
    ...overrides,
  };
}

describe('getSpadesHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getSpadesHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getSpadesHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getSpadesHint(makeState({ phase: SpadesPhase.TRICK_END }))).toBeNull();
  });

  // Bid phase
  it('returns bid hint when it is human turn to bid', () => {
    const result = getSpadesHint(makeState());
    expect(result?.targetAction).toMatch(/^bid:\d+$/);
    expect(result?.reason).toBe('hint.bidEstimate');
  });

  it('returns null in bid phase when not human turn', () => {
    expect(getSpadesHint(makeState({ bidPlayerIdx: 1 }))).toBeNull();
  });

  it('bid confidence is strong with many high cards', () => {
    const state = makeState();
    state.players[0].cards = [card('SPADE', 14), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11)];
    const result = getSpadesHint(state);
    expect(result?.confidence).toBe('strong');
  });

  it('bid confidence is moderate with few high cards', () => {
    const state = makeState();
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    const result = getSpadesHint(state);
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase - leading
  it('suggests leading non-spade when spades not broken', () => {
    const state = makeState({ phase: SpadesPhase.PLAY, currentPlayerIdx: 0 });
    const result = getSpadesHint(state);
    expect(result?.reason).toBe('hint.leadNonSpade');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests strategic lead when spades broken', () => {
    const state = makeState({ phase: SpadesPhase.PLAY, currentPlayerIdx: 0, spadesBroken: true });
    const result = getSpadesHint(state);
    expect(result?.reason).toBe('hint.leadStrategic');
  });

  // Play phase - following
  it('suggests following suit', () => {
    const state = makeState({
      phase: SpadesPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    const result = getSpadesHint(state);
    expect(result?.reason).toBe('hint.followSuit');
  });

  it('suggests trumping with spade when void', () => {
    const state = makeState({
      phase: SpadesPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('HEART', 3), card('CLOVER', 5)];
    const result = getSpadesHint(state);
    expect(result?.reason).toBe('hint.trumpWithSpade');
  });

  it('suggests discarding lowest when void with no spades', () => {
    const state = makeState({
      phase: SpadesPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 7) }],
    });
    state.players[0].cards = [card('HEART', 3), card('CLOVER', 5)];
    const result = getSpadesHint(state);
    expect(result?.reason).toBe('hint.discardLowest');
  });

  it('returns null when not current player turn in play phase', () => {
    const state = makeState({ phase: SpadesPhase.PLAY, currentPlayerIdx: 2 });
    expect(getSpadesHint(state)).toBeNull();
  });
});
