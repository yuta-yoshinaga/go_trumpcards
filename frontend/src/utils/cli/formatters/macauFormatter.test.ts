import { describe, expect, it } from 'vitest';
import type { Card, MacauResponse } from '../../../types/card';
import { formatMacauState } from './macauFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<MacauResponse> = {}): MacauResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('HEART', 5), card('SPADE', 7)],
        roundScore: 0,
        cumulativeScore: 0,
        hasDeclared: false,
      },
      { id: 1, isHuman: false, cardCount: 3, cards: [], roundScore: 0, cumulativeScore: 0, hasDeclared: false },
    ],
    phase: 0,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 9),
    drawPileCount: 20,
    chosenSuit: 0,
    penaltyDrawCount: 0,
    playableIndices: [],
    direction: 1,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 0, pointLimit: 100 },
    ...overrides,
  };
}

describe('formatMacauState', () => {
  it('formats header, discard and turn', () => {
    const out = formatMacauState(makeState());
    expect(out).toContain('Macau');
    expect(out).toContain('round: 1');
    expect(out).toContain('direction: ->');
    expect(out).toContain('discard:');
    expect(out).toContain('turn:');
  });

  it('shows reverse direction', () => {
    expect(formatMacauState(makeState({ direction: -1 }))).toContain('direction: <-');
  });

  it('shows chosen suit and penalty', () => {
    const out = formatMacauState(makeState({ chosenSuit: 3, penaltyDrawCount: 4 }));
    expect(out).toContain('chosen suit: Heart');
    expect(out).toContain('draw penalty: 4');
  });

  it('shows choose-suit prompt in phase 1', () => {
    expect(formatMacauState(makeState({ phase: 1 }))).toContain('Choose a suit');
  });

  it('shows declare prompt in phase 2', () => {
    expect(formatMacauState(makeState({ phase: 2 }))).toContain('Declare');
  });

  it('shows game over with winner', () => {
    const out = formatMacauState(makeState({ gameEndFlag: true, winnerIdx: 0, message: 'done' }));
    expect(out).toContain('Game Over!');
    expect(out).toContain('done');
  });
});
