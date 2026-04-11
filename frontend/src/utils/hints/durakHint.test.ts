import { describe, expect, it } from 'vitest';
import type { Card, DurakConfig, DurakResponse } from '../../types/card';
import { getDurakHint } from './durakHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

const defaultConfig: DurakConfig = { playerCount: 2, cpuDifficulty: 0, transferEnabled: false };

function makeState(overrides: Partial<DurakResponse> = {}): DurakResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        isFinished: false,
        cardCount: 4,
        cards: [card('SPADE', 6), card('CLOVER', 7), card('DIAMOND', 9), card('HEART', 12)],
      },
      { id: 1, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
    ],
    currentTurn: 0,
    phase: 0,
    attackerIdx: 0,
    defenderIdx: 1,
    tablePairs: [],
    trumpSuit: 'H',
    trumpCard: card('HEART', 10),
    stockCount: 12,
    loserIdx: -1,
    gameEndFlag: false,
    config: defaultConfig,
    cpuActions: [],
    humanAction: null,
    boutNumber: 1,
    sortMode: 0,
    message: '',
    ...overrides,
  };
}

describe('getDurakHint', () => {
  it('returns null when game ended', () => {
    expect(getDurakHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when no turn match', () => {
    expect(getDurakHint(makeState({ attackerIdx: 1, defenderIdx: 0 }))).toBeNull();
  });

  it('recommends attacking with low non-trump on a fresh attack', () => {
    const hint = getDurakHint(makeState());
    expect(hint?.targetAction).toBe('attack');
    expect(hint?.reason).toBe('hint.attackLowNonTrump');
  });

  it('recommends attack with only trumps when hand has no non-trump', () => {
    const state = makeState({
      players: [
        { id: 0, isHuman: true, isFinished: false, cardCount: 2, cards: [card('HEART', 7), card('HEART', 11)] },
        { id: 1, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
      ],
    });
    const hint = getDurakHint(state);
    expect(hint?.targetAction).toBe('attack');
    expect(hint?.reason).toBe('hint.attackOnlyTrumps');
  });

  it('recommends pass when continuing a bout without matching values', () => {
    const state = makeState({
      tablePairs: [{ attack: card('DIAMOND', 5), defense: card('DIAMOND', 10) }],
    });
    const hint = getDurakHint(state);
    expect(hint?.targetAction).toBe('pass');
  });

  it('recommends defending with same-suit higher card', () => {
    const state = makeState({
      attackerIdx: 1,
      defenderIdx: 0,
      phase: 1,
      tablePairs: [{ attack: card('SPADE', 4), defense: null }],
    });
    const hint = getDurakHint(state);
    expect(hint?.targetAction).toBe('defend');
    expect(hint?.reason).toBe('hint.defendSameSuit');
  });

  it('recommends defending with trump when no same-suit available', () => {
    const state = makeState({
      attackerIdx: 1,
      defenderIdx: 0,
      phase: 1,
      players: [
        { id: 0, isHuman: true, isFinished: false, cardCount: 2, cards: [card('HEART', 8), card('CLOVER', 2)] },
        { id: 1, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
      ],
      tablePairs: [{ attack: card('DIAMOND', 10), defense: null }],
    });
    const hint = getDurakHint(state);
    expect(hint?.targetAction).toBe('defend');
    expect(hint?.reason).toBe('hint.defendWithTrump');
  });

  it('recommends take when unable to defend', () => {
    const state = makeState({
      attackerIdx: 1,
      defenderIdx: 0,
      phase: 1,
      players: [
        { id: 0, isHuman: true, isFinished: false, cardCount: 2, cards: [card('CLOVER', 4), card('DIAMOND', 3)] },
        { id: 1, isHuman: false, isFinished: false, cardCount: 6, cards: [] },
      ],
      tablePairs: [{ attack: card('SPADE', 13), defense: null }],
    });
    const hint = getDurakHint(state);
    expect(hint?.targetAction).toBe('take');
  });
});
