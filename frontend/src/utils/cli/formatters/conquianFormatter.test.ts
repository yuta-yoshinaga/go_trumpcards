import { describe, expect, it } from 'vitest';
import type { ConquianResponse } from '../../../types/card';
import { formatConquianState } from './conquianFormatter';

const baseState: ConquianResponse = {
  players: [
    {
      id: 0,
      isHuman: true,
      cardCount: 10,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 10 },
      ],
      melds: [
        {
          cards: [
            { design: 'SPADE', value: 5 },
            { design: 'HEART', value: 5 },
            { design: 'CLOVER', value: 5 },
          ],
        },
      ],
      wins: 1,
    },
    {
      id: 1,
      isHuman: false,
      cardCount: 10,
      cards: [],
      melds: [],
      wins: 0,
    },
  ],
  layoffTargets: [],
  phase: 1,
  roundNumber: 2,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 28,
  gameEndFlag: false,
  winnerIdx: -1,
  roundWinnerIdx: -1,
  tookDiscard: true,
  message: '',
  messageCode: '',
  config: { cpuDifficulty: 1, targetWins: 3 },
};

describe('formatConquianState', () => {
  it('renders header, round, phase, melds, and took-discard note', () => {
    const out = formatConquianState(baseState);
    expect(out).toContain('Conquian');
    expect(out).toContain('round: 2');
    expect(out).toContain('MELD');
    expect(out).toContain('wins=1');
    expect(out).toContain('melds:');
    expect(out).toContain('took discard');
    expect(out).toContain('turn:');
  });

  it('shows winner at game end', () => {
    const out = formatConquianState({
      ...baseState,
      gameEndFlag: true,
      phase: 3,
      winnerIdx: 0,
    });
    expect(out).toContain('Game Over! Winner:');
  });

  it('shows draw at game end when no winner', () => {
    const out = formatConquianState({
      ...baseState,
      gameEndFlag: true,
      phase: 3,
      winnerIdx: -1,
    });
    expect(out).toContain('Game Over! Draw');
  });

  it('handles empty discard top', () => {
    const out = formatConquianState({ ...baseState, discardTop: null });
    expect(out).toContain('[  ]');
  });
});
