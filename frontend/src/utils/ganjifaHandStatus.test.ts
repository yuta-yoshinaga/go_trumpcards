import { describe, expect, it } from 'vitest';
import type { Card } from '../types/common';
import { ganjifaHandCardStatus, ganjifaSuitOf } from './ganjifaHandStatus';

describe('ganjifaSuitOf', () => {
  it('returns the design number for a strong suit glyph', () => {
    const card: Card = { design: 'SPADE', value: 12, glyph: '♛' };

    expect(ganjifaSuitOf(card)).toBe(1);
  });

  it('returns the design number for a weak suit glyph', () => {
    const card: Card = { design: 'SPADE', value: 1, glyph: '♪' };

    expect(ganjifaSuitOf(card)).toBe(5);
  });

  it('returns null for an unrecognized or missing glyph', () => {
    expect(ganjifaSuitOf({ design: 'SPADE', value: 7, glyph: '?' })).toBeNull();
    expect(ganjifaSuitOf({ design: 'SPADE', value: 7 })).toBeNull();
  });
});

describe('ganjifaHandCardStatus', () => {
  const t = (key: string, params: { suit: string }) => `${key}:${params.suit}`;

  it('uses the strong-suit wording for a strong suit', () => {
    expect(ganjifaHandCardStatus({ design: 'SPADE', value: 12, glyph: '♛' }, t)).toMatch(/^handCardStrong:/);
  });

  it('uses the weak-suit wording for a weak suit', () => {
    expect(ganjifaHandCardStatus({ design: 'SPADE', value: 1, glyph: '♪' }, t)).toMatch(/^handCardWeak:/);
  });

  it('returns undefined for a card without a known suit', () => {
    expect(ganjifaHandCardStatus({ design: 'SPADE', value: 7 }, t)).toBeUndefined();
  });
});
