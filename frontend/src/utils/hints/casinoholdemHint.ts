import type { Card, CasinoHoldemResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { CasinoHoldemPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Card value representing an Ace. */
const ACE_VALUE = 1;

/** Card value representing a King. */
const KING_VALUE = 13;

/**
 * Whether any of the five cards is an Ace or a King.
 *
 * Sync: `CasinoHoldem.RecommendCall` scans hole **and** community and calls on
 * *either* rank — unlike `oasispokerHint`'s `hasAceKing`, which needs both and
 * looks only at the hand. Narrowing this puts the CUI and the Web back out of
 * step, which is the bug this replaced (#4712).
 */
function hasAceOrKing(cards: readonly Card[]): boolean {
  return cards.some((c) => c.value === ACE_VALUE || c.value === KING_VALUE);
}

/**
 * Returns a Casino Hold'em hint for the FLOP (call/fold) decision.
 *
 * Casino Hold'em basic strategy: with ~82% positive EV on calls, the simplest
 * useful rule is "call any pair or better", and "call any flush draw" or
 * "call any open-ended straight draw". This implementation simplifies further
 * to: any pair-or-better, or any Ace or King among the five cards → call;
 * otherwise a moderate fold suggestion. The backend hand rank already reflects
 * the 5-card eval of (hole + flop) (`updatePlayerCurrentRank`), so we lean on it
 * directly instead of re-evaluating cards.
 *
 * **This must stay in step with `CasinoHoldem.RecommendCall`,** which the CUI
 * calls directly. The Ace/King half used to be missing here, so an A-K-high
 * hand got "call" in the CUI and "fold" on the Web (#4712).
 */
export function getCasinoHoldemHint(state: CasinoHoldemResponse): HintResult | null {
  if (state.phase !== CasinoHoldemPhase.FLOP) return null;
  if (!state.playerHand || state.playerHand.length === 0) return null;

  if (state.playerHandRank >= RANK_ONE_PAIR) {
    return { targetAction: 'call', reason: 'hint.pairOrBetter', confidence: 'strong' };
  }

  if (hasAceOrKing([...state.playerHand, ...(state.community ?? [])])) {
    return { targetAction: 'call', reason: 'hint.aceKingHigh', confidence: 'moderate' };
  }

  return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
}
