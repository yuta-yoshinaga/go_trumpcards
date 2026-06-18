import { describe, expect, it } from 'vitest';
import type { ChinchonResponse } from '../../../types/card';
import { formatChinchonState } from './chinchonFormatter';

const baseState: ChinchonResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 7,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 11 },
      ],
      roundScore: 12,
      cumulativeScore: 30,
      eliminated: false,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 7,
      cards: [],
      roundScore: 0,
      cumulativeScore: 0,
      eliminated: false,
    },
  ],
  phase: 1,
  roundNumber: 2,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 24,
  gameEndFlag: false,
  winnerIdx: -1,
  knockerIdx: -1,
  knockerMelds: [],
  message: '',
  messageCode: '',
  config: { cpuDifficulty: 1, playerCount: 2, knockThreshold: 5, eliminationLimit: 100 },
};

describe('formatChinchonState', () => {
  it('renders header, round, phase, and indexed human hand', () => {
    const out = formatChinchonState(baseState);
    expect(out).toContain('Chinchon');
    expect(out).toContain('round: 2');
    expect(out).toContain('DISCARD');
    expect(out).toContain('total=30');
    expect(out).toContain('stock: 24');
  });

  it('marks eliminated players', () => {
    const out = formatChinchonState({
      ...baseState,
      players: baseState.players.map((p) => (p.id === 1 ? { ...p, eliminated: true } : p)),
    });
    expect(out).toContain('[OUT]');
  });

  it('renders knocker melds when someone knocked', () => {
    const out = formatChinchonState({
      ...baseState,
      knockerIdx: 0,
      knockerMelds: [
        {
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
        },
      ],
    });
    expect(out).toContain('knocked!');
    expect(out).toContain('melds:');
  });

  it('shows winner at game end', () => {
    const out = formatChinchonState({ ...baseState, gameEndFlag: true, phase: 4, winnerIdx: 0 });
    expect(out).toContain('Game Over! Winner:');
  });

  it('handles empty discard top', () => {
    const out = formatChinchonState({ ...baseState, discardTop: null });
    expect(out).toContain('[  ]');
  });
});
