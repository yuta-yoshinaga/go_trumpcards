import type { Card, IndianRummyResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { IndianRummyPhase } from '../../types/phases';

/**
 * Returns a frontend HintResult for Indian Rummy, or null if no suggestion
 * applies (no human seat, empty hand, game over, or not the human's turn).
 */
export function getIndianRummyHint(state: IndianRummyResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;
  const human = state.players[humanIdx];
  if (human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.currentPlayerIdx !== humanIdx) return null;

  if (state.phase === IndianRummyPhase.DRAW) {
    return getDrawHint(human.cards, state.discardTop, state.wildRank);
  }
  if (state.phase === IndianRummyPhase.DISCARD) {
    return getDiscardHint(state.declarableDiscards);
  }
  return null;
}

/** Draw phase: take the discard top when it fits, otherwise draw from the stock. */
function getDrawHint(hand: Card[], discardTop: Card | null, wildRank: number): HintResult {
  if (discardTop && !isWild(discardTop, wildRank) && fitsWithHand(discardTop, hand, wildRank)) {
    return { targetAction: 'drawDiscard', reason: 'hint.drawFromDiscard', confidence: 'strong' };
  }
  return { targetAction: 'drawStock', reason: 'hint.drawFromStock', confidence: 'moderate' };
}

/**
 * Discard phase: declare for a server-approved discard, else drop deadwood.
 *
 * The server decides: a hand can reach zero deadwood and still be an illegal
 * declaration (two sequences are required, one of them pure), so deadwood is
 * not the question being asked here.
 */
function getDiscardHint(declarableDiscards: number[]): HintResult {
  if (declarableDiscards.length > 0) {
    return { targetAction: 'declare', reason: 'hint.declareNow', confidence: 'strong' };
  }
  return { targetAction: 'discard', reason: 'hint.discardDeadwood', confidence: 'moderate' };
}

/** A card is wild if it is a printed joker or matches the turned-up wild rank. */
export function isWild(card: Card, wildRank: number): boolean {
  return card.design === 'JOKER' || (wildRank > 0 && card.value === wildRank);
}

/** Check if a card fits with the hand to form (or extend) a potential meld. */
function fitsWithHand(card: Card, hand: Card[], wildRank: number): boolean {
  const natural = hand.filter((c) => !isWild(c, wildRank));

  // Set: another card of the same value.
  if (natural.filter((c) => c.value === card.value).length >= 1) return true;

  // Run: same suit within two ranks (adjacent or one-gap).
  for (const c of natural) {
    if (c.design !== card.design) continue;
    const diff = Math.abs(c.value - card.value);
    if (diff === 1 || diff === 2) return true;
  }
  return false;
}
