import { describe, expect, it } from 'vitest';
import type { Card, TongitsResponse } from '../../types/card';
import { TongitsPhase } from '../../types/phases';
import { getTongitsHint } from './tongitsHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });
function base(overrides: Partial<TongitsResponse> = {}): TongitsResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 3,
        cards: [card('SPADE', 5), card('SPADE', 6), card('CLOVER', 13)],
        melds: [],
        roundScore: 0,
        cumulativeScore: 0,
      },
      { id: 1, isHuman: false, cardCount: 5, cards: [], melds: [], roundScore: 0, cumulativeScore: 0 },
    ],
    phase: TongitsPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('SPADE', 7),
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    isTongits: false,
    remainingPoints: 6,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 50 },
    ...overrides,
  };
}
describe('getTongitsHint', () => {
  it('suggests taking a discard that completes a meld', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.DRAW }))?.targetAction).toBe('takeDiscard');
  });
  it('suggests drawing stock when discard does not improve the hand', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.DRAW, discardTop: card('DIAMOND', 2) }))?.targetAction).toBe(
      'drawStock',
    );
  });
  it('suggests challenge using server remaining points', () => {
    expect(getTongitsHint(base({ remainingPoints: 5 }))?.targetAction).toBe('challenge');
  });
  it('suggests a discard when challenge is not available', () => {
    expect(getTongitsHint(base({ remainingPoints: 6 }))?.targetAction).toBe('card-2');
  });
  it('returns null when the game is over or CPU is acting', () => {
    expect(getTongitsHint(base({ gameEndFlag: true }))).toBeNull();
    expect(getTongitsHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });
});
