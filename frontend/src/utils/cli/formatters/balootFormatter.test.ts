import { describe, expect, it } from 'vitest';
import type { BalootResponse, Card } from '../../../types/card';
import { formatBalootState } from './balootFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 1)] : [],
  hasBaloot: false,
  declared: false,
  trickCount: 0,
  ...over,
});

const state = (over: Partial<BalootResponse> = {}): BalootResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 1,
    mode: 1,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 0,
    declarerIdx: 0,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 2,
    scores: [40, 20],
    roundPoints: [0, 0],
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 152 },
    message: '',
    ...over,
  }) as unknown as BalootResponse;

describe('formatBalootState', () => {
  it('reports loading for a null state', () => {
    expect(formatBalootState(null)).toBe('Loading...');
  });

  it('shows the round, trick, target and score', () => {
    const out = formatBalootState(state());
    expect(out).toContain('round 2');
    expect(out).toContain('trick 4/8');
    expect(out).toContain('first to 152');
    expect(out).toContain('yours=40');
    expect(out).toContain('theirs=20');
  });

  // **有効な序列だけを出す。** モードで入れ替わるので、両方出すと読めない。
  it('prints the Sun order under Sun', () => {
    const out = formatBalootState(state({ mode: 1 }));
    expect(out).toContain('mode: Sun');
    expect(out).toContain('A=11 > 10 > K=4');
    expect(out).not.toContain('J=20');
  });

  it('prints the Hokom order under Hokom', () => {
    const out = formatBalootState(state({ mode: 2, trumpSuit: 3 }));
    expect(out).toContain('mode: Hokom, trump ♥');
    expect(out).toContain('J=20 > 9=14');
  });

  it('says the mode is undeclared before anyone declares', () => {
    const out = formatBalootState(state({ mode: 0, declarerIdx: -1, phase: 0 }));
    expect(out).toContain('undeclared');
    expect(out).not.toContain('J=20');
  });

  it('marks legal cards in the hand', () => {
    const out = formatBalootState(state({ validPlays: [1] }));
    expect(out).toMatch(/\[1\]\S+\*/);
    expect(out).not.toMatch(/\[0\]\S+\*/);
  });

  it('flags the seat holding Baloot', () => {
    const out = formatBalootState(
      state({ players: [seat(0, { hasBaloot: true }), seat(1), seat(2), seat(3)] } as Partial<BalootResponse>),
    );
    expect(out).toContain('Baloot(K+Q)=20');
    expect(out).toContain('no bonus');
  });

  it('shows the current trick when one is under way', () => {
    const out = formatBalootState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<BalootResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'team 0 wins'],
    [1, 'team 1 wins'],
    [-1, 'tie'],
  ])('reports outcome %s at game end', (winnerTeam, expected) => {
    expect(formatBalootState(state({ gameEndFlag: true, winnerTeam }))).toContain(expected);
  });

  it('appends the server message', () => {
    expect(formatBalootState(state({ message: 'hello' }))).toContain('hello');
  });
});
