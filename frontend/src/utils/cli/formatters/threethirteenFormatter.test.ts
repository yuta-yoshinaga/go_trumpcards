import { describe, expect, it } from 'vitest';
import type { ThreeThirteenResponse } from '../../../types/card';
import { formatThreeThirteenState } from './threethirteenFormatter';

const baseState: ThreeThirteenResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 4,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 11 },
      ],
      deadwood: 17,
      roundScore: 12,
      cumulativeScore: 30,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 4,
      cards: [],
      deadwood: 5,
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
  phase: 1,
  round: 2,
  wildRank: 4,
  dealCount: 4,
  currentPlayerIdx: 0,
  knockerIdx: -1,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 24,
  gameEndFlag: false,
  winnerIdx: -1,
  message: '',
  messageCode: '',
  config: { cpuDifficulty: 1, playerCount: 2 },
};

describe('formatThreeThirteenState', () => {
  it('renders header, round, wild, phase, and indexed human hand', () => {
    const out = formatThreeThirteenState(baseState);
    expect(out).toContain('Three Thirteen');
    expect(out).toContain('round: 2/11');
    expect(out).toContain('wild: 4');
    expect(out).toContain('DISCARD');
    expect(out).toContain('total=30');
    expect(out).toContain('stock: 24');
  });

  it('renders knocker line when someone knocked', () => {
    const out = formatThreeThirteenState({ ...baseState, knockerIdx: 0 });
    expect(out).toContain('knocked!');
  });

  it('shows winner at game end', () => {
    const out = formatThreeThirteenState({ ...baseState, gameEndFlag: true, phase: 3, winnerIdx: 0 });
    expect(out).toContain('Game Over! Winner:');
  });

  it('handles empty discard top', () => {
    const out = formatThreeThirteenState({ ...baseState, discardTop: null });
    expect(out).toContain('[  ]');
  });
});
