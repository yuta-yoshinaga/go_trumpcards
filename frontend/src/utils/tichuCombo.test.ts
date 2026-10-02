import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { classifyTichuCombo, compareTichuCombos } from './tichuCombo';

/** A normal suited card. */
const c = (design: Card['design'], value: number): Card => ({ design, value });
/** Tichu special cards (JOKER design): 1=Mahjong, 2=Dog, 3=Phoenix, 4=Dragon. */
const MAHJONG = c('JOKER', 1);
const DOG = c('JOKER', 2);
const PHOENIX = c('JOKER', 3);
const DRAGON = c('JOKER', 4);
const mixedStraight = (values: number[]): Card[] =>
  values.map((value, index) => c((['SPADE', 'HEART', 'CLOVER', 'DIAMOND'] as const)[index % 4], value));

describe('classifyTichuCombo', () => {
  it('returns invalid for an empty selection', () => {
    expect(classifyTichuCombo([])).toEqual({ type: 'invalid', length: 0 });
  });

  it('classifies a single normal card', () => {
    expect(classifyTichuCombo([c('SPADE', 7)])).toEqual({ type: 'single', length: 0 });
  });

  it('classifies the Mahjong as a single', () => {
    expect(classifyTichuCombo([MAHJONG]).type).toBe('single');
  });

  it('classifies the Dragon alone as a single', () => {
    expect(classifyTichuCombo([DRAGON]).type).toBe('single');
  });

  it('rejects the Dragon combined with other cards', () => {
    expect(classifyTichuCombo([DRAGON, c('SPADE', 7)]).type).toBe('invalid');
  });

  it('classifies the Dog alone as a dog lead', () => {
    expect(classifyTichuCombo([DOG])).toEqual({ type: 'dog', length: 0 });
  });

  it('rejects the Dog combined with other cards', () => {
    expect(classifyTichuCombo([DOG, c('SPADE', 7)]).type).toBe('invalid');
  });

  it('classifies a pair', () => {
    expect(classifyTichuCombo([c('SPADE', 9), c('HEART', 9)])).toEqual({ type: 'pair', length: 0 });
  });

  it('classifies a phoenix pair', () => {
    expect(classifyTichuCombo([PHOENIX, c('HEART', 9)]).type).toBe('pair');
  });

  it('rejects a phoenix paired with the Mahjong', () => {
    expect(classifyTichuCombo([PHOENIX, MAHJONG]).type).toBe('invalid');
  });

  it('rejects a mismatched two-card selection', () => {
    expect(classifyTichuCombo([c('SPADE', 9), c('HEART', 8)]).type).toBe('invalid');
  });

  it('classifies a triple', () => {
    expect(classifyTichuCombo([c('SPADE', 5), c('HEART', 5), c('CLOVER', 5)]).type).toBe('triple');
  });

  it('classifies a phoenix triple', () => {
    expect(classifyTichuCombo([c('SPADE', 5), c('HEART', 5), PHOENIX]).type).toBe('triple');
  });

  it('rejects a mismatched three-card selection', () => {
    expect(classifyTichuCombo([c('SPADE', 5), c('HEART', 5), c('CLOVER', 6)]).type).toBe('invalid');
  });

  it('classifies a full house (3 + 2)', () => {
    const cards = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('SPADE', 4), c('HEART', 4)];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'fullHouse', length: 0 });
  });

  it('classifies a phoenix full house from a triple + single', () => {
    const cards = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('SPADE', 4), PHOENIX];
    expect(classifyTichuCombo(cards).type).toBe('fullHouse');
  });

  it('classifies a phoenix full house from two pairs', () => {
    const cards = [c('SPADE', 7), c('HEART', 7), c('SPADE', 4), c('HEART', 4), PHOENIX];
    expect(classifyTichuCombo(cards).type).toBe('fullHouse');
  });

  it('classifies a straight of five', () => {
    const cards = [c('SPADE', 5), c('HEART', 6), c('CLOVER', 7), c('DIAMOND', 8), c('SPADE', 9)];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'straight', length: 5 });
  });

  it('classifies a phoenix straight that fills a gap', () => {
    const cards = [c('SPADE', 5), c('HEART', 6), c('DIAMOND', 8), c('SPADE', 9), PHOENIX];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'straight', length: 5 });
  });

  it('classifies a phoenix straight that extends an end', () => {
    const cards = [c('SPADE', 5), c('HEART', 6), c('CLOVER', 7), c('DIAMOND', 8), PHOENIX];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'straight', length: 5 });
  });

  it('classifies stairs (consecutive pairs)', () => {
    const cards = [c('SPADE', 5), c('HEART', 5), c('CLOVER', 6), c('DIAMOND', 6)];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'stairs', length: 4 });
  });

  it('classifies phoenix stairs', () => {
    const cards = [c('SPADE', 5), c('HEART', 5), c('CLOVER', 6), c('DIAMOND', 6), c('SPADE', 7), PHOENIX];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'stairs', length: 6 });
  });

  it('classifies a four-of-a-kind bomb', () => {
    const cards = [c('SPADE', 8), c('HEART', 8), c('CLOVER', 8), c('DIAMOND', 8)];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'bomb', length: 4 });
  });

  it('classifies a straight flush', () => {
    const cards = [c('SPADE', 5), c('SPADE', 6), c('SPADE', 7), c('SPADE', 8), c('SPADE', 9)];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'straightFlush', length: 5 });
  });

  it('rejects a selection with two phoenixes', () => {
    expect(classifyTichuCombo([PHOENIX, PHOENIX, c('SPADE', 5)]).type).toBe('invalid');
  });

  it('rejects a four-card selection that forms no combo', () => {
    const cards = [c('SPADE', 5), c('HEART', 6), c('CLOVER', 9), c('DIAMOND', 12)];
    expect(classifyTichuCombo(cards).type).toBe('invalid');
  });

  it('treats an Ace as the highest natural rank in a straight', () => {
    // 10-J-Q-K-A must be consecutive (Ace = 14).
    const cards = [c('SPADE', 10), c('HEART', 11), c('CLOVER', 12), c('DIAMOND', 13), c('SPADE', 1)];
    expect(classifyTichuCombo(cards)).toEqual({ type: 'straight', length: 5 });
  });
});

