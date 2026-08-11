import { describe, expect, it } from 'vitest';
import type { Card, ShelemResponse } from '../../../types/card';
import { formatShelemState } from './shelemFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 1)] : [],
  bid: -1,
  passed: false,
  declaredShelem: false,
  trickCount: 0,
  ...over,
});

const state = (over: Partial<ShelemResponse> = {}): ShelemResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 2,
    roundNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    declarerIdx: 0,
    contract: 130,
    shelemBid: false,
    minBid: 135,
    widowSize: 0,
    discardCount: 4,
    scores: [120, 60],
    roundPoints: [45, 20],
    teamTricks: [3, 1],
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    leadPlayerIdx: 0,
    dealerIdx: 2,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 500 },
    message: '',
    ...over,
  }) as unknown as ShelemResponse;

describe('formatShelemState', () => {
  it('reports loading for a null state', () => {
    expect(formatShelemState(null)).toBe('Loading...');
  });

  it('shows the round, trick, target and trump', () => {
    const out = formatShelemState(state());
    expect(out).toContain('round 2');
    expect(out).toContain('trick 4/12');
    expect(out).toContain('first to 500');
    expect(out).toContain('trump: ♥');
  });

  // **点になるのは A/10/5 だけ。** 盤面から読めないので常時出す。
  it('always states the point cards', () => {
    const out = formatShelemState(state());
    expect(out).toContain('A and 10 are 10');
    expect(out).toContain('exactly 100');
  });

  // 契約は未定・通常・Shelem の3通りを踏む。
  it('shows the contract in each of its shapes', () => {
    expect(formatShelemState(state({ declarerIdx: -1, contract: 0, minBid: 100 }))).toContain('bid at least 100');
    expect(formatShelemState(state())).toContain('contract: 130');
    expect(formatShelemState(state({ shelemBid: true }))).toContain('Shelem');
  });

  it('shows the widow while it is still face down', () => {
    expect(formatShelemState(state({ widowSize: 4 }))).toContain('widow: 4 cards');
    expect(formatShelemState(state({ widowSize: 0 }))).not.toContain('widow:');
  });

  // 競りでの立場が席ごとに出る。5 種すべて踏む。
  it('labels each seat with its bidding standing', () => {
    const out = formatShelemState(
      state({
        declarerIdx: 0,
        players: [seat(0, { bid: 130 }), seat(1, { passed: true }), seat(2, { bid: 120 }), seat(3)],
      } as Partial<ShelemResponse>),
    );
    expect(out).toContain('won 130');
    expect(out).toContain('passed');
    expect(out).toContain('bid 120');
    expect(out).toContain('bidding');

    const shelem = formatShelemState(
      state({
        shelemBid: true,
        players: [seat(0, { declaredShelem: true }), seat(1), seat(2), seat(3)],
      } as Partial<ShelemResponse>),
    );
    expect(shelem).toContain('declared Shelem');
  });

  it('marks legal cards in the hand', () => {
    const out = formatShelemState(state({ validPlays: [1] }));
    expect(out).toMatch(/\[1\]\S+\*/);
    expect(out).not.toMatch(/\[0\]\S+\*/);
  });

  it('shows the current trick when one is under way', () => {
    const out = formatShelemState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<ShelemResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'team 0 wins'],
    [-1, 'tie'],
  ])('reports outcome %s at game end', (winnerTeam, expected) => {
    expect(formatShelemState(state({ gameEndFlag: true, winnerTeam }))).toContain(expected);
  });

  it('appends the server message', () => {
    expect(formatShelemState(state({ message: 'hello' }))).toContain('hello');
  });
});
