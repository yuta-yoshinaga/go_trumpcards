import type { Card } from '../types/card';
import { compareScores, scoreFive } from './holdemBestFive';

/** The hole and board indices that make up an Omaha best-five hand. */
export interface OmahaBestFive {
  /** Indices into the hole-card array of the (exactly 2) cards used. */
  holeIdx: number[];
  /** Indices into the board array of the (exactly 3) cards used. */
  boardIdx: number[];
}

/**
 * Choose the best Omaha-style five-card hand under the "must use exactly two
 * hole cards and exactly three board cards" rule. Enumerates every 2-of-hole ×
 * 3-of-board combination, scores each via the shared {@link scoreFive}
 * evaluator, and returns the indices of the winning combination — or `null`
 * when there are fewer than 2 hole cards or 3 board cards.
 *
 * Works for any hole-card count (Omaha = 4, Big O = 5), so it is reused across
 * the Omaha-family pages.
 */
export function omahaBestFive(hole: readonly Card[], board: readonly Card[]): OmahaBestFive | null {
  if (hole.length < 2 || board.length < 3) return null;
  let bestScore: number[] = [];
  let best: OmahaBestFive | null = null;
  const h = hole.length;
  const b = board.length;
  for (let h1 = 0; h1 < h - 1; h1 += 1) {
    for (let h2 = h1 + 1; h2 < h; h2 += 1) {
      for (let b1 = 0; b1 < b - 2; b1 += 1) {
        for (let b2 = b1 + 1; b2 < b - 1; b2 += 1) {
          for (let b3 = b2 + 1; b3 < b; b3 += 1) {
            const score = scoreFive([hole[h1], hole[h2], board[b1], board[b2], board[b3]]);
            if (best === null || compareScores(score, bestScore) > 0) {
              bestScore = score;
              best = { holeIdx: [h1, h2], boardIdx: [b1, b2, b3] };
            }
          }
        }
      }
    }
  }
  return best;
}
