import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import type { CardDesign } from '../types/card';
import { cardAlt, isRedSuitDesign, isSuitDesign, suitSymbol } from './cardAlt';

describe('cardAlt', () => {
  it('returns localized joker text for JOKER', () => {
    expect(cardAlt({ design: 'JOKER', value: 0 })).toBe(i18n.t('common:card.joker'));
  });

  it.each<[CardDesign, string]>([
    ['SPADE', '♠'],
    ['HEART', '♥'],
    ['DIAMOND', '♦'],
    ['CLOVER', '♣'],
  ])('maps %s to %s', (design, symbol) => {
    expect(cardAlt({ design, value: 5 })).toBe(`${symbol} 5`);
  });

  it('converts face card values via valueName', () => {
    expect(cardAlt({ design: 'SPADE', value: 1 })).toBe('♠ A');
    expect(cardAlt({ design: 'HEART', value: 11 })).toBe('♥ J');
    expect(cardAlt({ design: 'DIAMOND', value: 12 })).toBe('♦ Q');
    expect(cardAlt({ design: 'CLOVER', value: 13 })).toBe('♣ K');
  });

  it('falls back to raw design for unknown design', () => {
    expect(cardAlt({ design: 'UNKNOWN' as CardDesign, value: 3 })).toBe('UNKNOWN 3');
  });

  it('uses the descriptor label and glyph for procedural (non-52 deck) cards', () => {
    expect(cardAlt({ design: 'JOKER', value: 1, label: 'Wizard', glyph: '✦', deck: 'wizard' })).toBe('Wizard ✦');
  });

  it('uses the descriptor label alone when no glyph is present', () => {
    expect(cardAlt({ design: 'JOKER', value: 1, label: 'Jester', deck: 'wizard' })).toBe('Jester');
  });
});

describe('suitSymbol', () => {
  it.each<[CardDesign, string]>([
    ['SPADE', '♠'],
    ['HEART', '♥'],
    ['DIAMOND', '♦'],
    ['CLOVER', '♣'],
  ])('maps %s to %s', (design, symbol) => {
    expect(suitSymbol(design)).toBe(symbol);
  });

  it('falls back to the raw design for unknown designs', () => {
    expect(suitSymbol('JOKER')).toBe('JOKER');
  });
});

describe('isSuitDesign', () => {
  it.each(['SPADE', 'CLOVER', 'HEART', 'DIAMOND'])('is true for the known suit %s', (design) => {
    expect(isSuitDesign(design)).toBe(true);
  });

  it.each(['JOKER', '', 'UNKNOWN'])('is false for the non-suit %s', (design) => {
    expect(isSuitDesign(design)).toBe(false);
  });
});

describe('isRedSuitDesign', () => {
  it.each(['HEART', 'DIAMOND'])('is true for the red suit %s', (design) => {
    expect(isRedSuitDesign(design)).toBe(true);
  });

  it.each(['SPADE', 'CLOVER', 'JOKER', ''])('is false for %s', (design) => {
    expect(isRedSuitDesign(design)).toBe(false);
  });
});
