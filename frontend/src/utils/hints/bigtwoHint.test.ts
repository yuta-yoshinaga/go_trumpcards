import { describe, expect, it } from 'vitest';
import type { BigTwoResponse, Card } from '../../types/card';
import { getBigTwoHint } from './bigtwoHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('DIAMOND', 5), card('DIAMOND', 9)], ...overrides }: Partial<BigTwoResponse> & Extra = {}) {
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
  } as unknown as BigTwoResponse;
}

describe('getBigTwoHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getBigTwoHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it("stays quiet on another seat's turn", () => {
    expect(getBigTwoHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('leads the weakest card onto an empty table', () => {
    expect(getBigTwoHint(base())?.targetAction).toBe('card-0');
  });

  it('ranks the two above the ace, which is above the king', () => {
    const s = base({ hand: [card('DIAMOND', 2), card('DIAMOND', 1), card('DIAMOND', 13)] });
    expect(getBigTwoHint(s)?.targetAction).toBe('card-2');
  });

  it('breaks a tie on suit, not just on value', () => {
    // **同じ値でもスートで順序が決まる。**値だけ見ると同点になり、先頭が返る。
    const s = base({ hand: [card('SPADE', 7), card('DIAMOND', 7)] });
    expect(getBigTwoHint(s)?.targetAction).toBe('card-1');
  });

  it("uses this game's suit order, which is the reverse of Tiến Lên's", () => {
    // DIAMOND が最弱、SPADE が最強。逆向きの表を写すとここが入れ替わる。
    const s = base({ hand: [card('DIAMOND', 3), card('SPADE', 3)] });
    expect(getBigTwoHint(s)?.targetAction).toBe('card-0');
  });

  it('plays the weakest card that still beats the table', () => {
    const s = base({
      hand: [card('DIAMOND', 5), card('DIAMOND', 9), card('DIAMOND', 11)],
      tableCards: [card('DIAMOND', 7)],
    });
    expect(getBigTwoHint(s)?.targetAction).toBe('card-1');
  });

  it('passes when nothing in hand beats the table', () => {
    const s = base({ hand: [card('DIAMOND', 5), card('DIAMOND', 6)], tableCards: [card('SPADE', 2)] });
    expect(getBigTwoHint(s)?.targetAction).toBe('pass');
  });

  it('measures a multi-card table by its strongest card', () => {
    // 組の強さは一番強い札で決まる。最初の札で見ると弱く見積もる。
    const s = base({ hand: [card('DIAMOND', 8)], tableCards: [card('DIAMOND', 4), card('DIAMOND', 10)] });
    expect(getBigTwoHint(s)?.targetAction).toBe('pass');
  });

  it('stays quiet without a hand', () => {
    expect(getBigTwoHint(base({ hand: [] }))).toBeNull();
  });
});
