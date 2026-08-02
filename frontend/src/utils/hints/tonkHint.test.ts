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

  it('takes the discard when it pairs a card in hand', () => {
    const state = base({ phase: TonkPhase.DRAW, discardTop: card('CLOVER', 13) });
    expect(getTonkHint(state)?.targetAction).toBe('takeDiscard');
  });

  it('takes the discard when it extends a run in the same suit', () => {
    const state = base({
      phase: TonkPhase.DRAW,
      hand: [card('SPADE', 5), card('HEART', 12)],
      discardTop: card('SPADE', 6),
    });
    expect(getTonkHint(state)?.targetAction).toBe('takeDiscard');
  });

  it('does not take a same-rank-adjacent card of another suit', () => {
    // 隣のランクでもスートが違えば繋がらない。ここを value だけで見ると拾ってしまう。
    const state = base({
      phase: TonkPhase.DRAW,
      hand: [card('SPADE', 5), card('HEART', 12)],
      discardTop: card('CLOVER', 6),
    });
    expect(getTonkHint(state)?.targetAction).toBe('drawStock');
  });

  it('draws from the stock when the discard connects with nothing', () => {
    const state = base({ phase: TonkPhase.DRAW, discardTop: card('CLOVER', 2) });
    expect(getTonkHint(state)?.targetAction).toBe('drawStock');
  });

  it('draws from the stock when the pile is empty', () => {
    expect(getTonkHint(base({ phase: TonkPhase.DRAW, discardTop: null }))?.targetAction).toBe('drawStock');
  });

  it('offers the knock only once the loose cards are within the threshold', () => {
    // 2 + 3 = 5 点ちょうど。閾値は「以下」なので通る。
    const at = base({ hand: [card('SPADE', 2), card('HEART', 3)] });
    expect(getTonkHint(at)?.targetAction).toBe('knock');

    // 2 + 4 = 6 点。1 点超えただけでサーバは拒否する。
    const over = base({ hand: [card('SPADE', 2), card('HEART', 4)] });
    expect(over && getTonkHint(over)?.targetAction).not.toBe('knock');
  });

  it('counts a melded card as costing nothing', () => {
    // K が 3 枚。生の合計は 30 点だが、繋がっているので端札は 0 点。
    const state = base({ hand: [card('SPADE', 13), card('HEART', 13), card('CLOVER', 13)] });
    expect(getTonkHint(state)?.targetAction).toBe('knock');
  });

  it('discards the heaviest card that connects with nothing', () => {
    const state = base({ hand: [card('SPADE', 13), card('HEART', 13), card('CLOVER', 9), card('DIAMOND', 12)] });
    // K は対で繋がっている。残る 9 と Q のうち重い Q (index 3) を捨てる。
    expect(getTonkHint(state)?.targetAction).toBe('card-3');
  });

  it('returns null when the human has no cards', () => {
    expect(getTonkHint(base({ hand: [] }))).toBeNull();
  });
});
