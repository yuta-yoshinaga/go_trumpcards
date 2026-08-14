import { describe, expect, it } from 'vitest';
import type { Card, SlobberhannesResponse } from '../../../types/card';
import { formatSlobberhannesState } from './slobberhannesFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

function makeState(overrides: Partial<SlobberhannesResponse> = {}): SlobberhannesResponse {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        cardCount: 2,
        cards: [card('SPADE', 1), card('HEART', 10)],
        score: -1,
        trickCount: 2,
        tookFirstTrick: true,
        tookLastTrick: false,
        tookQueen: false,
      },
      {
        id: 1,
        isHuman: false,
        cardCount: 2,
        cards: [],
        score: 0,
        trickCount: 1,
        tookFirstTrick: false,
        tookLastTrick: false,
        tookQueen: false,
      },
    ],
    phase: 0,
    roundNumber: 2,
    trickNumber: 3,
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
  } as unknown as SlobberhannesResponse;
}

describe('formatSlobberhannesState', () => {
  it('shows a loading line for a null state', () => {
    expect(formatSlobberhannesState(null)).toBe('Loading...');
  });

  it('renders the round, trick and the no-trump note', () => {
    const out = formatSlobberhannesState(makeState());
    expect(out).toContain('round 2/4');
    expect(out).toContain('trick 4/8');
    expect(out).toContain('no trump');
  });

  // 最初と最後のトリックでだけ警告が出る。中間では出ない（負のコントロール）。
  it.each([
    [0, 'FIRST trick'],
    [7, 'LAST trick'],
  ])('warns on trick index %i', (trickNumber, expected) => {
    expect(formatSlobberhannesState(makeState({ trickNumber }))).toContain(expected);
  });

  it('stays silent on a middle trick', () => {
    const out = formatSlobberhannesState(makeState({ trickNumber: 3 }));
    expect(out).not.toContain('FIRST trick');
    expect(out).not.toContain('LAST trick');
  });

  it('renders the penalty marks a seat has taken, and "clean" for one that has none', () => {
    const out = formatSlobberhannesState(makeState());
    expect(out).toContain('[1st]');
    expect(out).toContain('[clean]');
  });

  it('marks legal cards and leaves illegal ones unmarked', () => {
    const handLine = formatSlobberhannesState(makeState({ validPlays: [0] }))
      .split('\n')
      .find((l) => l.startsWith('your hand:'));
    expect(handLine?.match(/\*/g)).toHaveLength(1);
  });

  it('renders the current trick when cards are on the table', () => {
    const out = formatSlobberhannesState(
      makeState({ currentTrick: [{ playerIdx: 1, card: card('CLOVER', 12) }] } as Partial<SlobberhannesResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'winner'],
    [-1, 'tie'],
  ])('renders the game-over line for winnerIdx %i', (winnerIdx, expected) => {
    const out = formatSlobberhannesState(makeState({ gameEndFlag: true, winnerIdx, phase: 2 }));
    expect(out).toContain('game over');
    expect(out).toContain(expected);
  });

  it('appends a server message when present', () => {
    expect(formatSlobberhannesState(makeState({ message: 'must follow suit' }))).toContain('must follow suit');
  });
});
