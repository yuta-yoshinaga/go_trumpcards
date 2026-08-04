import type { Card } from '../types/card';
import { omahaBestFive } from './omahaBestFive';
import { evaluateFiveCardHand, pokerHandKey } from './pokerSquaresUtils';

/**
 * i18n key of the best hand the player currently holds under Omaha's
 * must-use-exactly-two-hole-cards rule, or null when no five-card hand exists
 * yet (pre-flop) or the cards do not evaluate.
 *
 * The preview matters more the wider the hole set gets: Big O deals five hole
 * cards, which is ten two-card combinations to weigh by eye every street.
 * @param hole - The player's hole cards.
 * @param board - The community cards.
 * @returns The hand-name i18n key, or null.
 */
export function omahaLiveHandKey(hole: readonly Card[], board: readonly Card[]): string | null {
  const best = omahaBestFive(hole, board);
  if (!best) return null;
  // omahaBestFive only returns a combination when it found five real cards, so
  // the indices always resolve and evaluateFiveCardHand always scores.
  const five = [...best.holeIdx.map((i) => hole[i]), ...best.boardIdx.map((i) => board[i])];
  const rank = evaluateFiveCardHand(five);
  return rank == null ? null : pokerHandKey(rank);
}
