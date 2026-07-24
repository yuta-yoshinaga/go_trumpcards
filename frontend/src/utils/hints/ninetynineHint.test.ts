import { describe, expect, it } from 'vitest';
import type { Card, NinetyNineResponse } from '../../types/card';
import { NinetyNinePhase } from '../../types/phases';
import { getNinetyNineHint, ninetynineBidValue, ninetynineDeclaredTricks } from './ninetynineHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('ninetynineBidValue', () => {
  // Must match ninetyNineSuitBidValue in internal/domain/NinetyNine.go exactly.
  it('maps each suit to its domain bid value', () => {
    expect(ninetynineBidValue('DIAMOND')).toBe(0);
    expect(ninetynineBidValue('SPADE')).toBe(1);
    expect(ninetynineBidValue('HEART')).toBe(2);
    expect(ninetynineBidValue('CLOVER')).toBe(3);
  });

  it('treats any other design (e.g. JOKER) as 0', () => {
    expect(ninetynineBidValue('JOKER')).toBe(0);
  });
});

describe('ninetynineDeclaredTricks', () => {
  it('returns 0 for no cards', () => {
    expect(ninetynineDeclaredTricks([])).toBe(0);
  });

  it('sums the buried cards suit values (max bid = 3 clubs = 9)', () => {
    expect(ninetynineDeclaredTricks([card('SPADE', 1), card('HEART', 2), card('CLOVER', 3)])).toBe(6);
    expect(ninetynineDeclaredTricks([card('CLOVER', 3), card('CLOVER', 4), card('CLOVER', 5)])).toBe(9);
    expect(ninetynineDeclaredTricks([card('DIAMOND', 3), card('DIAMOND', 4), card('DIAMOND', 5)])).toBe(0);
  });
});

function makeState(overrides: Partial<NinetyNineResponse> = {}): NinetyNineResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 9,
        cards: [card('SPADE', 14), card('HEART', 13), card('CLOVER', 5)],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        buriedCount: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 9,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        buriedCount: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 9,
        cards: [],
        bid: -1,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        buriedCount: 0,
      },
    ],
    phase: NinetyNinePhase.BID,
    dealNumber: 1,
    targetScore: 100,
    handSize: 9,
    trickNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 2,
    trumpSuit: 1,
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: 0,
    config: { cpuDifficulty: 0, targetScore: 100 },
    message: '',
    ...overrides,
  };
}

describe('getNinetyNineHint', () => {
  it('returns null when no human player', () => {
    const state = makeState();
    state.players = state.players.map((p) => ({ ...p, isHuman: false }));
    expect(getNinetyNineHint(state)).toBeNull();
  });

  it('returns null when human has no cards', () => {
    const state = makeState();
    state.players[0].cards = [];
    expect(getNinetyNineHint(state)).toBeNull();
  });

  it('returns null in TRICK_END phase', () => {
    expect(getNinetyNineHint(makeState({ phase: NinetyNinePhase.TRICK_END }))).toBeNull();
  });

  it('returns null in ROUND_END phase', () => {
    expect(getNinetyNineHint(makeState({ phase: NinetyNinePhase.ROUND_END }))).toBeNull();
  });

  it('returns null in GAME_END phase', () => {
    expect(getNinetyNineHint(makeState({ phase: NinetyNinePhase.GAME_END }))).toBeNull();
  });

  // Bid (bury) phase
  it('returns null in BID phase when not human bid turn', () => {
    expect(getNinetyNineHint(makeState({ phase: NinetyNinePhase.BID, bidPlayerIdx: 1 }))).toBeNull();
  });

  it('suggests a strategic bury when human bid turn', () => {
    const result = getNinetyNineHint(makeState({ phase: NinetyNinePhase.BID, bidPlayerIdx: 0 }));
    expect(result?.targetAction).toBe('bury');
    expect(result?.reason).toBe('hint.buryStrategic');
    expect(result?.confidence).toBe('moderate');
  });

  // Play phase
  it('returns null in PLAY phase when not human turn', () => {
    expect(getNinetyNineHint(makeState({ phase: NinetyNinePhase.PLAY, currentPlayerIdx: 2 }))).toBeNull();
  });

  it('suggests strategic lead when leading', () => {
    const result = getNinetyNineHint(makeState({ phase: NinetyNinePhase.PLAY, currentPlayerIdx: 0 }));
    expect(result?.reason).toBe('hint.leadStrategic');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests following suit', () => {
    const state = makeState({
      phase: NinetyNinePhase.PLAY,
      currentPlayerIdx: 0,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('HEART', 10), card('SPADE', 14)];
    const result = getNinetyNineHint(state);
    expect(result?.reason).toBe('hint.followSuit');
    expect(result?.confidence).toBe('strong');
  });

  it('suggests trumping when void in led suit and have trump', () => {
    const state = makeState({
      phase: NinetyNinePhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('SPADE', 14), card('CLOVER', 5)];
    const result = getNinetyNineHint(state);
    expect(result?.reason).toBe('hint.trumpWithCard');
    expect(result?.confidence).toBe('moderate');
  });

  it('suggests discarding lowest when void with no trump', () => {
    const state = makeState({
      phase: NinetyNinePhase.PLAY,
      currentPlayerIdx: 0,
      trumpSuit: 1,
      currentTrick: [{ playerIdx: 1, card: card('HEART', 7) }],
    });
    state.players[0].cards = [card('DIAMOND', 3), card('CLOVER', 5)];
    const result = getNinetyNineHint(state);
    expect(result?.reason).toBe('hint.discardLowest');
    expect(result?.confidence).toBe('moderate');
  });
});
