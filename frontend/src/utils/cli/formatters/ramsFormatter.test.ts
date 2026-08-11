import { describe, expect, it } from 'vitest';
import type { Card, RamsResponse } from '../../../types/card';
import { formatRamsState } from './ramsFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 1), card('SPADE', 9)] : [],
  chips: 57,
  inRound: false,
  decided: false,
  roundTricks: 0,
  trickCount: 0,
  ...over,
});

function makeState(overrides: Partial<RamsResponse> = {}): RamsResponse {
  return {
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    roundNumber: 2,
    trickNumber: 1,
    pot: 12,
    trumpSuit: 3,
    upCard: card('HEART', 9),
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 0,
    activeCount: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4, rounds: 4 },
    message: '',
    ...overrides,
  } as unknown as RamsResponse;
}

describe('formatRamsState', () => {
  it('shows a loading line for a null state', () => {
    expect(formatRamsState(null)).toBe('Loading...');
  });

  it('renders the round, trick, table size, pot, trump and the risk', () => {
    const out = formatRamsState(makeState());
    expect(out).toContain('round 2/4');
    expect(out).toContain('trick 2/5');
    // **人数は可変なので必ず出す。**
    expect(out).toContain('4 players');
    expect(out).toContain('pot: 12 chips');
    expect(out).toContain('trump:');
    expect(out).toContain('pay 5 more');
  });

  // 5人卓でもそのまま出る。
  it('reports a five-player table', () => {
    const out = formatRamsState(
      makeState({ config: { playerCnt: 5, rounds: 4 }, players: [seat(0), seat(1), seat(2), seat(3), seat(4)] }),
    );
    expect(out).toContain('5 players');
  });

  // 参加状況の3通りをすべて踏む。
  it('renders in / out / undecided per seat', () => {
    const out = formatRamsState(
      makeState({
        players: [
          seat(0, { decided: true, inRound: true, roundTricks: 2 }),
          seat(1, { decided: true, inRound: false }),
          seat(2),
          seat(3),
        ],
      } as Partial<RamsResponse>),
    );
    expect(out).toContain('[in]');
    expect(out).toContain('[out]');
    expect(out).toContain('[-]');
  });

  it('marks legal cards and leaves illegal ones unmarked', () => {
    const handLine = formatRamsState(makeState({ validPlays: [0] }))
      .split('\n')
      .find((l) => l.startsWith('your hand:'));
    expect(handLine?.match(/\*/g)).toHaveLength(1);
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatRamsState(
      makeState({ currentTrick: [{ playerIdx: 1, card: card('HEART', 13) }] } as Partial<RamsResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'winner'],
    [-1, 'tie'],
  ])('renders the game-over line for winnerIdx %i', (winnerIdx, expected) => {
    const out = formatRamsState(makeState({ gameEndFlag: true, winnerIdx, phase: 3 }));
    expect(out).toContain('game over');
    expect(out).toContain(expected);
  });

  it('appends a server message when present', () => {
    expect(formatRamsState(makeState({ message: 'must follow suit' }))).toContain('must follow suit');
  });
});
