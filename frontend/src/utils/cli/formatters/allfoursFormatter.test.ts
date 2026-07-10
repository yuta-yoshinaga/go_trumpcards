import { describe, expect, it } from 'vitest';
import type { AllFoursResponse } from '../../../types/card';
import { formatAllFoursState } from './allfoursFormatter';

function baseState(overrides: Partial<AllFoursResponse> = {}): AllFoursResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      { id: 1, isHuman: false, cardCount: 2, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
    ],
    phase: 0,
    roundNumber: 1,
    trickNumber: 0,
    dealerIdx: 1,
    nonDealerIdx: 0,
    currentPlayerIdx: 0,
    trumpSuit: 3,
    turnUp: { design: 'HEART', value: 7 },
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

describe('formatAllFoursState', () => {
  it('renders header, dealer, trump and turn-up', () => {
    const out = formatAllFoursState(baseState());
    expect(out).toContain('All Fours');
    expect(out).toContain('phase: BEG');
    expect(out).toContain('trump: Hearts');
    expect(out).toContain('dealer: 1');
  });

  it('renders human hand and current trick', () => {
    const state = baseState({
      phase: 2,
      players: [
        {
          id: 0,
          isHuman: true,
          cardCount: 1,
          cards: [{ design: 'SPADE', value: 1 }],
          roundScore: 0,
          cumulativeScore: 0,
          trickCount: 0,
        },
        { id: 1, isHuman: false, cardCount: 1, cards: [], roundScore: 0, cumulativeScore: 0, trickCount: 0 },
      ],
      currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 10 } }],
      message: 'your turn',
    });
    const out = formatAllFoursState(state);
    expect(out).toContain('phase: PLAY');
    expect(out).toContain('hand:');
    expect(out).toContain('trick:');
    expect(out).toContain('your turn');
  });

  it('handles null turn-up', () => {
    const out = formatAllFoursState(baseState({ turnUp: null, trumpSuit: 0 }));
    expect(out).toContain('turn-up: --');
    expect(out).toContain('trump: (unset)');
  });
});
