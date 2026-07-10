import type { Card, WizardResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { WizardPhase } from '../../types/phases';

/** Maps numeric trump suit to card design string. */
const SUIT_NUM_TO_DESIGN: Readonly<Record<number, Card['design']>> = {
  1: 'SPADE',
  2: 'CLOVER',
  3: 'HEART',
  4: 'DIAMOND',
};

/** High card threshold for bid estimation. */
const HIGH_CARD_VALUE = 11;
/** Minimum high cards for strong bid confidence. */
const STRONG_BID_THRESHOLD = 3;

/** True when a card is a Wizard (always wins a trick). */
function isWizard(card: Card): boolean {
  return card.deck === 'wizard' && card.label === 'Wizard';
}

/** True when a card is a Jester (always loses a trick). */
function isJester(card: Card): boolean {
  return card.deck === 'wizard' && card.label === 'Jester';
}

/** Returns a frontend HintResult for Wizard, or null if no suggestion. */
export function getWizardHint(state: WizardResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === WizardPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards);
  }

  if (state.phase === WizardPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate bid from Wizards, high cards, and trump count. */
function getBidHint(cards: Card[]): HintResult {
  // Wizards always take a trick, so each one is a near-certain point.
  const wizardCount = cards.filter(isWizard).length;
  const highCards = cards.filter((c) => !isWizard(c) && !isJester(c) && c.value >= HIGH_CARD_VALUE).length;
  const estimatedTricks = Math.max(0, wizardCount + Math.round(highCards * 0.5));

  const confidence = wizardCount + highCards >= STRONG_BID_THRESHOLD ? 'strong' : 'moderate';
  return { targetAction: `bid:${estimatedTricks}`, reason: 'hint.bidEstimate', confidence };
}

/** Hint for play phase: play a Wizard to win, dump a Jester, follow suit, trump, or discard. */
function getPlayHint(cards: Card[], state: WizardResponse): HintResult {
  const trumpDesign = SUIT_NUM_TO_DESIGN[state.trumpSuit];
  const trick = state.currentTrick;

  // A Wizard guarantees the trick (unless already claimed by an earlier Wizard).
  const trickHasWizard = trick.some((tc) => isWizard(tc.card));
  if (cards.some(isWizard) && !trickHasWizard) {
    return { targetAction: 'play', reason: 'hint.playWizard', confidence: 'strong' };
  }

  // Leading the trick
  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadStrategic', confidence: 'moderate' };
  }

  // The led suit is set by the first non-Jester card; a leading Jester leaves the trick open.
  const leadCard = trick.find((tc) => !isJester(tc.card) && !isWizard(tc.card))?.card;
  const ledSuit = leadCard?.design;
  const suitCards = ledSuit ? cards.filter((c) => !isWizard(c) && !isJester(c) && c.design === ledSuit) : [];

  if (suitCards.length > 0) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }

  // Void in led suit: trump or discard a Jester / low card.
  if (cards.some(isJester)) {
    return { targetAction: 'play', reason: 'hint.dumpJester', confidence: 'moderate' };
  }

  const hasTrump = cards.some((c) => !isWizard(c) && !isJester(c) && c.design === trumpDesign);
  if (hasTrump) {
    return { targetAction: 'play', reason: 'hint.trumpWithCard', confidence: 'moderate' };
  }

  return { targetAction: 'play', reason: 'hint.discardLowest', confidence: 'moderate' };
}
