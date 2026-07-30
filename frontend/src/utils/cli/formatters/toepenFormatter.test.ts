import { describe, expect, it } from 'vitest';
import type { Card, CardDesign, ToepenPlayer, ToepenResponse } from '../../../types/card';
import { formatToepenState } from './toepenFormatter';

const card = (design: CardDesign, value: number): Card => ({ design, value });

const human: ToepenPlayer = {
  id: 0,
  isHuman: true,
  cardCount: 2,
  cards: [card('SPADE', 10), card('HEART', 11)],
  lives: 3,
  folded: false,
  eliminated: false,
  hidden: false,
};

const cpu: ToepenPlayer = {
  id: 1,
  isHuman: false,
  cardCount: 2,
  cards: [],
  lives: 9,
  folded: true,
  eliminated: false,
  hidden: true,
};

function makeState(overrides?: Partial<ToepenResponse>): ToepenResponse {
  return {
    players: [human, cpu],
    phase: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    leadSuit: -1,
    trickNumber: 1,
    handNumber: 2,
    stake: 2,
    knockerIdx: -1,
    pendingRespondent: -1,
    lastTrickWinner: -1,
    maxLives: 10,
    validPlayIndices: [0, 1],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatToepenState', () => {
  it('prints the inverted ranking every time', () => {
    // Terminal output scrolls; a ranking shown once at the top is a ranking
    // nobody sees by trick three.
    expect(formatToepenState(makeState())).toContain('10 > 9 > 8 > 7 > A > K > Q > J');
  });

  it('indexes the visible hand and draws a hidden one from its count', () => {
    const out = formatToepenState(makeState());
    expect(out).toContain('0:');
    expect(out).toContain('2 cards');
  });

  it('marks folded and eliminated seats', () => {
    expect(formatToepenState(makeState())).toContain('[folded]');
    const out = formatToepenState(makeState({ players: [human, { ...cpu, folded: false, eliminated: true }] }));
    expect(out).toContain('[out]');
  });

  it('announces an outstanding toep with what it costs to answer', () => {
    const out = formatToepenState(makeState({ knockerIdx: 1, stake: 3 }));
    expect(out).toContain('toep by seat 1');
    expect(out).toContain('s to stay, f to fold');
  });

  it('shows the trick while one is open', () => {
    const out = formatToepenState(makeState({ currentTrick: [{ playerIdx: 0, card: card('SPADE', 10) }] }));
    expect(out).toContain('trick:');
  });

  it('reports each ending', () => {
    expect(formatToepenState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('you win');
    expect(formatToepenState(makeState({ gameEndFlag: true, winnerIdx: 2 }))).toContain('you lose');
  });
});
