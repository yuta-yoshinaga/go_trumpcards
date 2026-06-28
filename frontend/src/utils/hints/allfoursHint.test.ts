import { describe, expect, it } from 'vitest';
import type { AllFoursResponse, Card } from '../../types/card';
import { getAllFoursHint } from './allfoursHint';

function state(overrides: Partial<AllFoursResponse> = {}, humanCards: Card[] = []): AllFoursResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: humanCards.length,
        cards: humanCards,
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
      },
      { id: 1, isHuman: false, cardCount: 0, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 1,
    dealerIdx: 1,
    nonDealerIdx: 0,
    currentPlayerIdx: 0,
    trumpSuit: 3,
    turnUp: null,
    runCount: 0,
    currentTrick: [],
    gameEndFlag: false,
    winnerIdx: -1,
    leadPlayerIdx: -1,
    validPlayIndices: [],
    config: { cpuDifficulty: 1, pointLimit: 7 },
    message: '',
    ...overrides,
  } as AllFoursResponse;
}

describe('getAllFoursHint', () => {
  it('returns null when no human', () => {
    const s = state();
    s.players = s.players.map((p) => ({ ...p, isHuman: false }));
    expect(getAllFoursHint(s)).toBeNull();
  });

  it('beg phase: weak hand begs', () => {
    const s = state({ phase: 0 }, [{ design: 'SPADE', value: 5 }]);
    const hint = getAllFoursHint(s);
    expect(hint?.targetAction).toBe('beg:true');
  });

  it('beg phase: strong trump stands', () => {
    const s = state({ phase: 0, trumpSuit: 3 }, [
      { design: 'HEART', value: 1 },
      { design: 'HEART', value: 13 },
    ]);
    const hint = getAllFoursHint(s);
    expect(hint?.targetAction).toBe('beg:false');
  });

  it('beg phase: not the non-dealer returns null', () => {
    const s = state({ phase: 0, nonDealerIdx: 1 }, [{ design: 'SPADE', value: 5 }]);
    expect(getAllFoursHint(s)).toBeNull();
  });

  it('gift phase: dealer with weak trump runs', () => {
    const s = state({ phase: 1, dealerIdx: 0, trumpSuit: 3 }, [{ design: 'SPADE', value: 5 }]);
    const hint = getAllFoursHint(s);
    expect(hint?.targetAction).toBe('respond:true');
  });

  it('gift phase: dealer with strong trump gifts', () => {
    const s = state({ phase: 1, dealerIdx: 0, trumpSuit: 3 }, [
      { design: 'HEART', value: 1 },
      { design: 'HEART', value: 12 },
    ]);
    const hint = getAllFoursHint(s);
    expect(hint?.targetAction).toBe('respond:false');
  });

  it('gift phase: not the dealer returns null', () => {
    const s = state({ phase: 1, dealerIdx: 1 }, [{ design: 'SPADE', value: 5 }]);
    expect(getAllFoursHint(s)).toBeNull();
  });

  it('play phase: lead suggests leading strong', () => {
    const s = state({ phase: 2, currentTrick: [] }, [{ design: 'HEART', value: 1 }]);
    const hint = getAllFoursHint(s);
    expect(hint?.reason).toBe('hint.leadStrong');
  });

  it('play phase: follow suit when holding lead suit', () => {
    const s = state(
      { phase: 2, trumpSuit: 4, currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 10 } }] },
      [{ design: 'SPADE', value: 5 }],
    );
    const hint = getAllFoursHint(s);
    expect(hint?.reason).toBe('hint.followSuit');
  });

  it('play phase: trump cut when no lead suit but holding trump', () => {
    const s = state(
      { phase: 2, trumpSuit: 3, currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 10 } }] },
      [{ design: 'HEART', value: 5 }],
    );
    const hint = getAllFoursHint(s);
    expect(hint?.reason).toBe('hint.trumpCut');
  });

  it('play phase: discard low when no lead and no trump', () => {
    const s = state(
      { phase: 2, trumpSuit: 3, currentTrick: [{ playerIdx: 1, card: { design: 'SPADE', value: 10 } }] },
      [{ design: 'CLOVER', value: 5 }],
    );
    const hint = getAllFoursHint(s);
    expect(hint?.reason).toBe('hint.discardLow');
  });

  it('play phase: not current player returns null', () => {
    const s = state({ phase: 2, currentPlayerIdx: 1 }, [{ design: 'HEART', value: 1 }]);
    expect(getAllFoursHint(s)).toBeNull();
  });

  it('returns null for trick-end phase', () => {
    const s = state({ phase: 3 }, [{ design: 'HEART', value: 1 }]);
    expect(getAllFoursHint(s)).toBeNull();
  });
});
