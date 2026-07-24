import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { groupCapturedByCategory, KOIKOI_CATEGORY_ORDER, koikoiCategory } from './koikoiCategory';

/** Builds a minimal hanafuda card carrying the given ink color token. */
const hana = (color: string): Card => ({ design: 'SPADE', value: 0, glyph: '🎴', label: 'x', color, deck: 'hanafuda' });

describe('koikoiCategory', () => {
  it('maps gold to bright', () => {
    expect(koikoiCategory(hana('gold'))).toBe('bright');
  });

  it('maps purple to animal', () => {
    expect(koikoiCategory(hana('purple'))).toBe('animal');
  });

  it('maps red and blue to ribbon', () => {
    expect(koikoiCategory(hana('red'))).toBe('ribbon');
    expect(koikoiCategory(hana('blue'))).toBe('ribbon');
  });

  it('maps black (and any unknown color) to kasu', () => {
    expect(koikoiCategory(hana('black'))).toBe('kasu');
    expect(koikoiCategory(hana('teal'))).toBe('kasu');
    expect(koikoiCategory({ design: 'SPADE', value: 1 })).toBe('kasu');
  });
});

describe('groupCapturedByCategory', () => {
  it('returns all four keys even when empty', () => {
    const groups = groupCapturedByCategory([]);
    expect(Object.keys(groups).sort()).toEqual([...KOIKOI_CATEGORY_ORDER].sort());
    for (const cat of KOIKOI_CATEGORY_ORDER) {
      expect(groups[cat]).toEqual([]);
    }
  });

  it('buckets cards into their categories preserving order', () => {
    const cards = [hana('black'), hana('gold'), hana('purple'), hana('black'), hana('red')];
    const groups = groupCapturedByCategory(cards);
    expect(groups.bright).toHaveLength(1);
    expect(groups.animal).toHaveLength(1);
    expect(groups.ribbon).toHaveLength(1);
    expect(groups.kasu).toHaveLength(2);
  });
});
