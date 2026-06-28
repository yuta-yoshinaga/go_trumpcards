import { describe, expect, it } from 'vitest';
import type { Card, PrsiResponse } from '../../../types/card';
import { formatPrsiState } from './prsiFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<PrsiResponse> = {}): PrsiResponse {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: 2, cards: [card('HEART', 10), card('SPADE', 7)] },
      { id: 1, isHuman: false, cardCount: 3, cards: [] },
      { id: 2, isHuman: false, cardCount: 4, cards: [] },
      { id: 3, isHuman: false, cardCount: 5, cards: [] },
    ],
    phase: 0,
    currentPlayerIdx: 0,
    discardTop: card('HEART', 9),
    drawPileCount: 18,
    penaltyDrawCount: 0,
    pendingSkips: 0,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    config: { cpuDifficulty: 1 },
    ...overrides,
  };
}

describe('formatPrsiState', () => {
  it('shows the header, draw pile, discard top, and turn', () => {
    const out = formatPrsiState(makeState());
    expect(out).toContain('Prší');
    expect(out).toContain('draw pile: 18');
    expect(out).toContain('discard:');
    expect(out).toContain('turn:');
  });

  it('shows each player card count', () => {
    const out = formatPrsiState(makeState());
    expect(out).toContain('cards=2');
    expect(out).toContain('cards=3');
    expect(out).toContain('cards=5');
  });

  it('shows the human indexed hand', () => {
    const out = formatPrsiState(makeState());
    expect(out).toContain('[0]');
    expect(out).toContain('[1]');
  });

  it('shows the penalty indicator when a 7-stack is active', () => {
    const out = formatPrsiState(makeState({ penaltyDrawCount: 4 }));
    expect(out).toContain('penalty: 4');
  });

  it('omits the penalty indicator when there is no penalty', () => {
    const out = formatPrsiState(makeState());
    expect(out).not.toContain('penalty:');
  });

  it('omits the discard line when there is no discard top', () => {
    const out = formatPrsiState(makeState({ discardTop: null }));
    expect(out).not.toContain('discard:');
  });

  it('shows the winner and no turn line on game end', () => {
    const out = formatPrsiState(makeState({ gameEndFlag: true, winnerIdx: 0 }));
    expect(out).toContain('Game Over! Winner:');
    expect(out).not.toContain('turn:');
  });

  it('includes a server message when present', () => {
    const out = formatPrsiState(makeState({ message: 'Your turn' }));
    expect(out).toContain('Your turn');
  });
});
