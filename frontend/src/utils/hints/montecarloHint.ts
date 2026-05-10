import type { Card, MonteCarloResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { MonteCarloPhase } from '../../types/phases';

/** Board dimension (5x5). */
const SIZE = 5;

/**
 * Returns a Monte Carlo Solitaire hint:
 * - "remove-r1-c1-r2-c2" when an adjacent same-rank pair exists
 * - "deal" when no pair exists but the stock or compression can produce one
 * - null when the game is over or genuinely stuck
 *
 * Mirrors the backend's row-major scan + 4-direction "successor cells only"
 * adjacency check so each pair is evaluated exactly once.
 */
export function getMonteCarloHint(state: MonteCarloResponse): HintResult | null {
  if (state.phase !== MonteCarloPhase.PLAYING) return null;

  const pair = findFirstAdjacentPair(state);
  if (pair) {
    const { r1, c1, r2, c2 } = pair;
    return {
      targetAction: `remove-${r1}-${c1}-${r2}-${c2}`,
      reason: 'hint.removePair',
      confidence: 'strong',
    };
  }

  if (state.stockCount > 0 || hasCompressionGap(state)) {
    return {
      targetAction: 'deal',
      reason: 'hint.deal',
      confidence: 'moderate',
    };
  }

  return null;
}

interface Pair {
  r1: number;
  c1: number;
  r2: number;
  c2: number;
}

function findFirstAdjacentPair(state: MonteCarloResponse): Pair | null {
  // Successor directions only — right, down-left, down, down-right.
  // Mirrors the backend so the same pair is reported.
  const dirs: Array<[number, number]> = [
    [0, 1],
    [1, -1],
    [1, 0],
    [1, 1],
  ];
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      const a: Card | null | undefined = state.board[r]?.[c]?.card;
      if (!a) continue;
      for (const [dr, dc] of dirs) {
        const nr = r + dr;
        const nc = c + dc;
        if (nr < 0 || nr >= SIZE || nc < 0 || nc >= SIZE) continue;
        const b: Card | null | undefined = state.board[nr]?.[nc]?.card;
        if (!b) continue;
        if (a.value === b.value) return { r1: r, c1: c, r2: nr, c2: nc };
      }
    }
  }
  return null;
}

function hasCompressionGap(state: MonteCarloResponse): boolean {
  let seenNil = false;
  for (let r = 0; r < SIZE; r++) {
    for (let c = 0; c < SIZE; c++) {
      const card = state.board[r]?.[c]?.card;
      if (!card) {
        seenNil = true;
        continue;
      }
      if (seenNil) return true;
    }
  }
  return false;
}
