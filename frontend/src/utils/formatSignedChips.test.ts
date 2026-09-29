import { describe, expect, it } from 'vitest';
import { formatSignedChips } from './formatSignedChips';

describe('formatSignedChips', () => {
  it('formats positive amounts with a plus sign', () => {
    expect(formatSignedChips(12)).toBe('+12');
  });

  it('preserves the minus sign for negative amounts', () => {
    expect(formatSignedChips(-7)).toBe('-7');
  });

  it('formats zero with the plus-minus sign', () => {
    expect(formatSignedChips(0)).toBe('±0');
  });
});
