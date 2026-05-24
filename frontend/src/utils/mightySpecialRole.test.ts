import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { MIGHTY_SUIT, mightyRoleGlyph, mightySpecialRole } from './mightySpecialRole';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('mightySpecialRole', () => {
  it('returns "mighty" for ♠A when trump is not Spade', () => {
    expect(mightySpecialRole(c('SPADE', 1), MIGHTY_SUIT.HEART, null)).toBe('mighty');
    expect(mightySpecialRole(c('SPADE', 1), -1, null)).toBe('mighty');
  });

  it('returns "mighty" for ♦A when trump is Spade', () => {
    expect(mightySpecialRole(c('DIAMOND', 1), MIGHTY_SUIT.SPADE, null)).toBe('mighty');
    expect(mightySpecialRole(c('SPADE', 1), MIGHTY_SUIT.SPADE, null)).not.toBe('mighty');
  });

  it('returns "jokerCall" for ♣3 when trump is not Clover', () => {
    expect(mightySpecialRole(c('CLOVER', 3), MIGHTY_SUIT.HEART, null)).toBe('jokerCall');
  });

  it('returns "jokerCall" for ♠3 when trump is Clover', () => {
    expect(mightySpecialRole(c('SPADE', 3), MIGHTY_SUIT.CLOVER, null)).toBe('jokerCall');
    expect(mightySpecialRole(c('CLOVER', 3), MIGHTY_SUIT.CLOVER, null)).not.toBe('jokerCall');
  });

  it('returns "joker" for any JOKER-design card', () => {
    expect(mightySpecialRole(c('JOKER', 0), MIGHTY_SUIT.HEART, null)).toBe('joker');
  });

  it('returns "partner" when card matches partnerCard', () => {
    const partner = c('HEART', 12);
    expect(mightySpecialRole(c('HEART', 12), MIGHTY_SUIT.HEART, partner)).toBe('partner');
    expect(mightySpecialRole(c('HEART', 11), MIGHTY_SUIT.HEART, partner)).toBeNull();
  });

  it('returns null for ordinary cards', () => {
    expect(mightySpecialRole(c('DIAMOND', 7), MIGHTY_SUIT.HEART, null)).toBeNull();
  });

  it('prioritises joker before partner check', () => {
    const partner = c('JOKER', 0);
    expect(mightySpecialRole(c('JOKER', 0), MIGHTY_SUIT.HEART, partner)).toBe('joker');
  });
});

describe('mightyRoleGlyph', () => {
  it('returns distinct glyphs per role', () => {
    const glyphs = new Set(['mighty', 'jokerCall', 'joker', 'partner'].map((r) => mightyRoleGlyph(r as never)));
    expect(glyphs.size).toBe(4);
  });
});
