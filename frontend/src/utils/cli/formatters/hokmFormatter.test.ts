import { describe, expect, it } from 'vitest';
import type { Card, HokmResponse } from '../../../types/card';
import { formatHokmState } from './hokmFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  team: id % 2,
  cardCount: 2,
  cards: id === 0 ? [card('HEART', 11), card('SPADE', 1)] : [],
  isHakem: id === 0,
  trickCount: 0,
  ...over,
});

const state = (over: Partial<HokmResponse> = {}): HokmResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 1,
    handNumber: 2,
    trickNumber: 3,
    trumpSuit: 3,
    hakemIdx: 0,
    scores: [1, 2],
    teamTricks: [4, 2],
    tricksToWin: 7,
    lastHandKot: false,
    lastHandWinner: -1,
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    currentTrick: [],
    validPlays: [0],
    gameEndFlag: false,
    winnerTeam: -1,
    config: { target: 7 },
    message: '',
    ...over,
  }) as unknown as HokmResponse;

describe('formatHokmState', () => {
  it('reports loading for a null state', () => {
    expect(formatHokmState(null)).toBe('Loading...');
  });

  it('shows the hand, target and trump', () => {
    const out = formatHokmState(state());
    expect(out).toContain('hand 2');
    expect(out).toContain('first to 7 hands');
    expect(out).toContain('trump: ♥');
  });

  // **13 まで打たないので、進捗はトリック数の競り合いのほうに出る。**
  it('leads with the race to seven', () => {
    const out = formatHokmState(state());
    expect(out).toContain('yours=4');
    expect(out).toContain('theirs=2');
    expect(out).toContain('first to 7 takes the hand');
  });

  it('says trump is undeclared before the hakem chooses', () => {
    expect(formatHokmState(state({ trumpSuit: 0, phase: 0 }))).toContain('undeclared');
  });

  it('marks the hakem', () => {
    expect(formatHokmState(state())).toContain('[hakem]');
  });

  // **Kot は2点。** 何が起きたかを言わないと得点が飛んで見える。両側を踏む。
  it('explains how the hand ended', () => {
    expect(formatHokmState(state({ phase: 2, lastHandWinner: 0, lastHandKot: true }))).toContain('Kot, 2 points');
    const normal = formatHokmState(state({ phase: 2, lastHandWinner: 1, lastHandKot: false }));
    expect(normal).toContain('reached 7 tricks');
    expect(normal).not.toContain('Kot');
  });

  it('marks legal cards in the hand', () => {
    const out = formatHokmState(state({ validPlays: [1] }));
    expect(out).toMatch(/\[1\]\S+\*/);
    expect(out).not.toMatch(/\[0\]\S+\*/);
  });

  it('shows the current trick when one is under way', () => {
    const out = formatHokmState(
      state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 13) }] } as Partial<HokmResponse>),
    );
    expect(out).toContain('trick:');
  });

  it.each([
    [0, 'team 0 wins'],
    [1, 'team 1 wins'],
    [-1, 'tie'],
  ])('reports outcome %s at game end', (winnerTeam, expected) => {
    expect(formatHokmState(state({ gameEndFlag: true, winnerTeam }))).toContain(expected);
  });

  it('appends the server message', () => {
    expect(formatHokmState(state({ message: 'hello' }))).toContain('hello');
  });
});
