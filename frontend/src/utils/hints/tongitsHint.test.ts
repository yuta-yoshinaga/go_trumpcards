import { describe, expect, it } from 'vitest';
import i18n from '../../i18n';
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
  it('suggests taking a discard that completes a run', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.DRAW }))?.targetAction).toBe('takeDiscard');
  });
  it('suggests drawing stock when discard does not make a meld', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.DRAW, discardTop: card('DIAMOND', 2) }))?.targetAction).toBe(
      'drawStock',
    );
  });
  it('suggests drawing stock when there is no discard', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.DRAW, discardTop: null }))?.targetAction).toBe('drawStock');
  });
  it('suggests a challenge when remaining points are low', () => {
    expect(getTongitsHint(base({ remainingPoints: 5 }))?.targetAction).toBe('challenge');
  });
  it('suggests the best discard when challenge is unavailable', () => {
    expect(getTongitsHint(base({ remainingPoints: 6 }))?.targetAction).toBe('card-2');
  });
  it('returns no hint during round end', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.ROUND_END }))).toBeNull();
  });
  it('returns no hint after the game ends', () => {
    expect(getTongitsHint(base({ gameEndFlag: true }))).toBeNull();
  });
  it('returns no hint while a CPU is acting', () => {
    expect(getTongitsHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });
  it('does not suggest a pair as a meld', () => {
    expect(getTongitsHint(base({ phase: TongitsPhase.DRAW, discardTop: card('DIAMOND', 5) }))?.targetAction).toBe(
      'drawStock',
    );
  });
  it('returns no hint for an empty human hand', () => {
    const initial = base();
    expect(
      getTongitsHint(base({ players: [{ ...initial.players[0], cards: [], cardCount: 0 }, initial.players[1]] })),
    ).toBeNull();
  });
  it('uses translated hint sentences instead of exposing raw i18n keys', () => {
    const hints = [
      getTongitsHint(base({ phase: TongitsPhase.DRAW })),
      getTongitsHint(base({ phase: TongitsPhase.DRAW, discardTop: card('DIAMOND', 2) })),
      getTongitsHint(base({ remainingPoints: 5 })),
      getTongitsHint(base({ remainingPoints: 6 })),
    ];
    for (const hint of hints) {
      const sentence = i18n.t(`tongits:${hint?.reason}`);
      expect(sentence).not.toContain('tongits.');
      expect(sentence).not.toBe(hint?.reason);
    }
  });
});
