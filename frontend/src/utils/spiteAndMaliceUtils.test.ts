import { describe, expect, it } from 'vitest';
import { isSpiteAndMaliceWild } from './spiteAndMaliceUtils';

describe('isSpiteAndMaliceWild', () => {
  it('is true only for the King', () => {
    expect(isSpiteAndMaliceWild({ value: 13 })).toBe(true);
    expect(isSpiteAndMaliceWild({ value: 12 })).toBe(false);
    expect(isSpiteAndMaliceWild({ value: 1 })).toBe(false);
  });

  it('is false for a missing card', () => {
    expect(isSpiteAndMaliceWild(null)).toBe(false);
    expect(isSpiteAndMaliceWild(undefined)).toBe(false);
  });
});
