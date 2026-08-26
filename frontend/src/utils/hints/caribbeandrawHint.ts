import type { Card, CaribbeanDrawResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CaribbeanDrawPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Card value representing an Ace. */
const ACE_VALUE = 1;

/** Card value representing a King. */
const KING_VALUE = 13;

/**
 * Returns a Caribbean Draw Poker hint for the draw and action phases.
 *
 * Draw phase: stand pat once a made hand (one pair or better) is already there,
 * otherwise it is worth paying the fee to go for one. Action phase: the
 * simplified Caribbean basic strategy — call with a pair or better, or with
 * Ace-King high; fold otherwise.
 *
 * Kept deliberately in step with `CaribbeanDrawCuiPresenter.HintOutput`
 * (internal/adapter/presenter/CaribbeanDrawCuiPresenter.go): if the two drift,
 * the same hand gets opposite advice depending on which UI the player opened.
 */
export function getCaribbeanDrawHint(state: CaribbeanDrawResponse): HintResult | null {
  if (!state.playerHand || state.playerHand.length === 0) return null;

  if (state.phase === CaribbeanDrawPhase.DRAW) {
    if (state.playerHandRank >= RANK_ONE_PAIR) {
      return { targetAction: 'stand', reason: 'hint.standPat', confidence: 'strong' };
    }
    return { targetAction: 'draw', reason: 'hint.drawWeak', confidence: 'moderate' };
  }

  if (state.phase !== CaribbeanDrawPhase.ACTION) return null;

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
