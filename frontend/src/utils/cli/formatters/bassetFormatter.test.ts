import { describe, expect, it } from 'vitest';
import { formatBassetState } from './bassetFormatter';

describe('formatBassetState', () => {
  it('shows the paroli decision', () => {
    const text = formatBassetState({
      phase: 3,
      chips: 990,
      bet: { rank: 7, amount: 10, stage: 1 },
      bankerCard: null,
      playerCard: null,
      hit: true,
      turnsPlayed: 1,
      turnsTotal: 26,
      remaining: 50,
      remainingByRank: [],
      totalPayout: 0,
      gameEndFlag: false,
      message: '',
      messageCode: '',
    });
    expect(text).toContain('decision: take or paroli');
  });
});
