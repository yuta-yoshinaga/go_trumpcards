import { describe, expect, it } from 'vitest';
import { suitName, valueName } from './cardUtils';

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

describe('suitName', () => {
  it('returns SPADE for suit 1', () => {
    expect(suitName(1)).toBe('SPADE');
  });

  it('returns CLOVER for suit 2', () => {
    expect(suitName(2)).toBe('CLOVER');
  });

  it('returns HEART for suit 3', () => {
    expect(suitName(3)).toBe('HEART');
  });

  it('returns DIAMOND for suit 4', () => {
    expect(suitName(4)).toBe('DIAMOND');
  });

  it('returns empty string for unknown suit', () => {
    expect(suitName(0)).toBe('');
    expect(suitName(99)).toBe('');
  });
});
