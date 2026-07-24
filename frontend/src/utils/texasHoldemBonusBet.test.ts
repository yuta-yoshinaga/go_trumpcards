import { describe, expect, it } from 'vitest';
import {
  TEXASHOLDEMBONUS_FLOP_MULTIPLIER,
  TEXASHOLDEMBONUS_RAISE_MULTIPLIER,
  texasHoldemBonusBetCost,
} from './texasHoldemBonusBet';

describe('texasHoldemBonusBetCost', () => {
  it('computes the Play (flop) bet as 2× the ante', () => {
    expect(texasHoldemBonusBetCost(100, TEXASHOLDEMBONUS_FLOP_MULTIPLIER)).toBe(200);
    expect(texasHoldemBonusBetCost(10, TEXASHOLDEMBONUS_FLOP_MULTIPLIER)).toBe(20);
  });

  it('computes the Raise (flop/turn) bet as 1× the ante', () => {
    expect(texasHoldemBonusBetCost(100, TEXASHOLDEMBONUS_RAISE_MULTIPLIER)).toBe(100);
    expect(texasHoldemBonusBetCost(10, TEXASHOLDEMBONUS_RAISE_MULTIPLIER)).toBe(10);
  });

  it('returns 0 for a zero ante', () => {
    expect(texasHoldemBonusBetCost(0, TEXASHOLDEMBONUS_FLOP_MULTIPLIER)).toBe(0);
    expect(texasHoldemBonusBetCost(0, TEXASHOLDEMBONUS_RAISE_MULTIPLIER)).toBe(0);
  });

  it('clamps a negative ante to 0', () => {
    expect(texasHoldemBonusBetCost(-50, TEXASHOLDEMBONUS_FLOP_MULTIPLIER)).toBe(0);
  });

  it('exposes the domain multiples', () => {
    expect(TEXASHOLDEMBONUS_FLOP_MULTIPLIER).toBe(2);
    expect(TEXASHOLDEMBONUS_RAISE_MULTIPLIER).toBe(1);
  });
});
