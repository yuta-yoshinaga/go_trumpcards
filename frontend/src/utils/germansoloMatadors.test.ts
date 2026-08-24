import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { MATADOR_NAME_KEY, matadorRank } from './germansoloMatadors';

const card = (design: Card['design'], value: number): Card => ({ design, value });

describe('matadorRank', () => {
  it('returns null when trump is undecided (unset or none)', () => {
    expect(matadorRank(card('CLOVER', 12), -1)).toBeNull();
    expect(matadorRank(card('CLOVER', 12), 0)).toBeNull();
  });

  it('identifies Spadille (♣Q) as rank 1 for every trump suit', () => {
    for (const trump of [1, 2, 3, 4]) {
      expect(matadorRank(card('CLOVER', 12), trump)).toBe(1);
    }
  });

  it('identifies Basta (♠Q) as rank 3 for every trump suit', () => {
    for (const trump of [1, 2, 3, 4]) {
      expect(matadorRank(card('SPADE', 12), trump)).toBe(3);
    }
  });

  it('identifies Manille as the 7 of the trump suit (all four suit branches)', () => {
    expect(matadorRank(card('SPADE', 7), 1)).toBe(2); // spade trump → ♠7
    expect(matadorRank(card('CLOVER', 7), 2)).toBe(2); // club trump → ♣7
    expect(matadorRank(card('HEART', 7), 3)).toBe(2); // heart trump → ♥7
    expect(matadorRank(card('DIAMOND', 7), 4)).toBe(2); // diamond trump → ♦7
  });

  // **黒の Q は自分のスートが切り札でも Spadille / Basta のまま。** ここを
  // 「切り札スートの Q は普通の切り札」と読むと、♣ が切り札のときだけ最強札が
  // 消える。
  it('keeps the black queens as matadors even when their own suit is trumps', () => {
    expect(matadorRank(card('CLOVER', 12), 2)).toBe(1); // clubs are trumps, ♣Q is still Spadille
    expect(matadorRank(card('SPADE', 12), 1)).toBe(3); // spades are trumps, ♠Q is still Basta
  });

  it('does not mark a 7 of a non-trump suit as Manille', () => {
    expect(matadorRank(card('HEART', 7), 1)).toBeNull(); // trump is spades
    expect(matadorRank(card('DIAMOND', 7), 2)).toBeNull(); // trump is clubs
  });

  it('returns null for ordinary cards', () => {
    expect(matadorRank(card('HEART', 12), 1)).toBeNull(); // red queen is an ordinary card
    expect(matadorRank(card('DIAMOND', 1), 1)).toBeNull(); // aces are not matadors here
    expect(matadorRank(card('HEART', 1), 3)).toBeNull(); // not even the trump-suit ace
    expect(matadorRank(card('SPADE', 1), 1)).toBeNull(); // ♠A is Quadrille's Spadille, not this game's
  });

  it('exposes an i18n name key for each rank', () => {
    expect(MATADOR_NAME_KEY[1]).toBe('matador.spadille');
    expect(MATADOR_NAME_KEY[2]).toBe('matador.manille');
    expect(MATADOR_NAME_KEY[3]).toBe('matador.basta');
  });
});
