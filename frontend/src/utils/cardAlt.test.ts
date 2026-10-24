import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import type { CardDesign } from '../types/card';
import { cardAlt, isRedSuitDesign, isSuitDesign, suitSymbol, suitSymbolAt } from './cardAlt';

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

  it('uses the localized hanafuda label instead of the server label', async () => {
    await i18n.changeLanguage('ja');
    expect(
      cardAlt({ design: 'SPADE', value: 0, deck: 'hanafuda', month: 1, kind: 'bright', label: '松·光', glyph: '🌕' }),
    ).toBe('松·光 🌕');
    await i18n.changeLanguage('en');
    expect(
      cardAlt({ design: 'SPADE', value: 0, deck: 'hanafuda', month: 1, kind: 'bright', label: '松·光', glyph: '🌕' }),
    ).toBe('Pine·Bright 🌕');
    await i18n.changeLanguage('ja');
  });

  it('uses the localized Unsun Karuta name without the server glyph', async () => {
    const sota = {
      design: 'SPADE',
      value: 10,
      deck: 'unsun',
      kind: 'sota',
      unsunSuit: 'pao',
      label: 'ソウタ',
      glyph: '棒',
    } as const;
    await i18n.changeLanguage('ja');
    expect(cardAlt(sota)).toBe('ソウタ・ぱお');
    await i18n.changeLanguage('en');
    expect(cardAlt(sota)).toBe('Sota of Pao');
    await i18n.changeLanguage('ja');
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

describe('suitSymbolAt', () => {
  it.each<[number, string]>([
    [1, '♠'],
    [2, '♣'],
    [3, '♥'],
    [4, '♦'],
  ])('maps suit index %i to %s', (suit, symbol) => {
    expect(suitSymbolAt(suit)).toBe(symbol);
  });

  it.each([0, 5, -1])('uses an empty fallback for invalid index %i by default', (suit) => {
    expect(suitSymbolAt(suit)).toBe('');
  });

  it.each([0, 5, -1])('uses the supplied fallback for invalid index %i', (suit) => {
    expect(suitSymbolAt(suit, '?')).toBe('?');
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
