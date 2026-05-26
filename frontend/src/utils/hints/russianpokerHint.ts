import type { Card, RussianPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { RussianPokerPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Card value representing an Ace. */
const ACE_VALUE = 1;

/** Card value representing a King. */
const KING_VALUE = 13;

/**
 * Returns a Russian Poker hint for the action and post-action phases.
 * Basic strategy: call with pair or better, or with Ace-King high; fold otherwise.
 */
export function getRussianPokerHint(state: RussianPokerResponse): HintResult | null {
  if (state.phase !== RussianPokerPhase.ACTION && state.phase !== RussianPokerPhase.POST_ACTION) return null;
  if (!state.playerHand || state.playerHand.length === 0) return null;

  if (state.playerHandRank >= RANK_ONE_PAIR) {
    return { targetAction: 'play', reason: 'hint.pairOrBetter', confidence: 'strong' };
  }

  if (hasAceKing(state.playerHand)) {
    return { targetAction: 'play', reason: 'hint.aceKingHigh', confidence: 'moderate' };
  }

  return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
}

function hasAceKing(cards: Card[]): boolean {
  const values = new Set(cards.map((c) => c.value));
  return values.has(ACE_VALUE) && values.has(KING_VALUE);
}
