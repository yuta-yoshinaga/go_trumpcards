import type { Card, SevensResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Map Card design to suit index used by the Sevens table arrays. */
const SUIT_INDEX: Partial<Record<Card['design'], number>> = {
  SPADE: 1,
  CLOVER: 2,
  HEART: 3,
  DIAMOND: 4,
};

/** Returns a frontend HintResult for Sevens, or null if no suggestion. */
export function getSevensHint(state: SevensResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.isFinished) return null;
  if (state.gameEndFlag) return null;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  if (state.currentTurn !== humanIdx) return null;

  const playable = human.cards.some((c) => isPlayable(c, state));

  if (playable) {
    return { targetAction: 'play', reason: 'hint.playExtend', confidence: 'strong' };
  }

  // No playable cards — suggest pass
  if (human.maxPasses > 0 && human.maxPasses - human.passesUsed <= 1) {
    return { targetAction: 'pass', reason: 'hint.passLimitWarning', confidence: 'strong' };
  }

  return { targetAction: 'pass', reason: 'hint.passAvailable', confidence: 'moderate' };
}

/** Check if a card can be placed on the table (adjacent to current range, with tunnel wrap). */
function isPlayable(card: Card, state: SevensResponse): boolean {
  const suitIdx = SUIT_INDEX[card.design];
  if (suitIdx === undefined) return false;

  const min = state.tableMinVals[suitIdx];
  const max = state.tableMaxVals[suitIdx];

  if (card.value === min - 1 || card.value === max + 1) return true;

  // Tunnel rule: A and K wrap around
  if (state.config.tunnelEnabled) {
    if (min === 1 && card.value === 13) return true;
    if (max === 13 && card.value === 1) return true;
  }

  return false;
}
