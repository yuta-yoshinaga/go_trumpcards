import type { Card } from '../types/card';

/** Points threshold at or above which a hand "has Juego". Mirrors `MusJuegoThreshold` in internal/domain/Mus.go. */
export const MUS_JUEGO_THRESHOLD = 31;

/**
 * Pares (pair) category of a Mus hand.
 * `0` = none, `1` = par (one pair), `2` = medias (three of a kind), `3` = duples (two pairs / four of a kind).
 * Mirrors the return values of `musParesCategory` in internal/domain/Mus.go.
 */
export type MusParesCategory = 0 | 1 | 2 | 3;

/**
 * Grande/Chica card rank: A=1..7, Sota(J)=8, Caballo(Q)=9, Rey(K)=10.
 * Used to group cards for Pares detection. Mirrors `musCardRank` in internal/domain/Mus.go.
 */
export function musCardRank(value: number): number {
  switch (value) {
    case 11: // Sota (J)
      return 8;
    case 12: // Caballo (Q)
      return 9;
    case 13: // Rey (K)
      return 10;
    default: // A(1)..7 (and any other value maps to itself)
      return value;
  }
}

/**
 * Juego point value of a single card: A=1, 2..7 = face value, J/Q/K = 10.
 * Mirrors `musCardPoints` in internal/domain/Mus.go.
 */
export function musCardPoints(value: number): number {
  switch (value) {
    case 11:
    case 12:
    case 13:
      return 10;
    default:
      return value;
  }
}

/**
 * Classify the Pares (pair) type of a hand by its card ranks.
 * Mirrors `musParesCategory` in internal/domain/Mus.go: four of a kind or two
 * distinct pairs count as duples, three of a kind as medias, a single pair as par.
 */
export function musParesCategory(cards: readonly Card[]): MusParesCategory {
  const counts = new Map<number, number>();
  for (const c of cards) {
    const r = musCardRank(c.value);
    counts.set(r, (counts.get(r) ?? 0) + 1);
  }
  let pairs = 0;
  let triples = 0;
  let quads = 0;
  for (const count of counts.values()) {
    if (count === 2) pairs++;
    else if (count === 3) triples++;
    else if (count === 4) quads++;
  }
  if (quads > 0 || pairs >= 2) return 3; // duples
  if (triples > 0) return 2; // medias
  if (pairs === 1) return 1; // par
  return 0;
}

/** Sum of the Juego point values of a hand. */
export function musJuegoPoints(cards: readonly Card[]): number {
  let sum = 0;
  for (const c of cards) sum += musCardPoints(c.value);
  return sum;
}

/** Front-end evaluation of the human's Mus hand, used to render a betting-aid summary. */
export interface MusHandEval {
  /** Pares category: 0 none, 1 par, 2 medias, 3 duples. */
  paresCategory: MusParesCategory;
  /** Total Juego points (A=1..7, J/Q/K=10). */
  points: number;
  /** Whether the hand reaches the Juego threshold (>=31); otherwise it plays for Punto. */
  hasJuego: boolean;
  /** Highest card rank held, or null for an empty hand. Grande is bet on this. */
  highestRank: number | null;
  /** Lowest card rank held, or null for an empty hand. Chica is bet on this. */
  lowestRank: number | null;
}

/**
 * Evaluate a Mus hand into its Pares category and Juego points, purely on the
 * client, so the human can see how their hand rates before betting. Mirrors the
 * evaluation rules in internal/domain/Mus.go (READ-ONLY reference; no Go change).
 */
export function evalMusHand(cards: readonly Card[]): MusHandEval {
  const points = musJuegoPoints(cards);
  const ranks = cards.map((c) => musCardRank(c.value));
  return {
    paresCategory: musParesCategory(cards),
    points,
    hasJuego: points >= MUS_JUEGO_THRESHOLD,
    highestRank: ranks.length > 0 ? Math.max(...ranks) : null,
    lowestRank: ranks.length > 0 ? Math.min(...ranks) : null,
  };
}
