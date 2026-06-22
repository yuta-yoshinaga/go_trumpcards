import { describe, expect, it } from 'vitest';
import { heartsPassTarget } from './heartsPass';

describe('heartsPassTarget', () => {
  it('passes left to the next seat', () => {
    expect(heartsPassTarget(0, 0)).toBe(1);
    expect(heartsPassTarget(3, 0)).toBe(0);
  });

  it('passes right to the previous seat', () => {
    expect(heartsPassTarget(0, 1)).toBe(3);
    expect(heartsPassTarget(2, 1)).toBe(1);
  });

  it('passes across to the opposite seat', () => {
    expect(heartsPassTarget(0, 2)).toBe(2);
    expect(heartsPassTarget(3, 2)).toBe(1);
  });

  it('returns the same seat for the no-pass round', () => {
    expect(heartsPassTarget(2, 3)).toBe(2);
  });
});
