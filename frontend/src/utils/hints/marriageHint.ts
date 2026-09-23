import type { Card, MarriageResponse } from '../../types/card';
import { MarriagePhase } from '../../types/games/marriage';
import type { HintResult } from '../../types/hint';
import { marriageIsWild } from '../marriageDeclare';

/**
 * Returns a frontend HintResult for Marriage, or null if no suggestion
 * applies (no human seat, empty hand, game over, or not the human's turn).
 */
export function getMarriageHint(state: MarriageResponse): HintResult | null {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx === -1) return null;
  const human = state.players[humanIdx];
  if (human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.currentPlayerIdx !== humanIdx) return null;

  if (state.phase === MarriagePhase.DRAW) {
    return getDrawHint(human.cards, state.discardTop, state.wildRank);
  }
  if (state.phase === MarriagePhase.DISCARD) {
    return getDiscardHint(state.canDeclare);
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

/** Discard phase: declare when the backend confirms one discard makes a valid declaration. */
function getDiscardHint(canDeclare: boolean): HintResult {
  if (canDeclare) {
    return { targetAction: 'declare', reason: 'hint.declareNow', confidence: 'strong' };
  }
  return { targetAction: 'discard', reason: 'hint.discardDeadwood', confidence: 'moderate' };
}

/** A card is wild according to Marriage's wild-rank rules. */
export function isWild(card: Card, wildRank: number): boolean {
  return marriageIsWild(card, wildRank);
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
