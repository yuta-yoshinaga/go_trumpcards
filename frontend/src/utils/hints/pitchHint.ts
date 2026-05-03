import type { Card, PitchResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { PitchPhase } from '../../types/phases';

const HIGH_CARD_VALUE = 11;

/** Returns a frontend HintResult for Pitch, or null if no suggestion. */
export function getPitchHint(state: PitchResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;
  const humanIdx = state.players.findIndex((p) => p.isHuman);

  if (state.phase === PitchPhase.BID) {
    if (state.bidPlayerIdx !== humanIdx) return null;
    return getBidHint(human.cards, state.currentBid);
  }

  if (state.phase === PitchPhase.PLAY) {
    if (state.currentPlayerIdx !== humanIdx) return null;
    return getPlayHint(human.cards, state);
  }

  return null;
}

/** Estimate suggested bid from high-card count. */
function getBidHint(cards: Card[], currentBid: number): HintResult {
  const highCards = cards.filter((c) => c.value === 1 || c.value >= HIGH_CARD_VALUE).length;
  const estimated = Math.min(4, Math.max(0, Math.floor(highCards / 1.5)));
  if (estimated < 2 || estimated <= currentBid) {
    return { targetAction: 'bid:0', reason: 'hint.bidPass', confidence: 'moderate' };
  }
  return { targetAction: `bid:${estimated}`, reason: 'hint.bidStrong', confidence: 'strong' };
}

/** Hint for play phase. */
function getPlayHint(cards: Card[], state: PitchResponse): HintResult {
  if (state.currentTrick.length === 0) {
    if (state.trumpSuit === 0) {
      return { targetAction: 'play', reason: 'hint.setTrumpLead', confidence: 'strong' };
    }
    return { targetAction: 'play', reason: 'hint.leadStrong', confidence: 'moderate' };
  }
  const leadSuit = state.currentTrick[0].card.design;
  const hasLead = cards.some((c) => c.design === leadSuit);
  if (hasLead) {
    return { targetAction: 'play', reason: 'hint.followSuit', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'hint.discardLow', confidence: 'moderate' };
}
