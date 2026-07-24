import { describe, expect, it } from 'vitest';
import type { ScopaScoreDetail } from '../types/card';
import { scopaScoreBreakdown } from './scopaScoreBreakdown';

function detail(overrides: Partial<ScopaScoreDetail> = {}): ScopaScoreDetail {
  return {
    cards: { 0: 0, 1: 0 },
    diamonds: { 0: 0, 1: 0 },
    sevens: { 0: 0, 1: 0 },
    hasSetteBello: -1,
    scopas: { 0: 0, 1: 0 },
    gained: { 0: 0, 1: 0 },
    ...overrides,
  };
}

describe('scopaScoreBreakdown', () => {
  it('returns the four award categories in a stable order', () => {
    const rows = scopaScoreBreakdown(detail());
    expect(rows.map((r) => r.key)).toEqual(['cards', 'denari', 'primiera', 'settebello']);
  });

  it('awards a category to the unique count leader with 1 point', () => {
    const rows = scopaScoreBreakdown(
      detail({ cards: { 0: 20, 1: 16 }, diamonds: { 0: 4, 1: 6 }, sevens: { 0: 3, 1: 1 } }),
    );
    const byKey = Object.fromEntries(rows.map((r) => [r.key, r]));
    expect(byKey.cards).toEqual({ key: 'cards', winner: 0, points: 1 });
    expect(byKey.denari).toEqual({ key: 'denari', winner: 1, points: 1 });
    expect(byKey.primiera).toEqual({ key: 'primiera', winner: 0, points: 1 });
  });

  it('marks a tied count as unawarded (winner -1, 0 points)', () => {
    const rows = scopaScoreBreakdown(detail({ cards: { 0: 18, 1: 18 } }));
    const cards = rows.find((r) => r.key === 'cards');
    expect(cards).toEqual({ key: 'cards', winner: -1, points: 0 });
  });

  it('marks an all-zero count as unawarded', () => {
    const rows = scopaScoreBreakdown(detail({ diamonds: { 0: 0, 1: 0 } }));
    const denari = rows.find((r) => r.key === 'denari');
    expect(denari).toEqual({ key: 'denari', winner: -1, points: 0 });
  });

  it('takes the settebello winner directly from hasSetteBello', () => {
    const wonRows = scopaScoreBreakdown(detail({ hasSetteBello: 1 }));
    expect(wonRows.find((r) => r.key === 'settebello')).toEqual({ key: 'settebello', winner: 1, points: 1 });
    const noneRows = scopaScoreBreakdown(detail({ hasSetteBello: -1 }));
    expect(noneRows.find((r) => r.key === 'settebello')).toEqual({ key: 'settebello', winner: -1, points: 0 });
  });
});
