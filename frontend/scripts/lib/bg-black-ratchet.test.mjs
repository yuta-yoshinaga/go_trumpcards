import { describe, expect, it } from 'vitest';
import { addBgBlackRatchetViolation, checkBgBlackRatchet, countBgBlackUtilities } from './bg-black-ratchet.mjs';

describe('bg-black opacity ratchet', () => {
  it('counts opacity utilities and ignores unrelated classes', () => {
    expect(countBgBlackUtilities('bg-black/30 bg-black/20 hover:bg-black/40 bg-black')).toBe(3);
  });

  it('counts arbitrary opacity values', () => {
    expect(countBgBlackUtilities('bg-black/[0.3] bg-black/[.3]')).toBe(2);
  });

  it('adds an over-ceiling count to the shared violations list', () => {
    const violations = [];

    addBgBlackRatchetViolation(violations, 3, 2);

    expect(violations).toEqual([
      {
        file: 'src/components and src/pages',
        match: '3 bg-black utilities',
        message: 'Do not add new uses; lower the ceiling when existing uses are removed.',
      },
    ]);
  });

  it('does not add a violation when the count is within the ceiling', () => {
    const violations = [];

    addBgBlackRatchetViolation(violations, 2, 2);

    expect(violations).toEqual([]);
  });

  it('accepts counts at or below the cap and rejects additions', () => {
    expect(checkBgBlackRatchet(881, 881)).toBe(true);
    expect(checkBgBlackRatchet(880, 881)).toBe(true);
    expect(checkBgBlackRatchet(882, 881)).toBe(false);
  });
});
