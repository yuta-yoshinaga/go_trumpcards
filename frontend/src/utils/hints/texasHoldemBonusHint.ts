import type { Card, TexasHoldemBonusResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TexasHoldemBonusPhase } from '../../types/phases';

/** Card values for hint heuristics (sync: internal/domain/Card.go). */
const ACE_VALUE = 1;
const KING_VALUE = 13;
const QUEEN_VALUE = 12;
const JACK_VALUE = 11;
const TEN_VALUE = 10;

/** PokerHandOnePair = 1 (sync: internal/domain/PokerPlayer.go). */
const RANK_ONE_PAIR = 1;

/**
 * Returns a Texas Hold'em Bonus Poker hint for the current decision phase.
 * Uses simplified strategy: pre-flop play with any pocket pair / Ace-anything /
 * suited two broadway cards; otherwise fold. Post-flop raise on pair or better
 * (or strong draws like flush/straight on the board), otherwise check.
 */
export function getTexasHoldemBonusHint(state: TexasHoldemBonusResponse): HintResult | null {
  if (!state.playerHand || state.playerHand.length < 2) return null;

  if (state.phase === TexasHoldemBonusPhase.PRE_FLOP) {
    return preFlopHint(state.playerHand);
  }

  if (state.phase === TexasHoldemBonusPhase.FLOP || state.phase === TexasHoldemBonusPhase.TURN) {
    if (state.playerHandRank >= RANK_ONE_PAIR) {
      return { targetAction: 'raise', reason: 'hint.pairOrBetter', confidence: 'strong' };
    }
    return { targetAction: 'check', reason: 'hint.weakHand', confidence: 'moderate' };
  }

  return null;
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
  if (hasAce) {
    return { targetAction: 'play', reason: 'hint.acePlay', confidence: 'strong' };
  }
  if (suited && bothBroadway(va, vb)) {
    return { targetAction: 'play', reason: 'hint.suitedBroadway', confidence: 'moderate' };
  }
  if (bothBroadway(va, vb)) {
    return { targetAction: 'play', reason: 'hint.broadwayCards', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'hint.weakHand', confidence: 'moderate' };
}

function bothBroadway(a: number, b: number): boolean {
  return isBroadway(a) && isBroadway(b);
}

function isBroadway(v: number): boolean {
  return v === ACE_VALUE || v === KING_VALUE || v === QUEEN_VALUE || v === JACK_VALUE || v === TEN_VALUE;
}
