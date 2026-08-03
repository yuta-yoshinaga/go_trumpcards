import { describe, expect, it } from 'vitest';
import { formatGanjifaSuit, GANJIFA_STRONG_SUIT_MAX, GANJIFA_SUIT_COUNT, isGanjifaStrongSuit } from './ganjifa';

describe('isGanjifaStrongSuit', () => {
  // This predicate is the frontend's copy of `domain.GanjifaIsStrongSuit`; if the
  // boundary drifts the UI tells players the ranks read the wrong way round.
  it('treats designs 1-4 as strong and 5-8 as weak', () => {
    for (let design = 1; design <= GANJIFA_STRONG_SUIT_MAX; design++) {
      expect(isGanjifaStrongSuit(design)).toBe(true);
    }
    for (let design = GANJIFA_STRONG_SUIT_MAX + 1; design <= GANJIFA_SUIT_COUNT; design++) {
      expect(isGanjifaStrongSuit(design)).toBe(false);
    }
  });

  it('rejects designs outside the deck', () => {
    expect(isGanjifaStrongSuit(0)).toBe(false);
    expect(isGanjifaStrongSuit(-1)).toBe(false);
    expect(isGanjifaStrongSuit(GANJIFA_SUIT_COUNT + 1)).toBe(false);
  });
});

describe('formatGanjifaSuit', () => {
  it('gives every suit a distinct glyph and name', () => {
    const seen = new Set<string>();
    for (let design = 1; design <= GANJIFA_SUIT_COUNT; design++) {
      const label = formatGanjifaSuit(design);
      expect(label).not.toBe('?');
      expect(seen.has(label)).toBe(false);
      seen.add(label);
    }
    expect(seen.size).toBe(GANJIFA_SUIT_COUNT);
  });

  it('falls back to "?" outside the deck', () => {
    expect(formatGanjifaSuit(0)).toBe('?');
    expect(formatGanjifaSuit(GANJIFA_SUIT_COUNT + 1)).toBe('?');
  });
});
