import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import golden from './__fixtures__/zhengInvalidReason.golden.json';
import { zhengInvalidReason } from './zhengComboValidator';

/**
 * The "why can't I play this" classification lives twice: `ZhengInvalidReason`
 * in `internal/domain/ZhengEval.go` (which now shapes the CUI's rejection
 * message) and this module (which shapes the Web tooltip). These golden vectors
 * are also asserted by `TestZhengInvalidReason_GoldenVectors`, so a rule change
 * on one side alone fails that side.
 */
const DESIGNS: CardDesign[] = ['JOKER', 'SPADE', 'CLOVER', 'HEART', 'DIAMOND'];

const toCards = (in_: { suit: number; value: number }[]): Card[] =>
  in_.map((c) => ({ design: DESIGNS[c.suit], value: c.value }));

describe('zhengInvalidReason golden vectors (shared with the Go domain)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    const got = zhengInvalidReason(toCards(c.cards), toCards(c.table), c.tablePlayType);

    // Go 側は「出せる」を空文字で返し、TS 側は null を返す。
    expect(got ?? '').toBe(c.reason);
  });
});
