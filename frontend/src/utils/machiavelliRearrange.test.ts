import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  designToNum,
  evaluateRearrange,
  isMachiavelliRun,
  isMachiavelliSet,
  isMachiavelliValidMeld,
  machiavelliConserves,
} from './machiavelliRearrange';

const c = (design: Card['design'], value: number): Card => ({ design, value });

describe('machiavelliRearrange helpers', () => {
  describe('designToNum', () => {
    it('maps suits to the Go numeric constants', () => {
      expect(designToNum('SPADE')).toBe(1);
      expect(designToNum('CLOVER')).toBe(2);
      expect(designToNum('HEART')).toBe(3);
      expect(designToNum('DIAMOND')).toBe(4);
      expect(designToNum('JOKER')).toBe(0);
    });
  });

  describe('isMachiavelliSet', () => {
    it('accepts 3 same-rank distinct-suit cards', () => {
      expect(isMachiavelliSet([c('SPADE', 9), c('HEART', 9), c('CLOVER', 9)])).toBe(true);
    });
    it('rejects fewer than 3 cards', () => {
      expect(isMachiavelliSet([c('SPADE', 9), c('HEART', 9)])).toBe(false);
    });
    it('rejects mismatched ranks', () => {
      expect(isMachiavelliSet([c('SPADE', 9), c('HEART', 8), c('CLOVER', 9)])).toBe(false);
    });
    it('rejects duplicate suits', () => {
      expect(isMachiavelliSet([c('SPADE', 9), c('SPADE', 9), c('CLOVER', 9)])).toBe(false);
    });
    it('rejects jokers', () => {
      expect(isMachiavelliSet([c('JOKER', 0), c('HEART', 9), c('CLOVER', 9)])).toBe(false);
    });
  });

  describe('isMachiavelliRun', () => {
    it('accepts a same-suit consecutive run', () => {
      expect(isMachiavelliRun([c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)])).toBe(true);
    });
    it('accepts an ace-low run (A-2-3)', () => {
      expect(isMachiavelliRun([c('SPADE', 1), c('SPADE', 2), c('SPADE', 3)])).toBe(true);
    });
    it('accepts an ace-high run (Q-K-A)', () => {
      expect(isMachiavelliRun([c('SPADE', 12), c('SPADE', 13), c('SPADE', 1)])).toBe(true);
    });
    it('rejects a wrap-around run (K-A-2)', () => {
      expect(isMachiavelliRun([c('SPADE', 13), c('SPADE', 1), c('SPADE', 2)])).toBe(false);
    });
    it('rejects mixed suits', () => {
      expect(isMachiavelliRun([c('SPADE', 3), c('HEART', 4), c('SPADE', 5)])).toBe(false);
    });
    it('rejects duplicate values', () => {
      expect(isMachiavelliRun([c('SPADE', 3), c('SPADE', 3), c('SPADE', 4)])).toBe(false);
    });
    it('rejects a non-consecutive run', () => {
      expect(isMachiavelliRun([c('SPADE', 3), c('SPADE', 5), c('SPADE', 6)])).toBe(false);
    });
  });

  describe('isMachiavelliValidMeld', () => {
    it('accepts either a set or a run', () => {
      expect(isMachiavelliValidMeld([c('SPADE', 9), c('HEART', 9), c('CLOVER', 9)])).toBe(true);
      expect(isMachiavelliValidMeld([c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)])).toBe(true);
    });
    it('rejects an invalid group', () => {
      expect(isMachiavelliValidMeld([c('SPADE', 3), c('HEART', 8), c('CLOVER', 9)])).toBe(false);
    });
  });

  describe('machiavelliConserves', () => {
    const oldTable = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]];
    it('holds when the proposed table is old + played', () => {
      const proposed = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5), c('SPADE', 6)]];
      expect(machiavelliConserves(oldTable, [c('SPADE', 6)], proposed)).toBe(true);
    });
    it('fails when a card is missing', () => {
      const proposed = [[c('SPADE', 4), c('SPADE', 5), c('SPADE', 6)]];
      expect(machiavelliConserves(oldTable, [c('SPADE', 6)], proposed)).toBe(false);
    });
    it('fails when card counts differ', () => {
      const proposed = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]];
      expect(machiavelliConserves(oldTable, [c('SPADE', 6)], proposed)).toBe(false);
    });
  });

  describe('evaluateRearrange', () => {
    const oldTable = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]];

    it('allows submitting a valid layoff-style rearrangement', () => {
      const played = [c('SPADE', 6)];
      const groups = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5), c('SPADE', 6)]];
      const evalResult = evaluateRearrange(groups, oldTable, played);
      expect(evalResult.groupValidity).toEqual([true]);
      expect(evalResult.allMeldsValid).toBe(true);
      expect(evalResult.conserves).toBe(true);
      expect(evalResult.playsFromHand).toBe(true);
      expect(evalResult.canSubmit).toBe(true);
    });

    it('allows splitting a run into two valid melds using a played card', () => {
      // Table run 3-4-5-6-7 borrowed + played S8/S9 → 3-4-5 and 6-7-8-9.
      const table = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5), c('SPADE', 6), c('SPADE', 7)]];
      const played = [c('SPADE', 8), c('SPADE', 9)];
      const groups = [
        [c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)],
        [c('SPADE', 6), c('SPADE', 7), c('SPADE', 8), c('SPADE', 9)],
      ];
      expect(evaluateRearrange(groups, table, played).canSubmit).toBe(true);
    });

    it('blocks submission when a group is not a valid meld', () => {
      const played = [c('HEART', 9)];
      const groups = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5), c('HEART', 9)]];
      const evalResult = evaluateRearrange(groups, oldTable, played);
      expect(evalResult.allMeldsValid).toBe(false);
      expect(evalResult.canSubmit).toBe(false);
    });

    it('blocks submission when cards are not conserved (unassigned remain)', () => {
      const played = [c('SPADE', 6)];
      // Only the table meld is grouped; the played 6 is left out → not conserved.
      const groups = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]];
      const evalResult = evaluateRearrange(groups, oldTable, played);
      expect(evalResult.conserves).toBe(false);
      expect(evalResult.canSubmit).toBe(false);
    });

    it('blocks submission when no hand card is played', () => {
      const groups = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5)]];
      const evalResult = evaluateRearrange(groups, oldTable, []);
      expect(evalResult.playsFromHand).toBe(false);
      expect(evalResult.canSubmit).toBe(false);
    });

    it('ignores empty groups when evaluating', () => {
      const played = [c('SPADE', 6)];
      const groups = [[c('SPADE', 3), c('SPADE', 4), c('SPADE', 5), c('SPADE', 6)], []];
      expect(evaluateRearrange(groups, oldTable, played).canSubmit).toBe(true);
    });
  });
});
