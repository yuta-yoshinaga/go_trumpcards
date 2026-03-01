import { describe, expect, it } from 'vitest';
import { valueName } from './cardUtils';

describe('valueName', () => {
  it('returns A for value 1', () => {
    expect(valueName(1)).toBe('A');
  });

  it('returns J for value 11', () => {
    expect(valueName(11)).toBe('J');
  });

  it('returns Q for value 12', () => {
    expect(valueName(12)).toBe('Q');
  });

  it('returns K for value 13', () => {
    expect(valueName(13)).toBe('K');
  });

  it('returns string representation for other values', () => {
    expect(valueName(5)).toBe('5');
    expect(valueName(10)).toBe('10');
  });
});
