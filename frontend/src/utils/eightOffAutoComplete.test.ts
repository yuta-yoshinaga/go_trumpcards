import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { eightOffAutoCompleteReady } from './eightOffAutoComplete';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const emptyFoundation = (): Card[][] => [[], [], [], []];
const noCells = (): (Card | null)[] => [null, null, null, null, null, null, null, null];

describe('eightOffAutoCompleteReady', () => {
  it('is false when nothing can move', () => {
    expect(eightOffAutoCompleteReady(noCells(), [[card('SPADE', 7)]], emptyFoundation())).toBe(false);
  });

  // これが「姉妹ゲームのしきい値を写した実装」を落とすケース。
  // BeleagueredCastle は Reset で A を4枚積んでから配るので `length > 1` が成り立つが、
  // EightOff の組札は空で始まる。配った直後にタブロー末尾へ A があれば、掃き出しは動く。
  it('is true for an ace on a tableau end with every foundation still empty', () => {
    expect(eightOffAutoCompleteReady(noCells(), [[card('SPADE', 7), card('HEART', 1)]], emptyFoundation())).toBe(true);
  });

  it('is true for an ace parked in a free cell with every foundation still empty', () => {
    const cells = noCells();
    cells[3] = card('CLOVER', 1);
    expect(eightOffAutoCompleteReady(cells, [[card('SPADE', 7)]], emptyFoundation())).toBe(true);
  });

  it('is true when a card continues its own suit', () => {
    const foundation = emptyFoundation();
    foundation[2] = [card('HEART', 1)];
    expect(eightOffAutoCompleteReady(noCells(), [[card('HEART', 2)]], foundation)).toBe(true);
  });

  it('is false for the right rank in the wrong suit', () => {
    const foundation = emptyFoundation();
    foundation[2] = [card('HEART', 1)];
    expect(eightOffAutoCompleteReady(noCells(), [[card('SPADE', 2)]], foundation)).toBe(false);
  });

  it('is false for a card that skips a rank', () => {
    const foundation = emptyFoundation();
    foundation[2] = [card('HEART', 1)];
    expect(eightOffAutoCompleteReady(noCells(), [[card('HEART', 3)]], foundation)).toBe(false);
  });

  // 末尾だけが動かせる。埋まっている札は掃き出しの対象ではない。
  it('ignores a placeable card buried under the column end', () => {
    expect(eightOffAutoCompleteReady(noCells(), [[card('HEART', 1), card('SPADE', 7)]], emptyFoundation())).toBe(false);
  });

  it('is false for an empty board', () => {
    expect(eightOffAutoCompleteReady(noCells(), [[], []], emptyFoundation())).toBe(false);
  });
});
