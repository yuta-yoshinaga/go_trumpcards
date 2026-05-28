import { describe, expect, it } from 'vitest';
import type { DoudizhuResponse } from '../../types/card';
import { getDoudizhuHint } from './doudizhuHint';

function makeState(overrides: Partial<DoudizhuResponse> = {}): DoudizhuResponse {
  return {
    players: [
      { id: 0, isHuman: true, isFinished: false, isLandlord: true, cardCount: 20, cards: [] },
      { id: 1, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
      { id: 2, isHuman: false, isFinished: false, isLandlord: false, cardCount: 17, cards: [] },
    ],
    phase: 'play',
    currentTurn: 0,
    tableCards: [],
    tableCombo: '',
    kittyCards: [],
    landlordIdx: 0,
    baseBid: 1,
    highestBid: 1,
    bombCount: 0,
    scores: [0, 0, 0],
    gameEndFlag: false,
    config: { cpuDifficulty: 0 },
    cpuActions: [],
    humanAction: null,
    message: '',
    ...overrides,
  };
}

describe('getDoudizhuHint', () => {
  it('returns null when game has ended', () => {
    expect(getDoudizhuHint(makeState({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null when not human turn', () => {
    expect(getDoudizhuHint(makeState({ currentTurn: 1 }))).toBeNull();
  });

  it('returns null during bid phase', () => {
    expect(getDoudizhuHint(makeState({ phase: 'bid' }))).toBeNull();
  });

  it('suggests play lowest when leading', () => {
    const hint = getDoudizhuHint(makeState({ tableCards: [] }));
    expect(hint?.targetAction).toBe('play lowest');
  });

  it('suggests pass when table has cards', () => {
    const hint = getDoudizhuHint(makeState({ tableCards: [{ design: 'SPADE', value: 10 }] }));
    expect(hint?.targetAction).toBe('pass');
  });
});
