import { describe, expect, it } from 'vitest';
import type { Card, TonkResponse } from '../../types/card';
import { TonkPhase } from '../../types/phases';
import { getTonkHint } from './tonkHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 13), card('HEART', 12)], ...overrides }: Partial<TonkResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand, roundScore: 0, cumulativeScore: 0 },
      { id: 1, isHuman: false, cardCount: 5, cards: [], roundScore: 0, cumulativeScore: 0 },
    ],
    phase: TonkPhase.DISCARD,
    roundNumber: 1,
    currentPlayerIdx: 0,
    discardTop: card('CLOVER', 2),
    drawPileCount: 20,
    gameEndFlag: false,
    winnerIdx: -1,
    knockerIdx: -1,
    knockerMelds: [],
    knockerDeadwood: [],
    opponentMelds: [],
    opponentDeadwood: [],
    isTonk: false,
    isUndercut: false,
    config: { cpuDifficulty: 1, pointLimit: 100 },
    ...overrides,
  } as TonkResponse;
}

describe('getTonkHint', () => {
  it('returns null once the game is over', () => {
    expect(getTonkHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('returns null while a CPU seat is to act', () => {
    expect(getTonkHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('returns null in a phase with nothing to choose', () => {
    expect(getTonkHint(base({ phase: TonkPhase.ROUND_END }))).toBeNull();
  });

  it('takes the discard when it completes a set', () => {
    // 5-5 を持っていて 3 枚目の 5 が出ている。ペアだけでは組にならないので、
    // 3 枚目が来て初めてデッドウッドが減る。
    const state = base({
      phase: TonkPhase.DRAW,
      hand: [card('SPADE', 5), card('HEART', 5), card('CLOVER', 13)],
      discardTop: card('DIAMOND', 5),
    });
    expect(getTonkHint(state)?.targetAction).toBe('takeDiscard');
  });

  it('takes the discard when it completes a run', () => {
    const state = base({
      phase: TonkPhase.DRAW,
      hand: [card('SPADE', 5), card('SPADE', 6), card('CLOVER', 13)],
      discardTop: card('SPADE', 7),
    });
    expect(getTonkHint(state)?.targetAction).toBe('takeDiscard');
  });

  it('does not take a card that only makes a pair', () => {
    // **ペアはメルドではない。**拾ってもデッドウッドは 1 点も減らない。
    const state = base({
      phase: TonkPhase.DRAW,
      hand: [card('SPADE', 5), card('CLOVER', 13)],
      discardTop: card('DIAMOND', 5),
    });
    expect(getTonkHint(state)?.targetAction).toBe('drawStock');
  });

  it('does not take a same-rank-adjacent card of another suit', () => {
    const state = base({
      phase: TonkPhase.DRAW,
      hand: [card('SPADE', 5), card('SPADE', 6), card('CLOVER', 13)],
      discardTop: card('DIAMOND', 7),
    });
    expect(getTonkHint(state)?.targetAction).toBe('drawStock');
  });

  it('draws from the stock when the pile is empty', () => {
    expect(getTonkHint(base({ phase: TonkPhase.DRAW, discardTop: null }))?.targetAction).toBe('drawStock');
  });

  it('offers the knock only once the discarded hand is within the threshold', () => {
    // 3-3-3 のセットに 2 が 1 枚。捨てるべき札を捨てた残りは 0 点。
    const at = base({ hand: [card('SPADE', 3), card('HEART', 3), card('DIAMOND', 3), card('CLOVER', 2)] });
    expect(getTonkHint(at)?.targetAction).toBe('knock');

    // 端札 5 と 2 が残る。どちらを捨てても 5 点を超える…わけではない: 5 を捨てれば 2 点。
    // 閾値ちょうどを見るため、捨てたあとが 5 点になる手を使う。
    const edge = base({
      hand: [card('SPADE', 3), card('HEART', 3), card('DIAMOND', 3), card('CLOVER', 5), card('SPADE', 9)],
    });
    expect(getTonkHint(edge)?.targetAction).toBe('knock');

    // 捨てたあとが 6 点。1 点超えただけでサーバは拒否する。
    const over = base({
      hand: [card('SPADE', 3), card('HEART', 3), card('DIAMOND', 3), card('CLOVER', 6), card('SPADE', 9)],
    });
    expect(getTonkHint(over)?.targetAction).not.toBe('knock');
  });

  it('does not treat a pair as a meld when deciding to knock', () => {
    // 2-2-K-K-Q。ペアを組と数えると 0 点に見えてノックを勧めてしまうが、
    // 実際のデッドウッドは Q を捨てても 24 点で、サーバは ErrInvalidPlay を返す。
    const state = base({
      hand: [card('SPADE', 2), card('HEART', 2), card('DIAMOND', 13), card('CLOVER', 13), card('SPADE', 12)],
    });
    expect(getTonkHint(state)?.targetAction).not.toBe('knock');
  });

  it('discards the card that leaves the least deadwood', () => {
    // 5-5-5 はセット。残る K と 9 のうち重い K (index 3) を捨てると 9 点、
    // 9 を捨てると 10 点。**5 点は超えているのでノックには行かない。**
    const state = base({
      hand: [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5), card('CLOVER', 13), card('SPADE', 9)],
    });
    expect(getTonkHint(state)?.targetAction).toBe('card-3');
  });

  it('scores an ace as a single point', () => {
    // 4-4-4 のセットに A。K を捨てて残る A が 1 点なのでノックできる。
    const state = base({
      hand: [card('SPADE', 4), card('HEART', 4), card('DIAMOND', 4), card('CLOVER', 1), card('SPADE', 13)],
    });
    expect(getTonkHint(state)?.targetAction).toBe('knock');
  });

  it('returns null when the human has no cards', () => {
    expect(getTonkHint(base({ hand: [] }))).toBeNull();
  });
});
