import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import { badugiHandName } from './badugiHandName';

const t = (key: string) => i18n.t(`badugi:${key}`);

describe('badugiHandName', () => {
  it('names each hand size', () => {
    expect(badugiHandName(4, t)).toBe('バドゥーギ');
    expect(badugiHandName(3, t)).toBe('3カード');
    expect(badugiHandName(2, t)).toBe('2カード');
    expect(badugiHandName(1, t)).toBe('1カード');
  });

  // 未評価 (0) や範囲外は名前を持たない。空文字を返して呼び出し側が出さない。
  it.each([0, 5, -1])('returns nothing for out-of-range size %i', (size) => {
    expect(badugiHandName(size, t)).toBe('');
  });
});
