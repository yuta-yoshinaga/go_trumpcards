import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { analyzeRetourneMatch } from './bouillotteRetourne';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('analyzeRetourneMatch', () => {
  it('returns no matches when the retourne is absent', () => {
    const hand = [c('SPADE', 13), c('HEART', 13), c('CLOVER', 4)];
    expect(analyzeRetourneMatch(hand, null)).toEqual({ matchingIndices: [], noteKey: null });
    expect(analyzeRetourneMatch(hand, undefined)).toEqual({ matchingIndices: [], noteKey: null });
  });

  it('flags the single hand card sharing the retourne rank without a combo note', () => {
    const hand = [c('SPADE', 13), c('HEART', 11), c('CLOVER', 4)];
    const result = analyzeRetourneMatch(hand, c('CLOVER', 13));
    expect(result.matchingIndices).toEqual([0]);
    expect(result.noteKey).toBeNull();
  });

  it('reports a favori when a matching pair is completed by the retourne', () => {
    const hand = [c('SPADE', 12), c('HEART', 12), c('CLOVER', 8)];
    const result = analyzeRetourneMatch(hand, c('DIAMOND', 12));
    expect(result.matchingIndices).toEqual([0, 1]);
    expect(result.noteKey).toBe('favori');
  });

  it('reports a carre when a hand brelan matches the retourne rank', () => {
    const hand = [c('SPADE', 9), c('HEART', 9), c('CLOVER', 9)];
    const result = analyzeRetourneMatch(hand, c('DIAMOND', 9));
    expect(result.matchingIndices).toEqual([0, 1, 2]);
    expect(result.noteKey).toBe('carre');
  });

  it('returns no matches when no hand card shares the retourne rank', () => {
    const hand = [c('SPADE', 13), c('HEART', 11), c('CLOVER', 4)];
    const result = analyzeRetourneMatch(hand, c('DIAMOND', 8));
    expect(result.matchingIndices).toEqual([]);
    expect(result.noteKey).toBeNull();
  });
});
