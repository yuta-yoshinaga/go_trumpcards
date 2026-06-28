import type { AllFoursResponse, Card, CardDesign } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { AllFoursPhase } from '../../types/phases';

/** Trump-strength threshold below which begging / running is recommended. */
const WEAK_TRUMP_STRENGTH = 2;

/** Maps a card design string to its numeric suit id (matches the Go enum). */
const DESIGN_TO_SUIT: Readonly<Record<CardDesign, number>> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
  JOKER: 0,
};

/** True when the card's suit matches the numeric trump suit. */
function isTrump(card: Card, trumpSuit: number): boolean {
  return DESIGN_TO_SUIT[card.design] === trumpSuit;
}

/** Returns a frontend HintResult for All Fours, or null if no suggestion. */
export function getAllFoursHint(state: AllFoursResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === AllFoursPhase.BEG) {
    if (state.nonDealerIdx !== humanIdx) return null;
    return getBegHint(human.cards, state.trumpSuit);
  }

  if (state.phase === AllFoursPhase.GIFT) {
    if (state.dealerIdx !== humanIdx) return null;
    return getGiftHint(human.cards, state.trumpSuit);
  }

  if (state.phase === AllFoursPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx || human.cards.length === 0) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate the strength of the player's trump holding. */
function trumpStrength(cards: Card[], trumpSuit: number): number {
  let strength = 0;
  for (const c of cards) {
    if (!isTrump(c, trumpSuit)) continue;
    strength += 1;
    // Queen (12), King (13), Ace (1) count as a strong honour.
    if (c.value === 12 || c.value === 13 || c.value === 1) strength += 1;
  }
  return strength;
}

/** Hint for the beg/stand decision (non-dealer). */
function getBegHint(cards: Card[], trumpSuit: number): HintResult {
  if (trumpStrength(cards, trumpSuit) < WEAK_TRUMP_STRENGTH) {
    return { targetAction: 'beg:true', reason: 'hint.begBeg', confidence: 'moderate' };
  }
  return { targetAction: 'beg:false', reason: 'hint.begStand', confidence: 'moderate' };
}

/** Hint for the gift/run decision (dealer). */
function getGiftHint(cards: Card[], trumpSuit: number): HintResult {
  if (trumpStrength(cards, trumpSuit) < WEAK_TRUMP_STRENGTH) {
    return { targetAction: 'respond:true', reason: 'hint.giftRun', confidence: 'moderate' };
  }
  return { targetAction: 'respond:false', reason: 'hint.giftGift', confidence: 'moderate' };
}

/** Hint for the play phase. */
function getPlayHint(cards: Card[], state: AllFoursResponse): HintResult {
  if (state.currentTrick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadStrong', confidence: 'moderate' };
  }
  const leadSuit = state.currentTrick[0].card.design;
  const hasLead = cards.some((c) => c.design === leadSuit);
  if (hasLead) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }
  const hasTrump = cards.some((c) => isTrump(c, state.trumpSuit));
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpCut', confidence: 'moderate' };
  }
  return { targetAction: 'play', reason: 'hint.discardLow', confidence: 'moderate' };
}
