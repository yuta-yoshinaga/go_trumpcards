import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import golden from './__fixtures__/tarabishPoints.golden.json';
import { tarabishCardPoints } from './tarabishPoints';

/**
 * The point table lives twice: `TarabishCardPoints` in
 * `internal/domain/Tarabish.go` (which scores the round) and this module (which
 * labels the hand). `TestTarabishCardPoints_GoldenVectors` asserts the same
 * vectors from the Go side, so changing one alone fails that side.
 */
describe('tarabishCardPoints golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    expect(tarabishCardPoints({ design: c.design as Card['design'], value: c.value }, c.trumpSuit)).toBe(c.points);
  });

  it('scores nothing for a missing card', () => {
    expect(tarabishCardPoints(null, 1)).toBe(0);
  });

  // 黄金ベクタが踏まない唯一の枝: 未知のデザイン (ジョーカー等) は
  // 切り札扱いにならず、平札の表で 0 になる。
  it('treats an unknown design as a plain card', () => {
    expect(tarabishCardPoints({ design: 'JOKER' as Card['design'], value: 11 }, 1)).toBe(2);
    expect(tarabishCardPoints({ design: 'JOKER' as Card['design'], value: 9 }, 1)).toBe(0);
  });
});
