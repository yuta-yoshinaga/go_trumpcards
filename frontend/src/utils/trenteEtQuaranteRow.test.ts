import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { buildTrenteEtQuaranteRow, TRENTE_ET_QUARANTE_TARGET, trenteEtQuaranteCardValue } from './trenteEtQuaranteRow';

const card = (value: number): Card => ({ design: 'SPADE', value });

describe('trenteEtQuaranteCardValue', () => {
  it('counts A as 1 and pips at face value', () => {
    expect(trenteEtQuaranteCardValue(1)).toBe(1);
    expect(trenteEtQuaranteCardValue(7)).toBe(7);
    expect(trenteEtQuaranteCardValue(10)).toBe(10);
  });

  it('counts J/Q/K as 10', () => {
    expect(trenteEtQuaranteCardValue(11)).toBe(10);
    expect(trenteEtQuaranteCardValue(12)).toBe(10);
    expect(trenteEtQuaranteCardValue(13)).toBe(10);
  });
});

describe('buildTrenteEtQuaranteRow', () => {
  it('accumulates the running total across cards', () => {
    const steps = buildTrenteEtQuaranteRow([card(10), card(10), card(13)]);
    expect(steps.map((s) => s.cumulative)).toEqual([10, 20, 30]);
  });

  it('flags the first card to reach 31 as the crossing card', () => {
    // 10 + 10 + 10 + 5 = 35; the fourth card crosses 31.
    const steps = buildTrenteEtQuaranteRow([card(10), card(11), card(12), card(5)]);
    expect(steps.map((s) => s.crossing)).toEqual([false, false, false, true]);
    const crossing = steps.find((s) => s.crossing);
    expect(crossing?.cumulative).toBeGreaterThanOrEqual(TRENTE_ET_QUARANTE_TARGET);
    expect(crossing?.cumulative).toBe(35);
  });

  it('marks only one crossing card even when later totals stay above 31', () => {
    const steps = buildTrenteEtQuaranteRow([card(10), card(10), card(10), card(4)]);
    // 10,20,30,34 — third card (30) does not cross, fourth (34) does.
    expect(steps.filter((s) => s.crossing)).toHaveLength(1);
    expect(steps[3].crossing).toBe(true);
  });

  it('handles an exact landing on 31', () => {
    const steps = buildTrenteEtQuaranteRow([card(10), card(11), card(1)]);
    // 10, 20, 21 — never reaches 31, so no crossing.
    expect(steps.some((s) => s.crossing)).toBe(false);
    const crossed = buildTrenteEtQuaranteRow([card(10), card(11), card(11), card(1)]);
    // 10, 20, 30, 31 — the fourth card lands exactly on 31.
    expect(crossed[3].crossing).toBe(true);
    expect(crossed[3].cumulative).toBe(31);
  });

  it('returns an empty breakdown for no cards', () => {
    expect(buildTrenteEtQuaranteRow([])).toEqual([]);
  });
});
