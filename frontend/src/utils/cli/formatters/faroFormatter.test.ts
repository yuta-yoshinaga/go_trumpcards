import { describe, expect, it } from 'vitest';
import type { Card, FaroResponse } from '../../../types/card';
import { formatFaroState } from './faroFormatter';

const card = (design: Card['design'], value: number): Card => ({ design, value });

function makeState(overrides: Partial<FaroResponse> = {}): FaroResponse {
  return {
    phase: 1,
    chips: 1000,
    bets: [],
    soda: null,
    losingCard: null,
    winningCard: null,
    split: false,
    turnsPlayed: 0,
    turnsTotal: 25,
    remaining: 52,
    remainingByRank: Array.from({ length: 14 }, (_, i) => (i === 0 ? 0 : 4)),
    callCards: [],
    callOrder: [],
    callWon: false,
    totalPayout: 0,
    gameEndFlag: false,
    message: '',
    ...overrides,
  };
}

describe('formatFaroState', () => {
  it('renders the header, phase, chips, and a no-bets line', () => {
    const out = formatFaroState(makeState());
    expect(out).toContain('Faro');
    expect(out).toContain('Betting');
    expect(out).toContain('chips: 1000');
    expect(out).toContain('no bets placed');
  });

  it('lists bets, marking coppers', () => {
    const out = formatFaroState(
      makeState({
        bets: [
          { rank: 7, amount: 100, copper: false },
          { rank: 13, amount: 50, copper: true },
        ],
      }),
    );
    expect(out).toContain('rank 7: 100');
    expect(out).toContain('rank 13: 50 (copper)');
  });

  it('renders the last turn with a split note', () => {
    const out = formatFaroState(
      makeState({ phase: 2, losingCard: card('SPADE', 5), winningCard: card('HEART', 5), split: true }),
    );
    expect(out).toContain('losing card:');
    expect(out).toContain('winning card:');
    expect(out).toContain('split (bank takes half)');
  });

  it('shows the call result and round net at round end', () => {
    const won = formatFaroState(makeState({ phase: 4, callOrder: [3, 9, 12], callWon: true, totalPayout: 200 }));
    expect(won).toContain('Call won!');
    expect(won).toContain('round net: 200');
    const lost = formatFaroState(makeState({ phase: 4, callOrder: [3, 9, 12], callWon: false }));
    expect(lost).toContain('Call lost.');
  });

  it('shows the game-over line when out of chips', () => {
    const out = formatFaroState(makeState({ phase: 5, gameEndFlag: true, chips: 0 }));
    expect(out).toContain('Out of chips. Game over.');
  });
});
