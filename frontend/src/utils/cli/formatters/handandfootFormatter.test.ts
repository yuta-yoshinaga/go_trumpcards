import { describe, expect, it } from 'vitest';
import type { HandAndFootResponse } from '../../../types/card';
import { formatHandAndFootState } from './handandfootFormatter';

const baseState: HandAndFootResponse = {
  players: [
    {
      id: 0,
      team: 0,
      isHuman: true,
      cardCount: 13,
      cards: [
        { design: 'SPADE', value: 7 },
        { design: 'HEART', value: 10 },
      ],
      footCount: 11,
      inFoot: false,
      roundScore: 0,
      cumulativeScore: 0,
    },
    {
      id: 1,
      team: 1,
      isHuman: false,
      cardCount: 13,
      cards: [],
      footCount: 11,
      inFoot: true,
      roundScore: 0,
      cumulativeScore: 0,
    },
  ],
  teams: [
    {
      team: 0,
      melds: [
        {
          cards: [
            { design: 'SPADE', value: 7 },
            { design: 'HEART', value: 7 },
            { design: 'CLOVER', value: 7 },
          ],
          isNatural: true,
          isCanasta: false,
          rank: 7,
        },
      ],
      red3Count: 1,
      red3s: [{ design: 'HEART', value: 3 }],
    },
    { team: 1, melds: [], red3Count: 0, red3s: [] },
  ],
  phase: 1,
  roundNumber: 2,
  currentPlayerIdx: 0,
  discardTop: { design: 'SPADE', value: 5 },
  drawPileCount: 67,
  discardPileCount: 3,
  isFrozen: false,
  gameEndFlag: false,
  winnerTeam: -1,
  message: '',
  messageCode: '',
  config: { cpuDifficulty: 1, pointLimit: 5000 },
};

describe('formatHandAndFootState', () => {
  it('renders header, round, phase, and team melds', () => {
    const out = formatHandAndFootState(baseState);
    expect(out).toContain('Hand and Foot');
    expect(out).toContain('round: 2');
    expect(out).toContain('MELD');
    expect(out).toContain('Team 0');
    expect(out).toContain('meld:');
    expect(out).toContain('In Foot');
  });

  it('shows frozen indicator and winner at game end', () => {
    const out = formatHandAndFootState({
      ...baseState,
      isFrozen: true,
      gameEndFlag: true,
      phase: 4,
      winnerTeam: 0,
    });
    expect(out).toContain('[FROZEN]');
    expect(out).toContain('Game Over! Winner: Team 0');
  });

  it('handles empty discard top', () => {
    const out = formatHandAndFootState({ ...baseState, discardTop: null });
    expect(out).toContain('[  ]');
  });
});
