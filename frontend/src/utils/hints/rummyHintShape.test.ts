import { describe, expect, it } from 'vitest';
import type { Card } from '../../types/card';
import { aceHighAdjacent, heaviestSpare, isMaterial } from './rummyHintShape';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('aceHighAdjacent', () => {
  it('joins neighbouring ranks', () => {
    expect(aceHighAdjacent(5, 6)).toBe(true);
    expect(aceHighAdjacent(6, 5)).toBe(true);
  });

  it('joins the ace to the king as well as to the two', () => {
    expect(aceHighAdjacent(1, 2)).toBe(true);
    expect(aceHighAdjacent(1, 13)).toBe(true);
  });

  it('does not join ranks two apart', () => {
    expect(aceHighAdjacent(5, 7)).toBe(false);
  });

  it('does not wrap the king round to the two', () => {
    // K-A-2 は同じランの中では不可。A を両端に置けることと、輪になることは別。
    expect(aceHighAdjacent(13, 2)).toBe(false);
  });
});

describe('isMaterial', () => {
  it('matches a card of the same rank in any suit', () => {
    expect(isMaterial(card('SPADE', 9), [card('HEART', 9)], aceHighAdjacent)).toBe(true);
  });

  it('matches a neighbour only in the same suit', () => {
    expect(isMaterial(card('SPADE', 9), [card('SPADE', 10)], aceHighAdjacent)).toBe(true);
    expect(isMaterial(card('SPADE', 9), [card('HEART', 10)], aceHighAdjacent)).toBe(false);
  });

  it('finds nothing in an empty hand', () => {
    expect(isMaterial(card('SPADE', 9), [], aceHighAdjacent)).toBe(false);
  });

  it('takes the adjacency rule from its caller', () => {
    // 隣接を「差が 4」と定義すれば 7 と J が繋がる。Chinchón の 40 枚デッキが
    // これに相当する（あちらは位置に写してから比べる）。
    const gapFour = (a: number, b: number) => Math.abs(a - b) === 4;
    expect(isMaterial(card('SPADE', 7), [card('SPADE', 11)], gapFour)).toBe(true);
    expect(isMaterial(card('SPADE', 7), [card('SPADE', 11)], aceHighAdjacent)).toBe(false);
  });
});

describe('heaviestSpare', () => {
  it('throws the heaviest card that connects with nothing', () => {
    // 5♠6♠ は繋がっている。K♥ と 3♦ が浮いており、重いのは K。
    const hand = [card('SPADE', 5), card('SPADE', 6), card('HEART', 13), card('DIAMOND', 3)];
    expect(heaviestSpare(hand, aceHighAdjacent)).toBe(2);
  });

  it('keeps index 0 as a valid answer', () => {
    // **札 0 も捨て札になりうる。**真偽値で見ると先頭だけ落ちる。
    const hand = [card('HEART', 13), card('SPADE', 5), card('SPADE', 6)];
    expect(heaviestSpare(hand, aceHighAdjacent)).toBe(0);
  });

  it('falls back to the heaviest card when everything is material', () => {
    const hand = [card('SPADE', 5), card('SPADE', 6), card('SPADE', 7)];
    expect(heaviestSpare(hand, aceHighAdjacent)).toBe(2);
  });

  it('never picks a skipped card, even when it is the heaviest', () => {
    // Three Thirteen のワイルドがこれ。捨てると手が弱くなるだけ。
    const hand = [card('HEART', 13), card('SPADE', 4), card('DIAMOND', 9)];
    const idx = heaviestSpare(hand, aceHighAdjacent, (c) => c.value === 13);
    expect(idx).not.toBe(0);
    expect(idx).toBe(2);
  });

  it('returns -1 when skipping leaves nothing', () => {
    // ワイルドしか持っていない手。呼び出し側が -1 を扱う必要がある。
    const hand = [card('HEART', 13), card('SPADE', 13)];
    expect(heaviestSpare(hand, aceHighAdjacent, (c) => c.value === 13)).toBe(-1);
  });

  it('returns -1 for an empty hand', () => {
    expect(heaviestSpare([], aceHighAdjacent)).toBe(-1);
  });
});
