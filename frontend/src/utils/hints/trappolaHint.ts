import type { Card, TrappolaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TrappolaPhase } from '../../types/phases';

/**
 * Trappola trick strength (higher beats lower within the led suit).
 * Order: 3 > 2 > A > K > Q > J > 7 > 6 > 5 > 4.
 */
function strength(value: number): number {
  switch (value) {
    case 3:
      return 9;
    case 2:
      return 8;
    case 1:
      return 7;
    case 13:
      return 6;
    case 12:
      return 5;
    case 11:
      return 4;
    case 7:
      return 3;
    case 6:
      return 2;
    case 5:
      return 1;
    default:
      return 0;
  }
}

/** Returns a frontend HintResult for Trappola, or null if no suggestion. */
export function getTrappolaHint(state: TrappolaResponse): HintResult | null {
  if (state.phase !== TrappolaPhase.PLAY) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (humanIdx < 0 || state.currentPlayerIdx !== humanIdx) return null;
  const human = state.players[humanIdx];
  if (!human || human.cards.length === 0) return null;

  const trick = state.currentTrick;
  if (trick.length === 0) {
    return { targetAction: 'play', reason: 'hint.leadLow', confidence: 'moderate' };
  }

  const ledSuit = trick[0].card.design;
  const suitCards = human.cards.filter((c: Card) => c.design === ledSuit);
  if (suitCards.length === 0) {
    return { targetAction: 'play', reason: 'hint.discardLow', confidence: 'moderate' };
  }

  // Current winning card of the trick (no trump, so it is the highest led suit).
  let topStrength = -1;
  let topPlayerIdx = trick[0].playerIdx;
  for (const tc of trick) {
    if (tc.card.design === ledSuit && strength(tc.card.value) > topStrength) {
      topStrength = strength(tc.card.value);
      topPlayerIdx = tc.playerIdx;
    }
  }
  const partnerWinning = topPlayerIdx % 2 === humanIdx % 2;
  if (partnerWinning) {
    return { targetAction: 'play', reason: 'hint.givePartner', confidence: 'moderate' };
  }
  const canWin = suitCards.some((c: Card) => strength(c.value) > topStrength);
  if (canWin) {
    return { targetAction: 'play', reason: 'hint.followWin', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'hint.followDuck', confidence: 'moderate' };
}
