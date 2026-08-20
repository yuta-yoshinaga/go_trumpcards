import { describe, expect, it } from 'vitest';
import type { GoStopBreakdown } from '../types/card';
import golden from './__fixtures__/gostopNearYaku.golden.json';
import { computeNearYaku } from './gostopYaku';

/**
 * The near-yaku preview lives twice: `GoStopComputeNearYaku` in
 * `internal/domain/GoStop.go` (which the CUI shows) and this module (which the
 * Web decision panel shows). These golden vectors are also asserted by
 * `TestGoStopComputeNearYaku_GoldenVectors`, so changing the thresholds or the
 * preview window on one side alone fails that side.
 */
const breakdown = (bright: number, ribbon: number, animal: number, pi: number): GoStopBreakdown => ({
  gwang: 0,
  godori: 0,
  tti: 0,
  yeol: 0,
  pi: 0,
  base: 0,
  goCount: 0,
  goMult: 1,
  goScore: 0,
  brightCount: bright,
  ribbonCount: ribbon,
  animalCount: animal,
  piCount: pi,
});

describe('computeNearYaku golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    const got = computeNearYaku(breakdown(c.counts.bright, c.counts.ribbon, c.counts.animal, c.counts.pi));

    expect(got).toEqual(c.near);
  });
});
