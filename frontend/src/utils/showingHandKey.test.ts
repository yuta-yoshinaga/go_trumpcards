import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { showingHandKey } from './showingHandKey';

const card = (design: CardDesign, value: number): Card => ({ design, value });

describe('showingHandKey', () => {
  it('returns null when there are no visible cards', () => {
    expect(showingHandKey([])).toBeNull();
  });

  it('returns the high-card key for one visible card', () => {
    expect(showingHandKey([card('SPADE', 7)])).toBe('highCard');
  });

  it('returns the one-pair key for a visible pair', () => {
    expect(showingHandKey([card('SPADE', 7), card('HEART', 7)])).toBe('onePair');
  });
});
