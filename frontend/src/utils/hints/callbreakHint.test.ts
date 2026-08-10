import { describe, expect, it } from 'vitest';
import type { CallBreakResponse, Card } from '../../types/card';
import { CallBreakPhase } from '../../types/phases';
import { getCallBreakHint } from './callbreakHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<CallBreakResponse> = {}): CallBreakResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [card('SPADE', 13), card('HEART', 11), card('CLOVER', 5)],
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
    phase: CallBreakPhase.BID,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    currentTrick: [],
    spadesBroken: false,
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, maxRounds: 5 },
    validPlayIndices: [],
    ...overrides,
  };
}

describe('getCallBreakHint', () => {
  it('returns null when no human', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getCallBreakHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getCallBreakHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getCallBreakHint(makeState({ phase: CallBreakPhase.TRICK_END }))).toBeNull();
  });

  it('returns bid hint when human turn to bid', () => {
    const result = getCallBreakHint(makeState());
    expect(result?.targetAction).toMatch(/^bid:\d+$/);
    expect(result?.reason).toBe('hint.bidEstimate');
  });

  it('returns null in bid phase when not human turn', () => {
    expect(getCallBreakHint(makeState({ bidPlayerIdx: 1 }))).toBeNull();
  });

  it('bid confidence strong with many high cards', () => {
    const state = makeState();
    state.players[0].cards = [card('SPADE', 14), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11)];
    expect(getCallBreakHint(state)?.confidence).toBe('strong');
  });

  it('bid confidence moderate with few high cards', () => {
    const state = makeState();
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    expect(getCallBreakHint(state)?.confidence).toBe('moderate');
  });

  it('suggests leading non-spade when not broken', () => {
    const state = makeState({ phase: CallBreakPhase.PLAY, currentPlayerIdx: 0 });
    expect(getCallBreakHint(state)?.reason).toBe('hint.leadNonSpade');
  });

  it('suggests strategic lead when broken', () => {
    const state = makeState({ phase: CallBreakPhase.PLAY, currentPlayerIdx: 0, spadesBroken: true });
    expect(getCallBreakHint(state)?.reason).toBe('hint.leadStrategic');
  });

  it('suggests follow suit when holding lead suit', () => {
    const state = makeState({
      phase: CallBreakPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    expect(getCallBreakHint(state)?.reason).toBe('hint.followSuit');
  });

  it('suggests mustTrumpWithSpade when void with spade', () => {
    const state = makeState({
      phase: CallBreakPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 7) }],
    });
    state.players[0].cards = [card('SPADE', 13), card('CLOVER', 5)];
    const result = getCallBreakHint(state);
    expect(result?.reason).toBe('hint.mustTrumpWithSpade');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests discard lowest when void with no spade', () => {
    const state = makeState({
      phase: CallBreakPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 7) }],
    });
    state.players[0].cards = [card('CLOVER', 3), card('HEART', 5)];
    expect(getCallBreakHint(state)?.reason).toBe('hint.discardLowest');
  });

  it('returns null when not current player turn in play', () => {
    expect(getCallBreakHint(makeState({ phase: CallBreakPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });
});
