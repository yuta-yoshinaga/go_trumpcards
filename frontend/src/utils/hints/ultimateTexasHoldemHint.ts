import type { Card, UltimateTexasHoldemResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { UltimateTexasHoldemPhase } from '../../types/phases';

/** Card values for hint heuristics (sync: internal/domain/Card.go). */
const ACE_VALUE = 1;
const KING_VALUE = 13;
const QUEEN_VALUE = 12;
const JACK_VALUE = 11;
const TEN_VALUE = 10;

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/**
 * Returns an Ultimate Texas Hold'em hint for the current decision phase.
 *
 * Strategy summary:
 *   - Pre-flop: bet 4× with strong starters (pocket pair, suited Ace, Broadway pair)
 *     and 3× with weaker premiums; otherwise Check.
 *   - Flop: bet 2× when player already has at least a pair; otherwise Check.
 *   - River: Play 1× when player has any made hand; otherwise Fold.
 */
export function getUltimateTexasHoldemHint(state: UltimateTexasHoldemResponse): HintResult | null {
  if (!state.playerHand || state.playerHand.length < 2) return null;

  switch (state.phase) {
    case UltimateTexasHoldemPhase.PRE_FLOP:
      return preFlopHint(state.playerHand);
    case UltimateTexasHoldemPhase.FLOP:
      if (state.playerHandRank >= RANK_ONE_PAIR) {
        return { targetAction: 'raise', reason: 'hint.pairOrBetter', confidence: 'strong' };
      }
      return { targetAction: 'check', reason: 'hint.weakHand', confidence: 'moderate' };
    case UltimateTexasHoldemPhase.RIVER:
      if (state.playerHandRank >= RANK_ONE_PAIR) {
        return { targetAction: 'play', reason: 'hint.madeHandRiver', confidence: 'strong' };
      }
      return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
    default:
      return null;
  }
}

function preFlopHint(cards: Card[]): HintResult {
  const [a, b] = cards;
  const va = a.value;
  const vb = b.value;
  const suited = a.design === b.design;
  const hasAce = va === ACE_VALUE || vb === ACE_VALUE;
  const isPair = va === vb;

  if (isPair) {
    return { targetAction: 'play', reason: 'hint.pocketPair', confidence: 'strong' };
  }
  if (suited && hasAce) {
    return { targetAction: 'play', reason: 'hint.suitedBroadway', confidence: 'strong' };
  }
  if (hasAce) {
    return { targetAction: 'play', reason: 'hint.acePlay', confidence: 'strong' };
  }
  if (suited && bothBroadway(va, vb)) {
    return { targetAction: 'play', reason: 'hint.suitedBroadway', confidence: 'moderate' };
  }
  if (bothBroadway(va, vb)) {
    return { targetAction: 'play', reason: 'hint.broadwayCards', confidence: 'moderate' };
  }
  return { targetAction: 'check', reason: 'hint.weakHand', confidence: 'moderate' };
}

function bothBroadway(a: number, b: number): boolean {
  return isBroadway(a) && isBroadway(b);
}

function isBroadway(v: number): boolean {
  return v === ACE_VALUE || v === KING_VALUE || v === QUEEN_VALUE || v === JACK_VALUE || v === TEN_VALUE;
}
