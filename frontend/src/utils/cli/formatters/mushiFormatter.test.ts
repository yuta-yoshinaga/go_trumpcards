import { describe, expect, it } from 'vitest';
import type { MushiCard, MushiPlayer, MushiResponse } from '../../../types/card';
import { formatMushiState } from './mushiFormatter';

const card = (month: number, index: number, category = 0, isWild = false): MushiCard => ({
  // `design`/`value` are the wire's generic card fields; hanafuda identity
  // travels in `month`/`index`, which is what the UI reads.
  design: 'SPADE',
  value: index,
  month,
  index,
  category,
  points: [1, 5, 10, 20][category] ?? 1,
  isWild,
});

const human: MushiPlayer = {
  id: 0,
  isHuman: true,
  cardCount: 1,
  cards: [card(1, 1, 3)],
  captured: [card(3, 2, 1)],
  capturedPoints: 5,
  score: 12,
  roundResult: 12,
  hidden: false,
};

const cpu: MushiPlayer = {
  id: 1,
  isHuman: false,
  cardCount: 2,
  cards: [],
  captured: [card(12, 1, 3)],
  capturedPoints: 20,
  score: -12,
  roundResult: -12,
  hidden: true,
};

function makeState(overrides?: Partial<MushiResponse>): MushiResponse {
  return {
    players: [human, cpu],
    field: [card(5, 3), card(11, 4, 0, true)],
    phase: 0,
    currentPlayerIdx: 0,
    dealerIdx: 0,
    roundNumber: 2,
    targetRounds: 12,
    stockCount: 14,
    selectableIndices: [],
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  };
}

describe('formatMushiState', () => {
  it('indexes the field so the select command can name a card', () => {
    const out = formatMushiState(makeState());
    expect(out).toContain('field: 0:');
  });

  it('marks the lightning card', () => {
    expect(formatMushiState(makeState())).toContain('11-4(chaff)*');
  });

  it('prints captures for BOTH seats but the hand only for the visible one', () => {
    const out = formatMushiState(makeState());
    expect(out.match(/captured:/g)).toHaveLength(2);
    expect(out.match(/hand:/g)).toHaveLength(1);
  });

  it('shows the pending card while a choice is open', () => {
    const out = formatMushiState(makeState({ phase: 2, pendingCard: card(11, 4, 0, true) }));
    expect(out).toContain('choose a field card');
  });

  it('reports each ending', () => {
    expect(formatMushiState(makeState({ gameEndFlag: true, winnerIdx: 0 }))).toContain('you win');
    expect(formatMushiState(makeState({ gameEndFlag: true, winnerIdx: 1 }))).toContain('you lose');
    expect(formatMushiState(makeState({ gameEndFlag: true, winnerIdx: -1 }))).toContain('draw');
  });

  it('renders an empty zone rather than nothing', () => {
    const out = formatMushiState(makeState({ field: [] }));
    expect(out).toContain('field: -');
  });
});
