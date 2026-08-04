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

/** The subset of a poker player this preview needs. */
export interface LivePreviewPlayer {
  cards?: Card[];
  folded?: boolean;
}

/**
 * The preview key to show for a player mid-hand, or null when no preview
 * belongs on screen: before the hand is live, at showdown (where the final
 * hand is already highlighted), or for a player who has folded or is absent.
 *
 * The guard lives here rather than in each page so the three Omaha variants
 * cannot disagree about when a preview is appropriate.
 * @param player - The human player, if seated.
 * @param board - The community cards.
 * @param opts - Whether the hand is live and whether it has reached showdown.
 * @returns The hand-name i18n key, or null.
 */
export function omahaLivePreviewKey(
  player: LivePreviewPlayer | null | undefined,
  board: readonly Card[],
  opts: { isActive: boolean; isShowdown: boolean },
): string | null {
  if (!opts.isActive || opts.isShowdown) return null;
  if (!player || player.folded) return null;
  return omahaLiveHandKey(player.cards ?? [], board);
}
