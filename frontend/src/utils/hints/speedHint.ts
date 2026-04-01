import type { Card, SpeedResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { SpeedPhase } from '../../types/phases';

/** Maximum card value (King). */
const MAX_VALUE = 13;

/** Returns a frontend HintResult for Speed, or null if no suggestion. */
export function getSpeedHint(state: SpeedResponse): HintResult | null {
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;
  if (state.gameEndFlag) return null;
  if (state.phase !== SpeedPhase.PLAY) return null;

  const hasPlayable = human.cards.some((c) => canPlayOnAnyPile(c, state.centerPiles));

  if (hasPlayable) {
    return { targetAction: 'play', reason: 'hint.hasPlayable', confidence: 'strong' };
  }

  return { targetAction: 'wait', reason: 'hint.noPlayable', confidence: 'moderate' };
}

/** Check if a card can be played on any center pile (value +/- 1, with wrap-around). */
function canPlayOnAnyPile(card: Card, centerPiles: Card[]): boolean {
  return centerPiles.some((pile) => isAdjacent(card.value, pile.value));
}

/** Check if two values are adjacent (with King-Ace wrap-around). */
function isAdjacent(a: number, b: number): boolean {
  const diff = Math.abs(a - b);
  return diff === 1 || diff === MAX_VALUE - 1;
}
