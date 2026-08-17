import { describe, expect, it } from 'vitest';
import type { BuraPlayer, BuraResponse, Card, CardDesign } from '../../../types/card';
import { formatBuraState } from './buraFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const human: BuraPlayer = {
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [card('SPADE', 1), card('SPADE', 10)],
  points: 12,
  hidden: false,
};

// A hidden seat arrives with a count and NO cards.
const cpu: BuraPlayer = { id: 1, isHuman: false, cardCount: 2, cards: [], points: 5, hidden: true };

function makeState(overrides?: Partial<BuraResponse>): BuraResponse {
  return {
    players: [human, cpu],
    phase: 0,
    trickNumber: 3,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentLead: [],
    trumpSuit: 2,
    trumpCard: card('HEART', 7),
    stockRemaining: 20,
    winThreshold: 31,
    gameEndFlag: false,
    winnerIdx: -1,
    isDraw: false,
    winningCombinations: ['bura', 'moscow', 'littleMoscow', 'molodka'],
    message: '',
    ...overrides,
  };
}

describe('formatBuraState', () => {
  it('indexes the human hand so the play command can name a card', () => {
    const out = formatBuraState(makeState());
    expect(out).toContain('0:');
    expect(out).toContain('1:');
  });

  it('draws the opponent from its count, never from its cards', () => {
    const out = formatBuraState(makeState());
    const cpuLine = out.split('\n').find((l) => l.startsWith('cpu:')) ?? '';
    expect(cpuLine).toContain('[??] [??]');
  });

  it('reports the trump suit once the indicator has been drawn', () => {
    expect(formatBuraState(makeState({ trumpCard: undefined }))).toContain('trump suit: 2');
  });

  it('shows the led cards while a trick is open', () => {
    const out = formatBuraState(makeState({ currentLead: [card('CLOVER', 13)] }));
    expect(out).toContain('led:');
  });

  it('distinguishes a draw from a loss', () => {
    expect(formatBuraState(makeState({ gameEndFlag: true, isDraw: true, winnerIdx: -1 }))).toContain(
      'draw -- nobody claimed',
    );
    expect(formatBuraState(makeState({ gameEndFlag: true, winnerIdx: 1 }))).toContain('you lose');
    expect(formatBuraState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('you win');
  });
});
