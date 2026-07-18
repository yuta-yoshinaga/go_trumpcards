import { describe, expect, it } from 'vitest';
import type { Card, ContractRummyContractSlot } from '../types/card';
import { describeCariocaSlotShortfall, evaluateCariocaContractSlot, isCariocaRun, isCariocaSet } from './cariocaUtils';
import { CONTRACT_SLOT_RUN, CONTRACT_SLOT_SET } from './contractRummyUtils';

const card = (design: Card['design'], value: number): Card => ({ design, value });
const joker = (): Card => ({ design: 'JOKER', value: 0 });

describe('isCariocaSet', () => {
  it('accepts a plain 3-of-a-kind', () => {
    expect(isCariocaSet([card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5)])).toBe(true);
  });

  it('accepts a set completed by a joker', () => {
    expect(isCariocaSet([card('SPADE', 5), card('HEART', 5), joker()])).toBe(true);
  });

  it('rejects two jokers', () => {
    expect(isCariocaSet([card('SPADE', 5), joker(), joker()])).toBe(false);
  });

  it('rejects fewer than 3 cards', () => {
    expect(isCariocaSet([card('SPADE', 5), joker()])).toBe(false);
  });

  it('rejects mixed ranks', () => {
    expect(isCariocaSet([card('SPADE', 5), card('HEART', 6), card('DIAMOND', 5)])).toBe(false);
  });

  it('rejects an all-joker meld (no real rank)', () => {
    expect(isCariocaSet([joker(), joker(), joker()])).toBe(false);
  });
});

describe('isCariocaRun', () => {
  it('accepts a plain 4-card run', () => {
    expect(isCariocaRun([card('SPADE', 2), card('SPADE', 3), card('SPADE', 4), card('SPADE', 5)])).toBe(true);
  });

  it('accepts a run with a joker filling an interior gap', () => {
    // 7-8-[9]-10
    expect(isCariocaRun([card('HEART', 7), card('HEART', 8), joker(), card('HEART', 10)])).toBe(true);
  });

  it('accepts a run with a joker extending an end', () => {
    // 6-7-8-[9]
    expect(isCariocaRun([card('SPADE', 6), card('SPADE', 7), card('SPADE', 8), joker()])).toBe(true);
  });

  it('accepts an Ace-high run', () => {
    // J-Q-K-A
    expect(isCariocaRun([card('DIAMOND', 11), card('DIAMOND', 12), card('DIAMOND', 13), card('DIAMOND', 1)])).toBe(
      true,
    );
  });

  it('rejects mixed suits', () => {
    expect(isCariocaRun([card('SPADE', 2), card('HEART', 3), card('SPADE', 4), card('SPADE', 5)])).toBe(false);
  });

  it('rejects two jokers', () => {
    expect(isCariocaRun([card('SPADE', 2), card('SPADE', 3), joker(), joker()])).toBe(false);
  });

  it('rejects a non-consecutive spread even with a joker', () => {
    // 2, 5, 9 span too wide for one joker to bridge
    expect(isCariocaRun([card('SPADE', 2), card('SPADE', 5), card('SPADE', 9), joker()])).toBe(false);
  });

  it('rejects fewer than 4 cards', () => {
    expect(isCariocaRun([card('SPADE', 2), card('SPADE', 3), card('SPADE', 4)])).toBe(false);
  });
});