describe('compareTichuCombos', () => {
  it('matches the domain TichuCanBeat single comparison cases', () => {
    expect(compareTichuCombos([c('SPADE', 9)], [c('SPADE', 7)])).toBe('beats');
    expect(compareTichuCombos([c('SPADE', 7)], [c('SPADE', 9)])).toBe('cannotBeat');
    expect(compareTichuCombos([DRAGON], [c('SPADE', 1)])).toBe('beats');
    expect(compareTichuCombos([PHOENIX], [c('SPADE', 9)])).toBe('beats');
    expect(compareTichuCombos([PHOENIX], [DRAGON])).toBe('cannotBeat');
    expect(compareTichuCombos([PHOENIX], [c('SPADE', 13)])).toBe('beats');
    expect(compareTichuCombos([c('SPADE', 1)], [PHOENIX])).toBe('unknown');
  });

  it('compares bombs and straight flushes as the domain does', () => {
    const pair = [c('SPADE', 13), c('HEART', 13)];
    const bomb = [c('SPADE', 3), c('HEART', 3), c('CLOVER', 3), c('DIAMOND', 3)];
    const straightFlush = [3, 4, 5, 6, 7].map((value) => c('SPADE', value));
    expect(compareTichuCombos(bomb, pair)).toBe('beats');
    expect(compareTichuCombos(pair, bomb)).toBe('cannotBeat');
    expect(compareTichuCombos(straightFlush, bomb)).toBe('beats');
    expect(compareTichuCombos(bomb, straightFlush)).toBe('cannotBeat');
  });

  it('matches the domain type and length mismatch cases', () => {
    expect(compareTichuCombos([c('SPADE', 1)], [c('SPADE', 13), c('HEART', 13)])).toBe('incomparable');
    const straight5 = mixedStraight([3, 4, 5, 6, 7]);
    const straight6 = mixedStraight([3, 4, 5, 6, 7, 8]);
    expect(compareTichuCombos(straight6, straight5)).toBe('incomparable');
  });

  it('uses the completed top rank of a phoenix straight', () => {
    const natural = mixedStraight([2, 3, 4, 5, 6]);
    const phoenixExtends = [3, 4, 5, 6].map((value) => c('HEART', value)).concat(PHOENIX);
    const phoenixFillsGap = [3, 4, 6, 7].map((value) => c('CLOVER', value)).concat(PHOENIX);
    expect(compareTichuCombos(phoenixExtends, natural)).toBe('beats');
    const natural7 = mixedStraight([3, 4, 5, 6, 7]);
    expect(compareTichuCombos(phoenixFillsGap, natural7)).toBe('cannotBeat');
  });

  it('keeps the Dog outside comparisons', () => {
    expect(compareTichuCombos([DOG], [c('SPADE', 1)])).toBe('incomparable');
  });

  it('compares same-type plays, leaves unlike plays incomparable, and applies bomb ranking', () => {
    expect(compareTichuCombos([c('SPADE', 10)], [c('HEART', 9)])).toBe('beats');
    expect(compareTichuCombos([c('SPADE', 7), c('HEART', 7)], [c('SPADE', 6)])).toBe('incomparable');
    expect(compareTichuCombos([c('SPADE', 7), c('HEART', 7)], [c('SPADE', 6), c('HEART', 6)])).toBe('beats');
    expect(compareTichuCombos([c('SPADE', 8), c('HEART', 8), c('CLOVER', 8), c('DIAMOND', 8)], [c('SPADE', 1)])).toBe(
      'beats',
    );
    expect(compareTichuCombos([DOG], [c('SPADE', 1)])).toBe('incomparable');
  });
});
