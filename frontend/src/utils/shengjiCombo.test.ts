import { describe, expect, it } from 'vitest';
import golden from '../constants/shengjiCombo.json';
import type { Card, CardDesign } from '../types/card';
import { SHENGJI_COMBO, shengjiEvaluate } from './shengjiCombo';

const card = (design: CardDesign, value: number): Card => ({ design, value });
const spade = (value: number) => card('SPADE', value);
const heart = (value: number) => card('HEART', value);
const joker = (value: number) => card('JOKER', value);

const LEVEL = 5;
const TRUMP_SUIT = 1;

describe('shengjiEvaluate', () => {
  it('matches the shared Go golden fixture', () => {
    for (const goldenCase of golden) {
      const cards = goldenCase.cards.map(({ design, value }) => card(design as CardDesign, value));
      const got = shengjiEvaluate(cards, goldenCase.level, goldenCase.trumpSuit);
      if (goldenCase.kind === SHENGJI_COMBO.None) {
        expect(got, goldenCase.name).toBeNull();
      } else {
        expect(got, goldenCase.name).toEqual({
          kind: goldenCase.kind,
          rank: goldenCase.rank,
          size: goldenCase.size,
          trump: goldenCase.trump,
        });
      }
    }
  });

  it('classifies a single', () => {
    expect(shengjiEvaluate([spade(7)], LEVEL, TRUMP_SUIT)).toMatchObject({
      kind: SHENGJI_COMBO.Single,
      size: 1,
    });
  });

  it('requires two copies of the same card for a pair', () => {
    expect(shengjiEvaluate([spade(7), spade(7)], LEVEL, TRUMP_SUIT)).toMatchObject({
      kind: SHENGJI_COMBO.Pair,
      size: 2,
      trump: true,
    });
    expect(shengjiEvaluate([spade(7), heart(7)], LEVEL, TRUMP_SUIT)).toBeNull();
  });

  it('classifies consecutive same-card pairs as a tractor', () => {
    expect(shengjiEvaluate([spade(4), spade(4), spade(6), spade(6)], LEVEL, TRUMP_SUIT)).toMatchObject({
      kind: SHENGJI_COMBO.Tractor,
      size: 4,
    });
  });

  it('rejects non-consecutive pairs', () => {
    expect(shengjiEvaluate([spade(2), spade(2), spade(4), spade(4)], LEVEL, TRUMP_SUIT)).toBeNull();
  });

  it('uses the domain sequence where the level card is removed', () => {
    expect(shengjiEvaluate([spade(4), spade(4), spade(6), spade(6)], LEVEL, TRUMP_SUIT)).not.toBeNull();
  });

  it('treats level cards and jokers as one trump group', () => {
    expect(shengjiEvaluate([heart(LEVEL)], LEVEL, TRUMP_SUIT)).toMatchObject({
      kind: SHENGJI_COMBO.Single,
    });
    expect(shengjiEvaluate([joker(1), joker(2)], LEVEL, TRUMP_SUIT)).toBeNull();
  });
});
