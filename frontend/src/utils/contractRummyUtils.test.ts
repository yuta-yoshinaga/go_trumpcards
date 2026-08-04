import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  CONTRACT_SLOT_RUN,
  CONTRACT_SLOT_SET,
  evaluateContractSlot,
  isContractRummyMeld,
  isContractRun,
  isContractSet,
} from './contractRummyUtils';

const card = (design: string, value: number): Card => ({ design, value }) as Card;

describe('isContractSet', () => {
  it('returns true when all cards share the same rank', () => {
    expect(isContractSet([card('SPADE', 7), card('HEART', 7), card('CLOVER', 7)])).toBe(true);
  });

  it('returns false when ranks differ', () => {
    expect(isContractSet([card('SPADE', 7), card('HEART', 8)])).toBe(false);
  });

  it('returns false for a single card', () => {
    expect(isContractSet([card('SPADE', 7)])).toBe(false);
  });

  it('returns false for a 2-card pair (Contract Rummy sets require ≥3)', () => {
    expect(isContractSet([card('SPADE', 7), card('HEART', 7)])).toBe(false);
  });
});

describe('isContractRun', () => {
  it('accepts an ace-low run', () => {
    expect(isContractRun([card('SPADE', 1), card('SPADE', 2), card('SPADE', 3), card('SPADE', 4)])).toBe(true);
  });

  it('accepts an ace-high run (J-Q-K-A)', () => {
    expect(isContractRun([card('SPADE', 11), card('SPADE', 12), card('SPADE', 13), card('SPADE', 1)])).toBe(true);
  });

  it('rejects mixed suits', () => {
    expect(isContractRun([card('SPADE', 2), card('HEART', 3), card('SPADE', 4), card('SPADE', 5)])).toBe(false);
  });

  it('rejects non-consecutive values', () => {
    expect(isContractRun([card('SPADE', 2), card('SPADE', 3), card('SPADE', 5), card('SPADE', 6)])).toBe(false);
  });

  it('rejects K-A-2 wrap-around', () => {
    expect(isContractRun([card('SPADE', 13), card('SPADE', 1), card('SPADE', 2), card('SPADE', 3)])).toBe(false);
  });

  it('rejects duplicate ranks within the run', () => {
    expect(isContractRun([card('SPADE', 2), card('SPADE', 3), card('SPADE', 3), card('SPADE', 4)])).toBe(false);
  });
});

describe('evaluateContractSlot', () => {
  it('returns an empty state when no cards have been placed', () => {
    expect(evaluateContractSlot({ kind: CONTRACT_SLOT_SET, size: 3 }, [])).toEqual({
      required: 3,
      placed: 0,
      satisfied: false,
      invalid: false,
    });
  });

  it('returns in-progress when fewer cards than required', () => {
    const ev = evaluateContractSlot({ kind: CONTRACT_SLOT_SET, size: 3 }, [card('SPADE', 7)]);
    expect(ev).toEqual({ required: 3, placed: 1, satisfied: false, invalid: false });
  });

  it('returns satisfied when a valid set fills the slot', () => {
    const ev = evaluateContractSlot({ kind: CONTRACT_SLOT_SET, size: 3 }, [
      card('SPADE', 7),
      card('HEART', 7),
      card('CLOVER', 7),
    ]);
    expect(ev.satisfied).toBe(true);
    expect(ev.invalid).toBe(false);
  });

  it('flags invalid when filled with the wrong combination', () => {
    const ev = evaluateContractSlot({ kind: CONTRACT_SLOT_SET, size: 3 }, [
      card('SPADE', 7),
      card('HEART', 8),
      card('CLOVER', 7),
    ]);
    expect(ev.satisfied).toBe(false);
    expect(ev.invalid).toBe(true);
  });

  it('returns satisfied for a valid run', () => {
    const ev = evaluateContractSlot({ kind: CONTRACT_SLOT_RUN, size: 4 }, [
      card('SPADE', 4),
      card('SPADE', 5),
      card('SPADE', 6),
      card('SPADE', 7),
    ]);
    expect(ev.satisfied).toBe(true);
  });

  it('flags over-filled slots as invalid', () => {
    const ev = evaluateContractSlot({ kind: CONTRACT_SLOT_SET, size: 3 }, [
      card('SPADE', 7),
      card('HEART', 7),
      card('CLOVER', 7),
      card('DIAMOND', 7),
    ]);
    expect(ev.invalid).toBe(true);
    expect(ev.satisfied).toBe(false);
  });
});

describe('isContractRummyMeld', () => {
  const c = (design: Card['design'], value: number): Card => ({ design, value });

  it('accepts a set of three', () => {
    expect(isContractRummyMeld([c('SPADE', 7), c('HEART', 7), c('CLOVER', 7)])).toBe(true);
  });

  it('accepts a run of three in one suit', () => {
    expect(isContractRummyMeld([c('SPADE', 5), c('SPADE', 6), c('SPADE', 7)])).toBe(true);
  });

  it('rejects three unrelated cards, which the count-only check allowed', () => {
    expect(isContractRummyMeld([c('SPADE', 5), c('HEART', 9), c('CLOVER', 13)])).toBe(false);
  });

  it('rejects a run split across suits', () => {
    expect(isContractRummyMeld([c('SPADE', 5), c('HEART', 6), c('SPADE', 7)])).toBe(false);
  });

  it('rejects fewer than three', () => {
    expect(isContractRummyMeld([c('SPADE', 7), c('HEART', 7)])).toBe(false);
  });
});
