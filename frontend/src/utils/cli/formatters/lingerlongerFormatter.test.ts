import { describe, expect, it } from 'vitest';
import type { Card, LingerLongerResponse } from '../../../types/card';
import { formatLingerLongerState } from './lingerlongerFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as unknown as Card;

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  isHuman: id === 0,
  cardCount: 4,
  cards: id === 0 ? [card('SPADE', 9), card('HEART', 10)] : [],
  tricksWon: 0,
  eliminatedAt: 0,
  ...over,
});

const state = (over: Partial<LingerLongerResponse> = {}): LingerLongerResponse =>
  ({
    players: [seat(0), seat(1), seat(2), seat(3)],
    phase: 0,
    validPlays: [0],
    stockSize: 30,
    currentTrick: [],
    currentPlayerIdx: 0,
    leadPlayerIdx: 0,
    trickNumber: 2,
    lastDrawIdx: -1,
    eliminatedCnt: 0,
    discarded: 6,
    gameEndFlag: false,
    winnerIdx: -1,
    config: { playerCnt: 4 },
    message: '',
    ...over,
  }) as unknown as LingerLongerResponse;

describe('formatLingerLongerState', () => {
  it('reports loading for a null state', () => {
    expect(formatLingerLongerState(null)).toBe('Loading...');
  });

  it('shows the trick number, the stock and the rule', () => {
    const out = formatLingerLongerState(state());
    expect(out).toContain('trick 3');
    expect(out).toContain('stock 30');
    expect(out).toContain('PLAY');
    // **取っても得点にならない規則が要。**
    expect(out).toMatch(/one card from the stock/);
  });

  // **山札が尽きた瞬間から局は終わりに向かう。**
  it('announces the empty stock, and only while the game runs', () => {
    expect(formatLingerLongerState(state({ stockSize: 0 }))).toMatch(/nobody can refill/);
    expect(formatLingerLongerState(state())).not.toMatch(/nobody can refill/);
    expect(formatLingerLongerState(state({ stockSize: 0, gameEndFlag: true, winnerIdx: 0 }))).not.toMatch(
      /nobody can refill/,
    );
  });

  it('lists every seat with its hand size and tricks won', () => {
    const out = formatLingerLongerState(state({ players: [seat(0, { cardCount: 7, tricksWon: 3 }), seat(1)] }));
    expect(out).toMatch(/>あなた: 7 cards, won 3/);
    expect(out).toMatch(/ CPU 1: 4 cards, won 0/);
  });

  // **補充した席と脱落した席は盤面に痕跡が残らない。**
  it('marks the last draw and the eliminated seats', () => {
    expect(formatLingerLongerState(state({ lastDrawIdx: 1 }))).toContain('CPU 1[just drew]');
    expect(formatLingerLongerState(state({ players: [seat(0), seat(1, { eliminatedAt: 2 })] }))).toContain(
      'CPU 1[out 2]',
    );
    expect(formatLingerLongerState(state())).not.toContain('[just drew]');
  });

  it('stars the playable cards in your hand', () => {
    expect(formatLingerLongerState(state())).toMatch(/your hand: \[0\]\S+\*\s+\[1\]/);
  });

  it('shows the current trick when there is one', () => {
    const out = formatLingerLongerState(state({ currentTrick: [{ playerIdx: 1, card: card('SPADE', 14) }] }));
    expect(out).toMatch(/trick: CPU 1:/);
  });

  // **脱落しても局は続く。** 空の手札だけでは理由が読めない。
  it('says so when you are out but the game continues', () => {
    const out = formatLingerLongerState(
      state({ players: [seat(0, { cardCount: 0, cards: [], eliminatedAt: 2 }), seat(1)], currentPlayerIdx: 1 }),
    );
    expect(out).toMatch(/you are out/);
    expect(out).toContain('(empty)');
    // 負のコントロール: 在席中は出さない。
    expect(formatLingerLongerState(state())).not.toMatch(/you are out/);
  });

  it('names the winner once the game ends', () => {
    expect(formatLingerLongerState(state({ gameEndFlag: true, phase: 1, winnerIdx: 0 }))).toMatch(
      /game over — あなた held cards longest/,
    );
    expect(formatLingerLongerState(state({ gameEndFlag: true, phase: 1, winnerIdx: 2 }))).toMatch(
      /game over — CPU 2 held cards longest/,
    );
  });

  it('appends the server message', () => {
    expect(formatLingerLongerState(state({ message: 'boom' }))).toContain('boom');
  });
});
