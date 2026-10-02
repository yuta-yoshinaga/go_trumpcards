import { describe, expect, it } from 'vitest';
import type { Card } from '../types/common';
import { evaluateBiribaMeld } from './biribaMeld';

const c = (value: number, design: Card['design'] = 'HEART'): Card => ({ value, design });
const opts = { hasInitMeld: true, minMeld: 50 };

describe('evaluateBiribaMeld', () => {
  it('requires a selection and at least three cards', () => {
    expect(evaluateBiribaMeld([], [], opts).reason).toBe('selectCards');
    expect(evaluateBiribaMeld([c(5), c(6)], [], opts).reason).toBe('tooFewCards');
  });
  it('validates a same-suit consecutive run regardless of selection order', () => {
    expect(evaluateBiribaMeld([c(7), c(5), c(6)], [], opts)).toEqual({ ok: true });
    expect(evaluateBiribaMeld([c(5), c(6), c(8)], [], opts).reason).toBe('notConsecutive');
    expect(evaluateBiribaMeld([c(5), c(6), c(7, 'SPADE')], [], opts).reason).toBe('sameSuit');
    expect(evaluateBiribaMeld([c(5), c(5), c(6)], [], opts).reason).toBe('duplicateRank');
  });
  it('returns the first natural-card error in Go validation order', () => {
    expect(evaluateBiribaMeld([c(3, 'CLOVER'), c(5), c(7, 'SPADE')], [], opts).reason).toBe('blackThree');
    expect(evaluateBiribaMeld([c(5), c(6, 'SPADE'), c(6)], [], opts).reason).toBe('sameSuit');
    expect(evaluateBiribaMeld([c(5), c(6), c(5), c(8)], [], opts).reason).toBe('duplicateRank');
  });
  it('applies wild and black three restrictions', () => {
    expect(evaluateBiribaMeld([c(3, 'SPADE'), c(5), c(6)], [], opts).reason).toBe('blackThree');
    expect(evaluateBiribaMeld([c(5), c(6), c(0, 'JOKER')], [], opts)).toEqual({ ok: true });
    expect(
      evaluateBiribaMeld([c(5), c(6), c(0, 'JOKER'), c(0, 'JOKER'), c(0, 'JOKER'), c(0, 'JOKER')], [], opts).reason,
    ).toBe('tooManyWilds');
  });
  it('extends a compatible existing meld and rejects incompatible additions', () => {
    const meld = { cards: [c(5), c(6), c(7)], rank: 5, isNatural: true, isBiriba: false };
    expect(evaluateBiribaMeld([c(8)], [meld], opts)).toEqual({ ok: true });
    expect(evaluateBiribaMeld([c(9)], [meld], opts).reason).toBe('tooFewCards');
    expect(evaluateBiribaMeld([c(9)], [{ ...meld, cards: [c(5), c(6, 'SPADE'), c(7)] }], opts).reason).toBe(
      'tooFewCards',
    );
  });
  it('enforces the first meld minimum and discard top requirement', () => {
    expect(evaluateBiribaMeld([c(4), c(5), c(6)], [], { hasInitMeld: false, minMeld: 50 }).reason).toBe(
      'initialMinimum',
    );
    const run = [c(8), c(9), c(10), c(11), c(12)];
    expect(evaluateBiribaMeld(run, [], { hasInitMeld: false, minMeld: 50 })).toEqual({ ok: true });
    expect(evaluateBiribaMeld(run, [], { ...opts, drewFromDiscard: true, includesDrawnCard: false }).reason).toBe(
      'discardTopRequired',
    );
  });
});
