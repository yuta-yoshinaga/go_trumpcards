import { describe, expect, it } from 'vitest';
import type { Card, CardDesign } from '../types/card';
import type { DoubleKlondikeTableauCard } from '../types/games/doubleklondike';
import { doubleKlondikeCanPlaceOnFoundation, doubleKlondikeCanPlaceOnTableau } from './doubleKlondikeTargets';

const c = (design: CardDesign, value: number): Card => ({ design, value });
const up = (design: CardDesign, value: number): DoubleKlondikeTableauCard => ({ card: c(design, value), faceUp: true });
const down = (): DoubleKlondikeTableauCard => ({ card: null, faceUp: false });

describe('doubleKlondikeCanPlaceOnTableau', () => {
  const tableau = [[up('SPADE', 8)], [], [down()], [up('HEART', 8)]];

  it('takes a red card one rank below a black top', () => {
    expect(doubleKlondikeCanPlaceOnTableau(c('HEART', 7), tableau, 0)).toBe(true);
  });

  it('refuses the same colour', () => {
    expect(doubleKlondikeCanPlaceOnTableau(c('CLOVER', 7), tableau, 0)).toBe(false);
  });

  it('refuses a rank that does not descend by one', () => {
    expect(doubleKlondikeCanPlaceOnTableau(c('HEART', 6), tableau, 0)).toBe(false);
    expect(doubleKlondikeCanPlaceOnTableau(c('HEART', 9), tableau, 0)).toBe(false);
  });

  it('accepts only a King on an empty column', () => {
    expect(doubleKlondikeCanPlaceOnTableau(c('SPADE', 13), tableau, 1)).toBe(true);
    expect(doubleKlondikeCanPlaceOnTableau(c('SPADE', 12), tableau, 1)).toBe(false);
  });

  it('refuses a column whose top is face down', () => {
    expect(doubleKlondikeCanPlaceOnTableau(c('HEART', 7), tableau, 2)).toBe(false);
  });

  it('refuses a column that does not exist', () => {
    expect(doubleKlondikeCanPlaceOnTableau(c('HEART', 7), tableau, 99)).toBe(false);
  });
});

describe('doubleKlondikeCanPlaceOnFoundation', () => {
  // Two decks: eight piles, so the same card can be legal on more than one.
  const foundation = [[c('SPADE', 1)], [], [c('SPADE', 1)], [c('HEART', 1), c('HEART', 2)]];

  it('accepts only an Ace on an empty pile', () => {
    expect(doubleKlondikeCanPlaceOnFoundation(c('DIAMOND', 1), foundation, 1)).toBe(true);
    expect(doubleKlondikeCanPlaceOnFoundation(c('DIAMOND', 2), foundation, 1)).toBe(false);
  });

  it('builds up in the same suit', () => {
    expect(doubleKlondikeCanPlaceOnFoundation(c('SPADE', 2), foundation, 0)).toBe(true);
    expect(doubleKlondikeCanPlaceOnFoundation(c('HEART', 2), foundation, 0)).toBe(false);
    expect(doubleKlondikeCanPlaceOnFoundation(c('SPADE', 3), foundation, 0)).toBe(false);
  });

  it('is legal on either of the two piles holding the same suit', () => {
    expect(doubleKlondikeCanPlaceOnFoundation(c('SPADE', 2), foundation, 0)).toBe(true);
    expect(doubleKlondikeCanPlaceOnFoundation(c('SPADE', 2), foundation, 2)).toBe(true);
  });

  it('refuses a pile that does not exist', () => {
    expect(doubleKlondikeCanPlaceOnFoundation(c('SPADE', 2), foundation, 99)).toBe(false);
  });
});
