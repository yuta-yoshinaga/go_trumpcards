import { describe, expect, it } from 'vitest';
import i18n from '../i18n';
import type { CardDesign } from '../types/card';
import { cardAlt } from './cardAlt';

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
});
