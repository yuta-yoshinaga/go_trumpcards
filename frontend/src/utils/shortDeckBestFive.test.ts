import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import { holdemBestFive } from './holdemBestFive';
import { scoreFiveShortDeck, shortDeckBestFive, shortDeckStraightHigh } from './shortDeckBestFive';

const c = (design: CardDesign, value: number): Card => ({ design, value });

describe('shortDeckStraightHigh', () => {
  it('reads A-6-7-8-9 as a nine-high wheel', () => {
    expect(shortDeckStraightHigh([14, 9, 8, 7, 6])).toBe(9);
  });

  it('reads a normal run by its top card', () => {
    expect(shortDeckStraightHigh([12, 11, 10, 9, 8])).toBe(12);
  });

  it('reads A-10-J-Q-K as broadway', () => {
    expect(shortDeckStraightHigh([14, 13, 12, 11, 10])).toBe(14);
  });

  it('rejects a gap', () => {
    expect(shortDeckStraightHigh([13, 12, 11, 10, 8])).toBeNull();
  });

  it('rejects fewer than five distinct ranks', () => {
    expect(shortDeckStraightHigh([13, 12, 11, 10])).toBeNull();
  });

  it('rejects A-2-3-4-5, which is not a Short Deck hand', () => {
    // Those ranks are not in a 36-card deck, and the standard wheel must not
    // sneak in through the generic run check.
    expect(shortDeckStraightHigh([14, 5, 4, 3, 2])).toBeNull();
  });
});

describe('scoreFiveShortDeck', () => {
  const flush = [c('SPADE', 1), c('SPADE', 7), c('SPADE', 9), c('SPADE', 11), c('SPADE', 13)];
  const fullHouse = [c('SPADE', 8), c('HEART', 8), c('CLOVER', 8), c('SPADE', 10), c('HEART', 10)];

  it('ranks a flush above a full house, unlike standard poker', () => {
    // Mirrors TestShortDeckPlayer_EvalBestHand_FlushBeatsFullHouse.
    expect(scoreFiveShortDeck(flush)[0]).toBeGreaterThan(scoreFiveShortDeck(fullHouse)[0] ?? 0);
  });

  it('still ranks four of a kind above a flush', () => {
    const quads = [c('SPADE', 9), c('HEART', 9), c('CLOVER', 9), c('DIAMOND', 9), c('SPADE', 13)];
    expect(scoreFiveShortDeck(quads)[0]).toBeGreaterThan(scoreFiveShortDeck(flush)[0] ?? 0);
  });

  it('ranks the wheel as the weakest straight', () => {
    const wheel = [c('SPADE', 1), c('HEART', 6), c('CLOVER', 7), c('DIAMOND', 8), c('SPADE', 9)];
    const higher = [c('SPADE', 7), c('HEART', 8), c('CLOVER', 9), c('DIAMOND', 10), c('SPADE', 11)];
    expect(scoreFiveShortDeck(wheel)[0]).toBe(5);
    expect(compareTuples(scoreFiveShortDeck(higher), scoreFiveShortDeck(wheel))).toBeGreaterThan(0);
  });

  it('ranks a straight flush above everything below it', () => {
    const sf = [c('SPADE', 6), c('SPADE', 7), c('SPADE', 8), c('SPADE', 9), c('SPADE', 10)];
    expect(scoreFiveShortDeck(sf)[0]).toBe(9);
  });

  it('orders the low categories as usual', () => {
    const trips = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 7), c('DIAMOND', 9), c('SPADE', 11)];
    const twoPair = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 9), c('DIAMOND', 9), c('SPADE', 11)];
    const onePair = [c('SPADE', 7), c('HEART', 7), c('CLOVER', 9), c('DIAMOND', 10), c('SPADE', 11)];
    const high = [c('SPADE', 6), c('HEART', 8), c('CLOVER', 10), c('DIAMOND', 12), c('SPADE', 1)];
    const cats = [trips, twoPair, onePair, high].map((h) => scoreFiveShortDeck(h)[0] ?? 0);
    expect(cats).toEqual([4, 3, 2, 1]);
  });
});

describe('shortDeckBestFive', () => {
  it('returns null with fewer than five cards', () => {
    expect(shortDeckBestFive([c('SPADE', 7), c('HEART', 8)])).toBeNull();
  });

  it('cannot be asked to choose between a flush and a full house', () => {
    // Seven cards can never hold both: the trips occupy three suits and the pair
    // two, so at most two of them share the flush suit, leaving five more cards
    // needed to complete it — eight in total. The flush/full-house swap therefore
    // only decides showdowns between players, never which five to highlight.
    const cards = [
      c('SPADE', 8),
      c('SPADE', 7),
      c('CLOVER', 8),
      c('HEART', 8),
      c('SPADE', 10),
      c('HEART', 10),
      c('SPADE', 9),
    ];
    const picked = shortDeckBestFive(cards) ?? [];
    expect(picked).toHaveLength(5);
    // Four spades only, so the best available really is the full house.
    expect(scoreFiveShortDeck(picked.map((i) => cards[i] as Card))[0]).toBe(6);
  });

  it('differs from the standard evaluator exactly where it must', () => {
    // The reason this file exists: holdemBestFive does not know A-6-7-8-9, so on
    // this board it highlights an ace-high nothing while the server scores a
    // straight. Five spades never occur here, so the flush/full-house swap is
    // not what drives the difference.
    const cards = [
      c('SPADE', 1),
      c('HEART', 6),
      c('CLOVER', 7),
      c('DIAMOND', 8),
      c('SPADE', 9),
      c('HEART', 12),
      c('CLOVER', 13),
    ];
    const short = shortDeckBestFive(cards) ?? [];
    const standard = holdemBestFive(cards) ?? [];
    expect(scoreFiveShortDeck(short.map((i) => cards[i] as Card))[0]).toBe(5);
    expect([...short].sort()).not.toEqual([...standard].sort());
  });

  it('finds the wheel when it is the only straight', () => {
    const cards = [
      c('SPADE', 1),
      c('HEART', 6),
      c('CLOVER', 7),
      c('DIAMOND', 8),
      c('SPADE', 9),
      c('HEART', 12),
      c('CLOVER', 13),
    ];
    const picked = shortDeckBestFive(cards) ?? [];
    expect(scoreFiveShortDeck(picked.map((i) => cards[i] as Card))[0]).toBe(5);
  });
});

/** Lexicographic comparison mirroring the util's own ordering. */
function compareTuples(a: number[], b: number[]): number {
  for (let i = 0; i < Math.max(a.length, b.length); i += 1) {
    const d = (a[i] ?? 0) - (b[i] ?? 0);
    if (d !== 0) return d;
  }
  return 0;
}
