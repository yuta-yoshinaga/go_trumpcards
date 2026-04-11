import type { Card, CaribbeanStudResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CaribbeanStudPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Card value representing an Ace. */
const ACE_VALUE = 1;

/** Card value representing a King. */
const KING_VALUE = 13;

/**
 * Returns a Caribbean Stud Poker hint for the action (call/fold) phase.
 * Uses the simplified Caribbean Stud basic strategy: call with pair or better,
 * or with Ace-King high; fold otherwise.
 */
export function getCaribbeanStudHint(state: CaribbeanStudResponse): HintResult | null {
  if (state.phase !== CaribbeanStudPhase.ACTION) return null;
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
