import type { Card } from '../types/card';
import { compareScores, scoreFive } from './holdemBestFive';
import { evaluateFiveCardHand, pokerHandKey } from './pokerSquaresUtils';

/** Hole cards each Dramaha seat is dealt (sync: domain.DramahaHoleCards). */
export const DRAMAHA_HOLE_CARDS = 5;

/** The hole and board indices that make up one of Dramaha's two hands. */
export interface DramahaBestFive {
  /** Indices into the hole-card array of the cards used. */
  holeIdx: number[];
  /** Indices into the board array of the cards used. Empty for the draw hand. */
  boardIdx: number[];
}

/** One evaluated half of the split, with the i18n key of its hand name. */
export interface DramahaHand extends DramahaBestFive {
  /** Hand-name key inside the `dramaha` namespace, e.g. `hand.twoPair`. */
  key: string;
}

/**
 * Both halves Dramaha's pot always splits between, evaluated from the same
 * five hole cards. Either is `null` when there are not yet enough cards to
 * decide it.
 */
export interface DramahaHands {
  /** Exactly 2 hole + exactly 3 board (needs a flop). */
  omaha: DramahaHand | null;
  /** The five hole cards as they are — never reads the board. */
  draw: DramahaHand | null;
}

/**
 * Choose the best five-card hand under Dramaha's Omaha rule: exactly two hole
 * cards and exactly three board cards. Enumerates every 2-of-hole × 3-of-board
 * combination, scores each via the shared {@link scoreFive} evaluator, and
 * returns the indices of the winning combination — or `null` when there are
 * fewer than 2 hole cards or 3 board cards.
 *
 * This decides only the Omaha half of the pot. The other half is the draw
 * hand; see {@link dramahaHands}.
 */
export function dramahaBestFive(hole: readonly Card[], board: readonly Card[]): DramahaBestFive | null {
  if (hole.length < 2 || board.length < 3) return null;
  let bestScore: number[] = [];
  let best: DramahaBestFive | null = null;
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

/**
 * Evaluate both halves of a Dramaha holding.
 *
 * The two halves are computed from the same five cards but by different rules,
 * and the difference is the whole game:
 *
 *   - the Omaha half searches 2-of-hole × 3-of-board, so it changes every
 *     street as the board grows;
 *   - the draw half is the five hole cards exactly as they sit. **It never
 *     reads the board.** A board flush does not make the draw hand a flush,
 *     and it needs no flop to be decided — only the five cards.
 *
 * Unlike the Hi-Lo variant this was cloned from, the second half always
 * qualifies: five cards always rank, so `draw` is null only while the seat is
 * holding some other number of cards (mid-deal, or a folded seat).
 *
 * @param hole - The player's hole cards.
 * @param board - The community cards.
 * @returns Both halves; either may be `null` when undecidable.
 */
export function dramahaHands(hole: readonly Card[], board: readonly Card[]): DramahaHands {
  return { omaha: omahaHalf(hole, board), draw: drawHalf(hole) };
}

/** The Omaha half: the best 2-hole + 3-board five, named. */
function omahaHalf(hole: readonly Card[], board: readonly Card[]): DramahaHand | null {
  const best = dramahaBestFive(hole, board);
  if (!best) return null;
  const five = [...best.holeIdx.map((i) => hole[i]), ...best.boardIdx.map((i) => board[i])];
  const rank = evaluateFiveCardHand(five);
  return rank == null ? null : { ...best, key: pokerHandKey(rank) };
}

/**
 * The draw half: the five hole cards as dealt, named.
 *
 * Takes no board argument on purpose — there is no board reading to get wrong.
 */
function drawHalf(hole: readonly Card[]): DramahaHand | null {
  if (hole.length !== DRAMAHA_HOLE_CARDS) return null;
  const rank = evaluateFiveCardHand(hole);
  if (rank == null) return null;
  return { holeIdx: [0, 1, 2, 3, 4], boardIdx: [], key: pokerHandKey(rank) };
}