describe('evaluateCariocaContractSlot', () => {
  const setSlot: ContractRummyContractSlot = { kind: CONTRACT_SLOT_SET, size: 3 };
  const runSlot: ContractRummyContractSlot = { kind: CONTRACT_SLOT_RUN, size: 4 };

  it('reports satisfied for a joker-wild set slot', () => {
    const ev = evaluateCariocaContractSlot(setSlot, [card('SPADE', 5), card('HEART', 5), joker()]);
    expect(ev).toEqual({ required: 3, placed: 3, satisfied: true, invalid: false });
  });

  it('reports satisfied for a joker-wild run slot', () => {
    const ev = evaluateCariocaContractSlot(runSlot, [card('HEART', 7), card('HEART', 8), joker(), card('HEART', 10)]);
    expect(ev).toEqual({ required: 4, placed: 4, satisfied: true, invalid: false });
  });

  it('reports in-progress when fewer than required cards are placed', () => {
    const ev = evaluateCariocaContractSlot(setSlot, [card('SPADE', 5), card('HEART', 5)]);
    expect(ev).toEqual({ required: 3, placed: 2, satisfied: false, invalid: false });
  });

  it('reports invalid when too many cards are placed', () => {
    const ev = evaluateCariocaContractSlot(setSlot, [card('SPADE', 5), card('HEART', 5), card('DIAMOND', 5), joker()]);
    expect(ev.invalid).toBe(true);
    expect(ev.satisfied).toBe(false);
  });

  it('reports invalid when the count matches but the cards do not form a valid meld', () => {
    const ev = evaluateCariocaContractSlot(setSlot, [card('SPADE', 5), card('HEART', 6), card('DIAMOND', 7)]);
    expect(ev).toEqual({ required: 3, placed: 3, satisfied: false, invalid: true });
  });

  it('reports empty as neither satisfied nor invalid', () => {
    const ev = evaluateCariocaContractSlot(setSlot, []);
    expect(ev).toEqual({ required: 3, placed: 0, satisfied: false, invalid: false });
  });
});

describe('describeCariocaSlotShortfall', () => {
  const setSlot: ContractRummyContractSlot = { kind: CONTRACT_SLOT_SET, size: 3 };
  const runSlot: ContractRummyContractSlot = { kind: CONTRACT_SLOT_RUN, size: 4 };

  it('returns null for a satisfied set', () => {
    expect(describeCariocaSlotShortfall(setSlot, [card('SPADE', 5), card('HEART', 5), joker()])).toBeNull();
  });

  it('returns null for a satisfied run', () => {
    expect(
      describeCariocaSlotShortfall(runSlot, [card('HEART', 7), card('HEART', 8), joker(), card('HEART', 10)]),
    ).toBeNull();
  });

  it('reports empty for a slot with no cards', () => {
    expect(describeCariocaSlotShortfall(setSlot, [])).toEqual({ code: 'empty' });
  });

  it('reports how many more cards a set needs', () => {
    expect(describeCariocaSlotShortfall(setSlot, [card('SPADE', 5)])).toEqual({ code: 'needMoreSet', count: 2 });
  });

  it('reports how many more cards a run needs', () => {
    expect(describeCariocaSlotShortfall(runSlot, [card('SPADE', 2), card('SPADE', 3)])).toEqual({
      code: 'needMoreRun',
      count: 2,
    });
  });

  it('reports excess cards as tooMany', () => {
    expect(
      describeCariocaSlotShortfall(setSlot, [
        card('SPADE', 5),
        card('HEART', 5),
        card('DIAMOND', 5),
        card('CLOVER', 5),
      ]),
    ).toEqual({ code: 'tooMany', count: 1 });
  });

  it('reports a rank mismatch for a full but invalid set', () => {
    expect(describeCariocaSlotShortfall(setSlot, [card('SPADE', 5), card('HEART', 6), card('DIAMOND', 7)])).toEqual({
      code: 'setRankMismatch',
    });
  });

  it('reports a suit mismatch for a full run with mixed suits', () => {
    expect(
      describeCariocaSlotShortfall(runSlot, [card('SPADE', 2), card('HEART', 3), card('SPADE', 4), card('SPADE', 5)]),
    ).toEqual({ code: 'runSuitMismatch' });
  });

  it('reports a broken sequence for a full same-suit run with a gap', () => {
    expect(
      describeCariocaSlotShortfall(runSlot, [card('SPADE', 2), card('SPADE', 3), card('SPADE', 4), card('SPADE', 9)]),
    ).toEqual({ code: 'runNotConsecutive' });
  });
});
