import { describe, expect, it } from 'vitest';
import type { Card, PolignacResponse } from '../../../types/card';
import { formatPolignacState } from './polignacFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('SPADE', 11), card('HEART', 10)] : [],
  score: 0,
  roundPenalty: 0,
  trickCount: 0,
  declaredCapot: false,
  ...over,
});

function makeState(overrides: Partial<PolignacResponse> = {}): PolignacResponse {
  return {
    players: [seat(0, { score: 2, roundPenalty: 2 }), seat(1), seat(2), seat(3)],
    phase: 1,
    roundNumber: 2,
    trickNumber: 3,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    capotIdx: -1,
    capotTricks: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as PolignacResponse;
}

describe('formatPolignacState', () => {
  it('shows a loading line for a null state', () => {
    expect(formatPolignacState(null)).toBe('Loading...');
  });

  it('renders the round, trick and the penalty rule', () => {
    const out = formatPolignacState(makeState());
    expect(out).toContain('round 2/4');
    expect(out).toContain('trick 4/8');
    // 盤面から読み取れない規則なので必ず出す。
    expect(out).toContain('the J of spades costs 2');
  });

  // capot 宣言中だけ警告が出る。負のコントロール付き。
  it('announces a declared capot with its progress', () => {
    const out = formatPolignacState(
      makeState({
        capotIdx: 2,
        capotTricks: 3,
        players: [seat(0), seat(1), seat(2, { declaredCapot: true }), seat(3)],
      }),
    );
    expect(out).toContain('declared capot (3/8)');
    expect(out).toContain('[capot]');
  });

  it('says nothing about capot when nobody declared', () => {
    const out = formatPolignacState(makeState());
    expect(out).not.toContain('capot');
  });

  it('reports both the running total and the round penalty', () => {
    expect(formatPolignacState(makeState())).toContain('2 pts (round 2)');
  });

  it('marks legal cards and leaves illegal ones unmarked', () => {
    const handLine = formatPolignacState(makeState({ validPlays: [0] }))
      .split('\n')
      .find((l) => l.startsWith('your hand:'));
    expect(handLine?.match(/\*/g)).toHaveLength(1);
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatPolignacState(
      makeState({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 11) }] } as Partial<PolignacResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'winner'],
    [-1, 'tie'],
  ])('renders the game-over line for winnerIdx %i', (winnerIdx, expected) => {
    const out = formatPolignacState(makeState({ gameEndFlag: true, winnerIdx, phase: 3 }));
    expect(out).toContain('game over');
    expect(out).toContain(expected);
  });

  it('appends a server message when present', () => {
    expect(formatPolignacState(makeState({ message: 'must follow suit' }))).toContain('must follow suit');
  });
});
