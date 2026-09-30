import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import golden from './__fixtures__/canastaFamilyCardValue.golden.json';
import { canastaFamilyCardValue } from './canastaFamilyScore';

describe('canastaFamilyCardValue golden vectors (shared with Go)', () => {
  it('checks every card in the shared fixture', () => {
    expect(golden.cases).toHaveLength(53);
    for (const c of golden.cases)
      expect(canastaFamilyCardValue({ design: c.design as Card['design'], value: c.value })).toBe(c.points);
  });
});
