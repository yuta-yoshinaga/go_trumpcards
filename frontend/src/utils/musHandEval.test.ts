import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import {
  evalMusHand,
  MUS_JUEGO_THRESHOLD,
  musCardPoints,
  musCardRank,
  musJuegoPoints,
  musParesCategory,
} from './musHandEval';

/** Build a card of the given value (suit is irrelevant to Mus evaluation). */
const c = (value: number, design: Card['design'] = 'SPADE'): Card => ({ design, value });

describe('musCardRank', () => {
  it('maps A(1)..7 to themselves', () => {
    expect(musCardRank(1)).toBe(1);
    expect(musCardRank(7)).toBe(7);
  });
  it('maps face cards J/Q/K to 8/9/10', () => {
    expect(musCardRank(11)).toBe(8);
    expect(musCardRank(12)).toBe(9);
    expect(musCardRank(13)).toBe(10);
  });
});

describe('musCardPoints', () => {
  it('scores A..7 at face value', () => {
    expect(musCardPoints(1)).toBe(1);
    expect(musCardPoints(7)).toBe(7);
  });
  it('scores J/Q/K as 10', () => {
    expect(musCardPoints(11)).toBe(10);
    expect(musCardPoints(12)).toBe(10);
    expect(musCardPoints(13)).toBe(10);
  });
});

describe('musJuegoPoints', () => {
  it('sums the point values of the hand', () => {
    expect(musJuegoPoints([c(13), c(13), c(13), c(1)])).toBe(31); // 10+10+10+1
    expect(musJuegoPoints([c(1), c(12), c(12), c(7)])).toBe(28); // 1+10+10+7
  });
});

describe('musParesCategory', () => {
  it('returns 0 (none) when all ranks differ', () => {
    expect(musParesCategory([c(1), c(2), c(3), c(4)])).toBe(0);
  });
  it('returns 1 (par) for a single pair', () => {
    expect(musParesCategory([c(12), c(12), c(1), c(7)])).toBe(1);
  });
  it('returns 2 (medias) for three of a kind', () => {
    expect(musParesCategory([c(5), c(5), c(5), c(2)])).toBe(2);
  });
  it('returns 3 (duples) for two pairs', () => {
    expect(musParesCategory([c(3), c(3), c(7), c(7)])).toBe(3);
  });
  it('returns 3 (duples) for four of a kind', () => {
    expect(musParesCategory([c(4), c(4), c(4), c(4)])).toBe(3);
  });
  it('treats face cards of equal rank as a pair (J is not K)', () => {
    // Two Kings (13,13) form a pair; the two others (11,12) are distinct ranks.
    expect(musParesCategory([c(13), c(13), c(11), c(12)])).toBe(1);
  });
});

describe('evalMusHand', () => {
  it('flags Juego with the best hand at 31 points', () => {
    const e = evalMusHand([c(13), c(13), c(13), c(1)]);
    expect(e.points).toBe(31);
    expect(e.hasJuego).toBe(true);
    expect(e.paresCategory).toBe(2); // three Kings = medias
  });

  it('flags Juego at 40 points (four ten-value cards)', () => {
    const e = evalMusHand([c(13), c(12), c(11), c(13)]);
    expect(e.points).toBe(40);
    expect(e.hasJuego).toBe(true);
    expect(e.points).toBeGreaterThanOrEqual(MUS_JUEGO_THRESHOLD);
  });

  it('reports Punto (no Juego) below the threshold', () => {
    const e = evalMusHand([c(1), c(12), c(12), c(7)]);
    expect(e.points).toBe(28);
    expect(e.hasJuego).toBe(false);
    expect(e.paresCategory).toBe(1); // pair of Queens = par
  });

  it('reports no pares and no juego for a low scattered hand', () => {
    const e = evalMusHand([c(1), c(2), c(4), c(6)]);
    expect(e.paresCategory).toBe(0);
    expect(e.points).toBe(13);
    expect(e.hasJuego).toBe(false);
  });
});
