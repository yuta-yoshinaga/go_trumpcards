import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { frenchTarotTarget, heldBouts } from './frenchTarotBouts';

const trump = (value: number): Card => ({
  design: 'JOKER',
  value,
  glyph: '✦',
  label: String(value),
  color: 'purple',
  deck: 'tarot',
});
const excuse = (): Card => ({ design: 'JOKER', value: 0, glyph: '★', label: 'Excuse', color: 'gold', deck: 'tarot' });
const suit = (value: number): Card => ({
  design: 'CLOVER',
  value,
  glyph: '♣',
  label: String(value),
  color: 'black',
  deck: 'tarot',
});

describe('heldBouts', () => {
  it('detects the 21 and the Excuse and returns them in canonical order', () => {
    expect(heldBouts([excuse(), suit(5), trump(21)])).toEqual(['twentyOne', 'excuse']);
  });

  it('detects the Petit (trump valued 1) but not a suit Ace also valued 1', () => {
    expect(heldBouts([trump(1), suit(1)])).toEqual(['petit']);
  });

  it('returns all three bouts when the whole set is held', () => {
    expect(heldBouts([trump(1), trump(21), excuse()])).toEqual(['twentyOne', 'petit', 'excuse']);
  });

  it('returns an empty array when no bouts are held', () => {
    expect(heldBouts([suit(1), suit(14), trump(5), trump(20)])).toEqual([]);
  });
});

describe('frenchTarotTarget', () => {
  it('maps bout count to the required card-point target', () => {
    expect(frenchTarotTarget(0)).toBe(56);
    expect(frenchTarotTarget(1)).toBe(51);
    expect(frenchTarotTarget(2)).toBe(41);
    expect(frenchTarotTarget(3)).toBe(36);
  });

  it('treats negative counts as zero bouts', () => {
    expect(frenchTarotTarget(-1)).toBe(56);
  });
});
