import { describe, expect, it } from 'vitest';
import type { Card, ReversisResponse } from '../../../types/card';
import { formatReversisState } from './reversisFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 9)] : [],
  chips: 45,
  roundPenalty: 0,
  trickCount: 0,
  tookQuinola: false,
  tookDiamondAce: false,
  ...over,
});

function makeState(overrides: Partial<ReversisResponse> = {}): ReversisResponse {
  return {
    players: [seat(0, { roundPenalty: 7, tookQuinola: true }), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 2,
    trickNumber: 3,
    pool: 25,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as ReversisResponse;
}

describe('formatReversisState', () => {
  it('shows a loading line for a null state', () => {
    expect(formatReversisState(null)).toBe('Loading...');
  });

  it('renders the round, trick, pool and penalty scale', () => {
    const out = formatReversisState(makeState());
    expect(out).toContain('round 2/4');
    expect(out).toContain('trick 4/12');
    // 盤面から読めない2つの情報は常時出す。
    expect(out).toContain('pool: 25 chips');
    expect(out).toContain('A=4 K=3 Q=2 J=1');
  });

  it('renders the marks a seat has taken, and "clean" for one that has none', () => {
    const out = formatReversisState(makeState());
    expect(out).toContain('[J♥]');
    expect(out).toContain('[clean]');
  });

  it('reports chips and the round penalty separately', () => {
    expect(formatReversisState(makeState())).toContain('45 chips | 7 penalty');
  });

  it('marks legal cards and leaves illegal ones unmarked', () => {
    const handLine = formatReversisState(makeState({ validPlays: [0] }))
      .split('\n')
      .find((l) => l.startsWith('your hand:'));
    expect(handLine?.match(/\*/g)).toHaveLength(1);
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatReversisState(
      makeState({ currentTrick: [{ playerIdx: 1, card: card('DIAMOND', 1) }] } as Partial<ReversisResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'winner'],
    [-1, 'tie'],
  ])('renders the game-over line for winnerIdx %i', (winnerIdx, expected) => {
    const out = formatReversisState(makeState({ gameEndFlag: true, winnerIdx, phase: 2 }));
    expect(out).toContain('game over');
    expect(out).toContain(expected);
  });

  it('appends a server message when present', () => {
    expect(formatReversisState(makeState({ message: 'must follow suit' }))).toContain('must follow suit');
  });
});
