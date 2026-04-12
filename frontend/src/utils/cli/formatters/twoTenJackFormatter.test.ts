import { describe, expect, it } from 'vitest';
import type { TwoTenJackResponse } from '../../../types/card';
import { formatTwoTenJackState } from './twoTenJackFormatter';

function makeState(overrides?: Partial<TwoTenJackResponse>): TwoTenJackResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 13,
        cards: [{ design: 'SPADE', value: 1 }],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        capturedPoints: 0,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        capturedPoints: 0,
      },
      {
        id: 2,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        capturedPoints: 0,
      },
      {
        id: 3,
        isHuman: false,
        cardCount: 13,
        cards: [],
        roundScore: 0,
        cumulativeScore: 0,
        trickCount: 0,
        capturedPoints: 0,
      },
    ],
    phase: 1,
    roundNumber: 1,
    trickNumber: 2,
    currentPlayerIdx: 0,
    declarerIdx: 0,
    trumpSuit: 1,
    currentTrick: [],
    gameEndFlag: false,
    winnerTeam: -1,
    leadPlayerIdx: 0,
    message: '',
    config: { cpuDifficulty: 1, pointLimit: 50 },
    ...overrides,
  };
}

describe('formatTwoTenJackState', () => {
  it('formats basic state with header and round/trick info', () => {
    const output = formatTwoTenJackState(makeState());
    expect(output).toContain('Two Ten Jack');
    expect(output).toContain('round: 1');
    expect(output).toContain('trick: 2');
  });

  it('shows human player hand cards', () => {
    const output = formatTwoTenJackState(makeState());
    expect(output).toContain('[0]');
  });

  it('shows current trick when cards are present', () => {
    const output = formatTwoTenJackState(
      makeState({
        currentTrick: [{ playerIdx: 1, card: { design: 'CLOVER', value: 3 } }],
      }),
    );
    expect(output).toContain('trick:');
  });

  it('shows message when present', () => {
    const output = formatTwoTenJackState(makeState({ message: 'Some message' }));
    expect(output).toContain('Some message');
  });

  it('shows Game Over when gameEndFlag is true', () => {
    const output = formatTwoTenJackState(makeState({ gameEndFlag: true }));
    expect(output).toContain('Game Over');
  });

  it('does not show trick section when currentTrick is empty', () => {
    const output = formatTwoTenJackState(makeState({ currentTrick: [] }));
    // Header always contains "trick: <num>"; the trick card section uses "trick: <name>=<card>".
    expect(output).not.toMatch(/trick: [^\d]/);
  });
});
