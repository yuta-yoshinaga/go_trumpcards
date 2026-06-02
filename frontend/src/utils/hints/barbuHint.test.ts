import { describe, expect, it } from 'vitest';
import type { BarbuResponse } from '../../types/card';
import { getBarbuHint } from './barbuHint';

function baseState(overrides: Partial<BarbuResponse> = {}): BarbuResponse {
  return {
    message: '',
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 1,
        cards: [{ design: 'SPADE', value: 5 }],
        trickCount: 0,
        dominoRank: 0,
        totalScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 1, cards: [], trickCount: 0, dominoRank: 0, totalScore: 0 },
    ],
    phase: 'play',
    dealNumber: 0,
    totalDeals: 28,
    dealerIdx: 0,
    currentTurn: 0,
    currentContract: 0,
    trumpSuit: -1,
    trickNumber: 1,
    currentTrick: [],
    lastTrick: [],
    lastTrickWinner: -1,
    tablePlaced: [0, 0, 0, 0, 0],
    dominoPlayable: [],
    usedContracts: [false, false, false, false, false, false, false],
    gameEndFlag: false,
    config: { cpuDifficulty: 1 },
    roundWinners: [],
    lastDealDetail: null,
    ...overrides,
  } as BarbuResponse;
}

describe('getBarbuHint', () => {
  it('returns null when state is missing', () => {
    expect(getBarbuHint(null as unknown as BarbuResponse)).toBeNull();
  });

  it('returns null when the game has ended', () => {
    expect(getBarbuHint(baseState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null outside the play phase', () => {
    expect(getBarbuHint(baseState({ phase: 'selectContract' }))).toBeNull();
  });

  it('returns null when it is not the human turn', () => {
    expect(getBarbuHint(baseState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null when the human hand is empty', () => {
    const s = baseState();
    s.players[0].cards = [];
    expect(getBarbuHint(s)).toBeNull();
  });

  it('advises avoiding tricks on negative contracts', () => {
    const hint = getBarbuHint(baseState({ currentContract: 1 }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('barbu.hint.avoid');
  });

  it('advises winning on the Trumps contract', () => {
    const hint = getBarbuHint(baseState({ currentContract: 5, trumpSuit: 1 }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('barbu.hint.win');
  });

  it('advises placing a domino when one is playable', () => {
    const hint = getBarbuHint(baseState({ currentContract: 6, dominoPlayable: [0] }));
    expect(hint?.targetAction).toBe('play');
    expect(hint?.reason).toBe('barbu.hint.placeDomino');
  });

  it('advises passing in Dominoes when nothing is playable', () => {
    const hint = getBarbuHint(baseState({ currentContract: 6, dominoPlayable: [] }));
    expect(hint?.targetAction).toBe('pass');
    expect(hint?.reason).toBe('barbu.hint.pass');
  });
});
