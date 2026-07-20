import type { Card, OasisPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { OasisPokerPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Card value representing an Ace. */
const ACE_VALUE = 1;

/** Card value representing a King. */
const KING_VALUE = 13;

/**
 * Returns an Oasis Poker hint for the exchange and action phases.
 *
 * Oasis Poker is a Caribbean Stud variant where the player may buy a card
 * exchange before the play/fold decision. The heuristic is a simplified basic
 * strategy (a suggestion, not a guarantee):
 * - Exchange phase: keep (stand) a made hand of a pair or better; otherwise
 *   exchange weak cards to try to improve, since only a made hand or Ace-King
 *   beats a qualifying dealer.
 * - Action phase: play with a pair or better, or with Ace-King high; fold otherwise.
 */
export function getOasisPokerHint(state: OasisPokerResponse): HintResult | null {
  if (!state.playerHand || state.playerHand.length === 0) return null;

  if (state.phase === OasisPokerPhase.EXCHANGE) {
    if (state.playerHandRank >= RANK_ONE_PAIR) {
      return { targetAction: 'stand', reason: 'hint.exchangeKeep', confidence: 'strong' };
    }
    return { targetAction: 'exchange', reason: 'hint.exchangeImprove', confidence: 'moderate' };
  }

  if (state.phase === OasisPokerPhase.ACTION) {
    if (state.playerHandRank >= RANK_ONE_PAIR) {
      return { targetAction: 'play', reason: 'hint.pairOrBetter', confidence: 'strong' };
    }
    if (hasAceKing(state.playerHand)) {
      return { targetAction: 'play', reason: 'hint.aceKingHigh', confidence: 'moderate' };
    }
    return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
  }

  return null;
}

function hasAceKing(cards: Card[]): boolean {
  const values = new Set(cards.map((c) => c.value));
  return values.has(ACE_VALUE) && values.has(KING_VALUE);
}
