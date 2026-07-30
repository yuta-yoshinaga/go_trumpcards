import { describe, expect, it } from 'vitest';
import type { CardDesign, ChineseTenCard, ChineseTenPlayer, ChineseTenResponse } from '../../../types/card';
import { formatChineseTenState } from './chinesetenFormatter';

const card = (design: CardDesign, value: number, points = 0): ChineseTenCard => ({
  design,
  value,
  points,
  isRed: design === 'HEART' || design === 'DIAMOND',
});

const human: ChineseTenPlayer = {
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [card('SPADE', 1), card('HEART', 9, 10)],
  captured: [card('HEART', 5, 5)],
  score: 5,
  hidden: false,
};

const cpu: ChineseTenPlayer = {
  id: 1,
  isHuman: false,
  cardCount: 2,
  cards: [],
  captured: [card('DIAMOND', 1, 20)],
  score: 20,
  hidden: true,
};

function makeState(overrides?: Partial<ChineseTenResponse>): ChineseTenResponse {
  return {
    players: [human, cpu],
    layout: [card('SPADE', 9), card('DIAMOND', 3, 3)],
    phase: 0,
    currentPlayerIdx: 0,
    stockCount: 20,
    selectableIndices: [],
    tieScore: 105,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatChineseTenState', () => {
  it('prints both capture rules every frame', () => {
    // Terminal output scrolls; a rule shown once is a rule nobody sees later.
    expect(formatChineseTenState(makeState())).toContain('A-9 sum to ten, 10-K by rank');
  });

  it('annotates only the cards that score', () => {
    const out = formatChineseTenState(makeState());
    expect(out).toContain('(3)'); // the red three
    expect(out).toMatch(/layout: 0:[^(]+ 1:/); // the black nine carries no value
  });

  it('prints captures for BOTH seats but the hand only for the visible one', () => {
    const out = formatChineseTenState(makeState());
    expect(out.match(/captured:/g)).toHaveLength(2);
    expect(out).toContain('2 cards'); // the hidden hand is a count
  });

  it('shows the pending card while a choice is open', () => {
    const out = formatChineseTenState(makeState({ phase: 1, pendingCard: card('SPADE', 1) }));
    expect(out).toContain('choose a layout card');
  });

  it('reports each ending', () => {
    expect(formatChineseTenState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('you win');
    expect(formatChineseTenState(makeState({ gameEndFlag: true, winnerIdx: 1 }))).toContain('you lose');
    expect(formatChineseTenState(makeState({ gameEndFlag: true, winnerIdx: -1 }))).toContain('draw');
  });

  it('renders an empty layout rather than nothing', () => {
    expect(formatChineseTenState(makeState({ layout: [] }))).toContain('layout: -');
  });
});
