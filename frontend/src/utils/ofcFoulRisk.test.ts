import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { ofcPlacementFouls, ofcRowsAlreadyFouled } from './ofcFoulRisk';

const card = (design: CardDesign, value: number): Card => ({ design, value });

/** ♠A ♥A ♣A: フロントに置くと最強クラスの3枚。 */
const tripAces = [card('SPADE', 1), card('HEART', 1), card('CLOVER', 1)];
/** ばらばらの5枚 (ハイカード)。 */
const junkMiddle = [card('SPADE', 9), card('HEART', 7), card('CLOVER', 5), card('DIAMOND', 4), card('SPADE', 3)];
/** junkMiddle より強いハイカード5枚。 */
const strongBack = [card('HEART', 13), card('CLOVER', 11), card('DIAMOND', 8), card('SPADE', 6), card('HEART', 2)];

describe('ofcRowsAlreadyFouled', () => {
  it('does not flag rows that are still filling up', () => {
    expect(
      ofcRowsAlreadyFouled({
        front: [card('SPADE', 1)],
        middle: [card('HEART', 2)],
        back: [],
      }),
    ).toBe(false);
  });

  // **埋まった2段だけで確定する。**3段そろうのを待つと、取り返しがつかなく
  // なってから知らせることになる。
  it('flags a full front that already beats a full middle', () => {
    expect(ofcRowsAlreadyFouled({ front: tripAces, middle: junkMiddle, back: [] })).toBe(true);
  });

  it('flags a full middle that already beats a full back', () => {
    expect(ofcRowsAlreadyFouled({ front: [], middle: strongBack, back: junkMiddle })).toBe(true);
  });

  it('accepts a legal pair of full rows', () => {
    expect(ofcRowsAlreadyFouled({ front: [], middle: junkMiddle, back: strongBack })).toBe(false);
  });

  // **空きのある段のせいで警告してはいけない。**あとで引く札で役は変わる。
  it('does not blame a row that still has room', () => {
    expect(
      ofcRowsAlreadyFouled({
        front: [],
        middle: junkMiddle,
        back: [card('HEART', 2), card('CLOVER', 3)],
      }),
    ).toBe(false);
  });

  it('flags a complete arrangement that fouls', () => {
    expect(ofcRowsAlreadyFouled({ front: tripAces, middle: junkMiddle, back: strongBack })).toBe(true);
  });

  it('accepts a complete arrangement that does not foul', () => {
    expect(
      ofcRowsAlreadyFouled({
        front: [card('SPADE', 2), card('HEART', 4), card('CLOVER', 6)],
        middle: junkMiddle,
        back: strongBack,
      }),
    ).toBe(false);
  });
});

describe('ofcPlacementFouls', () => {
  // フロントに ♣A を置くと ♠A ♥A ♣A のスリーカードになり、ハイカードのミドルを
  // 上回って反則が確定する。
  it('warns about the card that completes a fouling front row', () => {
    const rows = { front: [card('SPADE', 1), card('HEART', 1)], middle: junkMiddle, back: strongBack };
    expect(ofcPlacementFouls(rows, card('CLOVER', 1), 'front')).toBe(true);
  });

  // **同じ札でも置く段が違えば警告しない。**行ごとに判定していない実装だと
  // ここで落ちる。
  it('does not warn when the same card goes somewhere harmless', () => {
    const rows = {
      front: [card('SPADE', 1), card('HEART', 1)],
      middle: junkMiddle,
      back: [card('HEART', 13), card('CLOVER', 11), card('DIAMOND', 8), card('SPADE', 6)],
    };
    expect(ofcPlacementFouls(rows, card('CLOVER', 1), 'back')).toBe(false);
  });

  it('does not warn while the rows are still incomplete', () => {
    const rows = { front: [card('SPADE', 1)], middle: [], back: [] };
    expect(ofcPlacementFouls(rows, card('HEART', 1), 'front')).toBe(false);
  });

  // **置く前の状態は変えない。**警告のたびに手札が増えたら表示が壊れる。
  it('leaves the caller rows untouched', () => {
    const front = [card('SPADE', 1), card('HEART', 1)];
    const rows = { front, middle: junkMiddle, back: strongBack };
    ofcPlacementFouls(rows, card('CLOVER', 1), 'front');
    expect(front).toHaveLength(2);
  });
});
