import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  canBeatDoudizhu,
  classifyDoudizhuCombo,
  type DoudizhuCombo,
  doudizhuCardStrength,
  doudizhuInvalidReason,
} from './doudizhuComboValidator';

/** Builds a standard-suit card of the given rank value. */
const c = (value: number, design: Card['design'] = 'SPADE'): Card => ({ design, value });
/** Small joker (value 1) and big joker (value 2). */
const smallJoker: Card = { design: 'JOKER', value: 1 };
const bigJoker: Card = { design: 'JOKER', value: 2 };

describe('doudizhuCardStrength', () => {
  it('orders 3 < K < A < 2 < small joker < big joker', () => {
    expect(doudizhuCardStrength(c(3))).toBe(3);
    expect(doudizhuCardStrength(c(13))).toBe(13);
    expect(doudizhuCardStrength(c(1))).toBe(14); // Ace
    expect(doudizhuCardStrength(c(2))).toBe(15); // 2
    expect(doudizhuCardStrength(smallJoker)).toBe(16);
    expect(doudizhuCardStrength(bigJoker)).toBe(17);
  });
});

describe('classifyDoudizhuCombo', () => {
  it('returns null for an empty selection', () => {
    expect(classifyDoudizhuCombo([])).toBeNull();
  });

  it('classifies a single', () => {
    expect(classifyDoudizhuCombo([c(7)])).toEqual({ type: 'single', rank: 7, length: 1 });
  });

  it('classifies a pair and rejects a mismatched two-card selection', () => {
    expect(classifyDoudizhuCombo([c(9, 'SPADE'), c(9, 'HEART')])).toEqual({ type: 'pair', rank: 9, length: 1 });
    expect(classifyDoudizhuCombo([c(9), c(10)])).toBeNull();
  });

  it('classifies a trio, trio+single, and trio+pair', () => {
    expect(classifyDoudizhuCombo([c(5), c(5), c(5)])).toEqual({ type: 'trio', rank: 5, length: 1 });
    expect(classifyDoudizhuCombo([c(5), c(5), c(5), c(8)])).toEqual({ type: 'trioSingle', rank: 5, length: 1 });
    expect(classifyDoudizhuCombo([c(5), c(5), c(5), c(8), c(8)])).toEqual({ type: 'trioPair', rank: 5, length: 1 });
  });

  it('classifies a bomb and a rocket', () => {
    expect(classifyDoudizhuCombo([c(6), c(6), c(6), c(6)])).toEqual({ type: 'bomb', rank: 6, length: 1 });
    expect(classifyDoudizhuCombo([smallJoker, bigJoker])).toEqual({ type: 'rocket', rank: 17, length: 1 });
  });

  it('classifies a 5-card straight but rejects one containing a 2', () => {
    expect(classifyDoudizhuCombo([c(3), c(4), c(5), c(6), c(7)])).toEqual({ type: 'straight', rank: 3, length: 5 });
    // A(14) and 2(15) cannot be part of a straight
    expect(classifyDoudizhuCombo([c(1), c(13), c(12), c(11), c(10)])).toMatchObject({ type: 'straight' });
    expect(classifyDoudizhuCombo([c(2), c(3), c(4), c(5), c(6)])).toBeNull();
  });

  it('classifies consecutive pairs (3 pairs) but rejects only 2 pairs', () => {
    const threePairs = [c(4), c(4), c(5), c(5), c(6), c(6)];
    expect(classifyDoudizhuCombo(threePairs)).toEqual({ type: 'consecutivePair', rank: 4, length: 3 });
    expect(classifyDoudizhuCombo([c(4), c(4), c(5), c(5)])).toBeNull();
  });

  it('classifies an airplane and airplane with wings', () => {
    // two consecutive trios (555 666)
    expect(classifyDoudizhuCombo([c(5), c(5), c(5), c(6), c(6), c(6)])).toEqual({
      type: 'airplane',
      rank: 5,
      length: 2,
    });
    // airplane + single wings (555 666 + 3 8)
    expect(classifyDoudizhuCombo([c(5), c(5), c(5), c(6), c(6), c(6), c(3), c(8)])).toEqual({
      type: 'airplaneSingle',
      rank: 5,
      length: 2,
    });
    // airplane + pair wings (555 666 + 33 88)
    expect(classifyDoudizhuCombo([c(5), c(5), c(5), c(6), c(6), c(6), c(3), c(3), c(8), c(8)])).toEqual({
      type: 'airplanePair',
      rank: 5,
      length: 2,
    });
  });

  it('rejects a garbage selection', () => {
    expect(classifyDoudizhuCombo([c(3), c(7), c(11)])).toBeNull();
  });
});

/** Classifies cards, throwing if the selection is not a valid combo (test helper). */
function combo(cards: Card[]): DoudizhuCombo {
  const result = classifyDoudizhuCombo(cards);
  if (!result) throw new Error('expected a valid combo');
  return result;
}

describe('canBeatDoudizhu', () => {
  const trio5 = combo([c(5), c(5), c(5)]);
  const trio8 = combo([c(8), c(8), c(8)]);
  const pair9 = combo([c(9), c(9)]);
  const bomb6 = combo([c(6), c(6), c(6), c(6)]);
  const bomb9 = combo([c(9), c(9), c(9), c(9)]);
  const rocket = combo([smallJoker, bigJoker]);

  it('requires same type and higher rank', () => {
    expect(canBeatDoudizhu(trio8, trio5)).toBe(true);
    expect(canBeatDoudizhu(trio5, trio8)).toBe(false);
  });

  it('rejects a different type', () => {
    expect(canBeatDoudizhu(pair9, trio5)).toBe(false);
  });

  it('lets a bomb beat a non-bomb and a higher bomb beat a lower bomb', () => {
    expect(canBeatDoudizhu(bomb6, trio8)).toBe(true);
    expect(canBeatDoudizhu(bomb9, bomb6)).toBe(true);
    expect(canBeatDoudizhu(bomb6, bomb9)).toBe(false);
  });

  it('lets a rocket beat everything, including a bomb', () => {
    expect(canBeatDoudizhu(rocket, bomb9)).toBe(true);
    expect(canBeatDoudizhu(bomb9, rocket)).toBe(false);
  });
});

describe('doudizhuInvalidReason', () => {
  it('returns notCombo for a wrong-count / not-a-combo selection', () => {
    // two mismatched cards do not form a pair
    expect(doudizhuInvalidReason([c(9), c(10)], [])).toBe('notCombo');
  });

  it('returns noBeat for a valid combo that is too low to beat the table', () => {
    const table = [c(8), c(8), c(8)]; // trio of 8
    const selection = [c(5), c(5), c(5)]; // trio of 5 — valid type, too low
    expect(doudizhuInvalidReason(selection, table)).toBe('noBeat');
  });

  it('returns null for a valid beating play', () => {
    const table = [c(5), c(5), c(5)];
    const selection = [c(8), c(8), c(8)];
    expect(doudizhuInvalidReason(selection, table)).toBeNull();
  });

  it('returns null for any valid combo on a fresh lead (empty table)', () => {
    expect(doudizhuInvalidReason([c(3), c(4), c(5), c(6), c(7)], [])).toBeNull();
  });
});
