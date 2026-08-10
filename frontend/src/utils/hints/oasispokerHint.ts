import type { Card, OasisPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { OasisPokerPhase } from '../../types/phases';

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/** Card value representing an Ace. */
const ACE_VALUE = 1;

/** Card value representing a King. */
const KING_VALUE = 13;

/** Lowest value counted as a high card to hold (Jack). */
const JACK_VALUE = 11;

/**
 * Hand indices worth swapping out: cards that are neither part of a pair nor a
 * high card (J/Q/K/A).
 *
 * Sync: `oasisPokerExchangeIndices` in `OasisPokerCuiPresenter.go`, which the
 * CUI already prints as `[1]♠5 [3]♣2`. The Web hint only said "exchange" and
 * left the player to work out which of the five to click (#4711).
 *
 * **An Ace is `value === 1`,** so a plain "is it big?" comparison throws away
 * the best card in the hand.
 */
function exchangeIndices(hand: readonly Card[]): number[] {
  const rankCount = new Map<number, number>();
  for (const c of hand) {
    rankCount.set(c.value, (rankCount.get(c.value) ?? 0) + 1);
  }
  const out: number[] = [];
  hand.forEach((c, i) => {
    const isPair = (rankCount.get(c.value) ?? 0) >= 2;
    const isHigh = c.value === ACE_VALUE || c.value >= JACK_VALUE;
    if (!isPair && !isHigh) out.push(i);
  });
  return out;
}

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
    const toSwap = exchangeIndices(state.playerHand);
    // **交換するものが無いときだけ stand。**低いペアがあるからと「そのまま」
    // と言うと、残り3枚を引き直す機会を捨てさせる。CUI は ex が空のときだけ
    // hintStand を出している。
    if (toSwap.length === 0) {
      return { targetAction: 'stand', reason: 'hint.exchangeKeep', confidence: 'strong' };
    }
    return {
      targetAction: 'exchange',
      reason: 'hint.exchangeImprove',
      confidence: state.playerHandRank >= RANK_ONE_PAIR ? 'strong' : 'moderate',
      targetIndices: toSwap,
    };
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
