import { describe, expect, it } from 'vitest';
import type { DurakPlayerData, DurakResponse } from '../../../types/card';
import { formatDurakState } from './durakFormatter';

const player = (over: Partial<DurakPlayerData> = {}): DurakPlayerData => ({
  id: 0,
  isHuman: true,
  isFinished: false,
  cardCount: 2,
  cards: [
    { design: 'SPADE', value: 5 },
    { design: 'HEART', value: 13 },
  ],
  ...over,
});

const baseState: DurakResponse = {
  message: '',
  players: [player(), player({ id: 1, isHuman: false, cardCount: 3, cards: [] })],
  currentTurn: 0,
  phase: 1,
  attackerIdx: 1,
  defenderIdx: 0,
  tablePairs: [],
  trumpSuit: 'SPADE',
  trumpCard: { design: 'SPADE', value: 6 },
  stockCount: 12,
  loserIdx: -1,
  gameEndFlag: false,
  config: { playerCount: 2, cpuDifficulty: 1, transferEnabled: false },
  cpuActions: [],
  humanAction: null,
  boutNumber: 1,
  sortMode: 0,
};

describe('formatDurakState', () => {
  it('returns a loading placeholder for null state', () => {
    expect(formatDurakState(null)).toBe('Loading...');
  });

  it('renders trump, stock, and roles', () => {
    const out = formatDurakState(baseState);
    expect(out).toContain('Durak');
    expect(out).toContain('trump: ♠6');
    expect(out).toContain('stock: 12');
    expect(out).toContain('attacker:');
    expect(out).toContain('defender:');
  });

  it('shows an empty table and the indexed human hand', () => {
    const out = formatDurakState(baseState);
    expect(out).toContain('table: (empty)');
    expect(out).toContain('[0]♠5');
    expect(out).toContain('[1]♥K');
  });

  it('renders table pairs with undefended and defended entries', () => {
    const out = formatDurakState({
      ...baseState,
      tablePairs: [
        { attack: { design: 'CLOVER', value: 7 }, defense: null },
        { attack: { design: 'DIAMOND', value: 8 }, defense: { design: 'DIAMOND', value: 10 } },
      ],
    });
    expect(out).toContain('[0] ♣7 -> (undefended)');
    expect(out).toContain('[1] ♦8 -> ♦10');
  });

  it('shows opponent card counts and finished status', () => {
    const out = formatDurakState({
      ...baseState,
      players: [player(), player({ id: 1, isHuman: false, cardCount: 0, isFinished: true, cards: [] })],
    });
    expect(out).toContain('0 cards (out)');
  });

  it('reports the loser when the game ends', () => {
    const out = formatDurakState({ ...baseState, gameEndFlag: true, loserIdx: 0 });
    expect(out).toContain('loser (durak):');
  });

  it('reports a draw when the game ends with no loser', () => {
    const out = formatDurakState({ ...baseState, gameEndFlag: true, loserIdx: -1 });
    expect(out).toContain('draw');
  });

  it('appends a server message when present', () => {
    const out = formatDurakState({ ...baseState, message: 'Your turn to attack' });
    expect(out).toContain('Your turn to attack');
  });
});
