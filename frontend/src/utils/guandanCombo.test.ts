import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { GUANDAN_COMBO, guandanEvaluate, guandanIsWild } from './guandanCombo';

const c = (design: CardDesign, value: number): Card => ({ design, value });
const s = (v: number) => c('SPADE', v);
const h = (v: number) => c('HEART', v);
const d = (v: number) => c('DIAMOND', v);
const cl = (v: number) => c('CLOVER', v);
const blackJoker = c('JOKER', 1);
const redJoker = c('JOKER', 2);

// Level 5 throughout unless a case is specifically about the level.
const LV = 5;

describe('guandanIsWild', () => {
  it('is only the level card in hearts', () => {
    expect(guandanIsWild(h(5), 5)).toBe(true);
    expect(guandanIsWild(s(5), 5)).toBe(false);
    expect(guandanIsWild(h(6), 5)).toBe(false);
    expect(guandanIsWild(redJoker, 5)).toBe(false);
  });
});

describe('guandanEvaluate — same-rank combos', () => {
  it('reads one card as a single', () => {
    expect(guandanEvaluate([s(7)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Single, size: 1 });
  });

  it('reads two of a rank as a pair', () => {
    expect(guandanEvaluate([s(7), h(7)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Pair, size: 2 });
  });

  it('reads three of a rank as a triple', () => {
    expect(guandanEvaluate([s(7), h(7), d(7)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Triple, size: 3 });
  });

  it('reads four of a rank as a bomb, and a fifth makes it bigger', () => {
    expect(guandanEvaluate([s(7), h(7), d(7), cl(7)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Bomb, size: 4 });
    expect(guandanEvaluate([s(7), h(7), d(7), cl(7), s(7)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Bomb, size: 5 });
  });

  it('lets a wild stand in for the missing card of a pair', () => {
    // The heart 5 is the level card, so it plays as a second seven.
    expect(guandanEvaluate([s(7), h(LV)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Pair, size: 2 });
  });

  it('does not treat the off-suit level card as wild', () => {
    // Spade 5 is the level card but not a heart, so this is 7 + 5, not a pair.
    expect(guandanEvaluate([s(7), s(LV)], LV)).toBeNull();
  });

  it('reads four jokers as the joker bomb', () => {
    const combo = guandanEvaluate([blackJoker, blackJoker, redJoker, redJoker], LV);
    expect(combo).toMatchObject({ kind: GUANDAN_COMBO.JokerBomb, size: 4 });
  });

  it('does not let a wild complete a joker bomb', () => {
    // Three jokers plus the wild is not the joker bomb; four of a rank is a
    // plain bomb, and jokers do not share a rank, so this is nothing.
    expect(guandanEvaluate([blackJoker, redJoker, redJoker, h(LV)], LV)?.kind).not.toBe(GUANDAN_COMBO.JokerBomb);
  });
});

describe('guandanEvaluate — five-card combos', () => {
  it('reads a full house', () => {
    expect(guandanEvaluate([s(7), h(7), d(7), s(9), h(9)], LV)).toMatchObject({
      kind: GUANDAN_COMBO.FullHouse,
      size: 5,
    });
  });

  it('reads a straight', () => {
    expect(guandanEvaluate([s(6), h(7), d(8), cl(9), s(10)], LV)).toMatchObject({
      kind: GUANDAN_COMBO.Straight,
      size: 5,
    });
  });

  it('reads the ace low, as A-2-3-4-5', () => {
    expect(guandanEvaluate([s(1), h(2), d(3), cl(4), s(6)], LV)).toBeNull();
    expect(guandanEvaluate([s(1), h(2), d(3), cl(4), s(5)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Straight });
  });

  it('reads a same-suit straight as a straight flush', () => {
    expect(guandanEvaluate([s(6), s(7), s(8), s(9), s(10)], LV)).toMatchObject({
      kind: GUANDAN_COMBO.StraightFlush,
      size: 5,
    });
  });

  it('demotes a wild-completed same-suit run to a plain straight', () => {
    // The wild has no suit of its own, so the run is not a flush.
    expect(guandanEvaluate([s(6), s(7), s(8), s(9), h(LV)], LV)).toMatchObject({ kind: GUANDAN_COMBO.Straight });
  });

  it('rejects a run that a joker is sitting in', () => {
    expect(guandanEvaluate([s(6), h(7), d(8), cl(9), redJoker], LV)).toBeNull();
  });
});

describe('guandanEvaluate — six-card runs', () => {
  it('reads two consecutive triples as a plate', () => {
    expect(guandanEvaluate([s(7), h(7), d(7), s(8), h(8), d(8)], LV)).toMatchObject({
      kind: GUANDAN_COMBO.Plate,
      size: 6,
    });
  });

  it('reads three consecutive pairs as a tube', () => {
    expect(guandanEvaluate([s(7), h(7), s(8), d(8), s(9), h(9)], LV)).toMatchObject({
      kind: GUANDAN_COMBO.Tube,
      size: 6,
    });
  });

  it('counts consecutive pairs at the level card natural position', () => {
    // Level 5: the 5s rank above the ace for strength, but 4-5-6 is still a
    // legal tube. Ranking them at 15 would blow the window apart.
    expect(guandanEvaluate([s(4), d(4), s(LV), cl(LV), s(6), d(6)], LV)).toMatchObject({
      kind: GUANDAN_COMBO.Tube,
      size: 6,
    });
  });
});

describe('guandanEvaluate — nothing', () => {
  it('is null for an empty selection', () => {
    expect(guandanEvaluate([], LV)).toBeNull();
  });

  it('is null for two unrelated cards', () => {
    expect(guandanEvaluate([s(7), h(9)], LV)).toBeNull();
  });

  it('is null for four cards that form no combo', () => {
    expect(guandanEvaluate([s(2), h(6), d(9), cl(13)], LV)).toBeNull();
  });
});
