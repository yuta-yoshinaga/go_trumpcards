import { describe, expect, it } from 'vitest';
import { checkBgBlackRatchet, countBgBlackUtilities } from './bg-black-ratchet.mjs';

describe('bg-black opacity ratchet', () => {
  it('counts opacity utilities and ignores unrelated classes', () => {
    expect(countBgBlackUtilities('bg-black/30 bg-black/20 hover:bg-black/40 bg-black')).toBe(3);
  });

  it('accepts counts at or below the cap and rejects additions', () => {
    expect(checkBgBlackRatchet(881, 881)).toBe(true);
    expect(checkBgBlackRatchet(880, 881)).toBe(true);
    expect(checkBgBlackRatchet(882, 881)).toBe(false);
  });
});
