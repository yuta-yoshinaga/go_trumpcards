import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { canfieldAutoCompleteReady } from './canfieldAutoComplete';

const c = (value: number): Card => ({ design: 'SPADE', value });
const base = { reserve: [] as Card[], stockCount: 0, waste: [] as Card[] };

describe('canfieldAutoCompleteReady', () => {
  it('is ready once reserve, stock and waste are all empty', () => {
    expect(canfieldAutoCompleteReady(base)).toBe(true);
  });

  // ドメインの AutoComplete が弾く3条件。1つずつ踏まないと、
  // 「常に true」の実装でも通ってしまう。
  it.each([
    ['reserve', { ...base, reserve: [c(5)] }],
    ['stock', { ...base, stockCount: 1 }],
    ['waste', { ...base, waste: [c(5)] }],
  ])('is not ready while the %s still holds a card', (_name, state) => {
    expect(canfieldAutoCompleteReady(state)).toBe(false);
  });
});
