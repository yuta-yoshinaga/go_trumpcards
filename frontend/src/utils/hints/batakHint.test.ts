import { describe, expect, it } from 'vitest';
import type { BatakResponse, Card } from '../../types/card';
import { BatakPhase } from '../../types/phases';
import { getBatakHint } from './batakHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<BatakResponse> = {}): BatakResponse {
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
      },
    ],
    phase: BatakPhase.BID,
    roundNumber: 1,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    declarerIdx: -1,
    highBid: 0,
    minLegalBid: 5,
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

describe('getBatakHint', () => {
  it('returns null when no human', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getBatakHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getBatakHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getBatakHint(makeState({ phase: BatakPhase.TRICK_END }))).toBeNull();
  });

  it('suggests pass when estimated bid is below minLegalBid (does not emit bid:0)', () => {
    const result = getBatakHint(makeState());
    expect(result?.targetAction).toBe('pass');
    expect(result?.targetAction).not.toBe('bid:0');
    expect(result?.reason).toBe('hint.passEstimate');
  });

  it('suggests pass when minLegalBid is 0 (only pass possible)', () => {
    const state = makeState({ minLegalBid: 0 });
    state.players[0].cards = [
      card('SPADE', 14),
      card('SPADE', 13),
      card('SPADE', 12),
      card('SPADE', 11),
      card('HEART', 14),
      card('DIAMOND', 14),
      card('CLOVER', 14),
    ];
    const result = getBatakHint(state);
    expect(result?.targetAction).toBe('pass');
    expect(result?.targetAction).not.toBe('bid:0');
    expect(result?.reason).toBe('hint.passEstimate');
  });

  it('returns bid hint when estimated tricks reach minLegalBid (5-13)', () => {
    const state = makeState({ minLegalBid: 5 });
    state.players[0].cards = [
      card('SPADE', 14),
      card('SPADE', 13),
      card('SPADE', 12),
      card('SPADE', 11),
      card('HEART', 14),
      card('DIAMOND', 14),
      card('CLOVER', 14),
    ];
    const result = getBatakHint(state);
    expect(result?.targetAction).toMatch(/^bid:[5-9]|1[0-3]$/);
    expect(result?.reason).toBe('hint.bidEstimate');
  });

  it('returns null in bid phase when not human turn', () => {
    expect(getBatakHint(makeState({ bidPlayerIdx: 1 }))).toBeNull();
  });

  it('bid confidence strong with many high cards', () => {
    const state = makeState();
    state.players[0].cards = [card('SPADE', 14), card('HEART', 13), card('DIAMOND', 12), card('CLOVER', 11)];
    expect(getBatakHint(state)?.confidence).toBe('strong');
  });

  it('bid confidence moderate with few high cards', () => {
    const state = makeState();
    state.players[0].cards = [card('CLOVER', 3), card('DIAMOND', 5), card('HEART', 7)];
    expect(getBatakHint(state)?.confidence).toBe('moderate');
  });

  it('suggests leading non-spade when not broken', () => {
    const state = makeState({ phase: BatakPhase.PLAY, currentPlayerIdx: 0 });
    expect(getBatakHint(state)?.reason).toBe('hint.leadNonSpade');
  });

  it('suggests strategic lead when broken', () => {
    const state = makeState({ phase: BatakPhase.PLAY, currentPlayerIdx: 0, spadesBroken: true });
    expect(getBatakHint(state)?.reason).toBe('hint.leadStrategic');
  });

  it('suggests follow suit when holding lead suit', () => {
    const state = makeState({
      phase: BatakPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    expect(getBatakHint(state)?.reason).toBe('hint.followSuit');
  });

  it('suggests mustTrumpWithSpade when void with spade', () => {
    const state = makeState({
      phase: BatakPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 7) }],
    });
    state.players[0].cards = [card('SPADE', 13), card('CLOVER', 5)];
    const result = getBatakHint(state);
    expect(result?.reason).toBe('hint.mustTrumpWithSpade');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests discard lowest when void with no spade', () => {
    const state = makeState({
      phase: BatakPhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 7) }],
    });
    state.players[0].cards = [card('CLOVER', 3), card('HEART', 5)];
    expect(getBatakHint(state)?.reason).toBe('hint.discardLowest');
  });

  it('returns null when not current player turn in play', () => {
    expect(getBatakHint(makeState({ phase: BatakPhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });
});
