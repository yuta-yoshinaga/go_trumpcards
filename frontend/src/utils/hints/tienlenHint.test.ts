import { describe, expect, it } from 'vitest';
import type { Card, TienLenResponse } from '../../types/card';
import { getTienLenHint } from './tienlenHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 5), card('SPADE', 9)], ...overrides }: Partial<TienLenResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand, isFinished: false, rank: 0 },
      { id: 1, isHuman: false, cardCount: 5, cards: [], isFinished: false, rank: 0 },
    ],
    currentTurn: 0,
    tableCards: [],
    tablePlayType: 0,
    lastPlayPlayerIdx: -1,
    gameEndFlag: false,
    cpuActions: [],
    humanAction: null,
    config: { cpuDifficulty: 1 },
    ...overrides,
  } as unknown as TienLenResponse;
}

describe('getTienLenHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getTienLenHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it("stays quiet on another seat's turn", () => {
    expect(getTienLenHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('leads the weakest card onto an empty table', () => {
    expect(getTienLenHint(base())?.targetAction).toBe('card-0');
  });

  it('ranks the two above the ace, which is above the king', () => {
    const s = base({ hand: [card('SPADE', 2), card('SPADE', 1), card('SPADE', 13)] });
    expect(getTienLenHint(s)?.targetAction).toBe('card-2');
  });

  it('breaks a tie on suit, not just on value', () => {
    // **同じ値でもスートで順序が決まる。**値だけ見ると同点になり、先頭が返る。
    const s = base({ hand: [card('HEART', 7), card('SPADE', 7)] });
    expect(getTienLenHint(s)?.targetAction).toBe('card-1');
  });

  it("uses this game's suit order, which is the reverse of Big Two's", () => {
    // SPADE が最弱、HEART が最強。逆向きの表を写すとここが入れ替わる。
    const s = base({ hand: [card('SPADE', 3), card('HEART', 3)] });
    expect(getTienLenHint(s)?.targetAction).toBe('card-0');
  });

  it('plays the weakest card that still beats the table', () => {
    const s = base({ hand: [card('SPADE', 5), card('SPADE', 9), card('SPADE', 11)], tableCards: [card('SPADE', 7)] });
    expect(getTienLenHint(s)?.targetAction).toBe('card-1');
  });

  it('passes when nothing in hand beats the table', () => {
    const s = base({ hand: [card('SPADE', 5), card('SPADE', 6)], tableCards: [card('HEART', 2)] });
    expect(getTienLenHint(s)?.targetAction).toBe('pass');
  });

  it('measures a multi-card table by its strongest card', () => {
    // 組の強さは一番強い札で決まる。最初の札で見ると弱く見積もる。
    const s = base({ hand: [card('SPADE', 8)], tableCards: [card('SPADE', 4), card('SPADE', 10)] });
    expect(getTienLenHint(s)?.targetAction).toBe('pass');
  });

  it('stays quiet without a hand', () => {
    expect(getTienLenHint(base({ hand: [] }))).toBeNull();
  });
});
