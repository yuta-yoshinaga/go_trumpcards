import type { ScopaResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { scopaCaptureValue, scopaTakeCandidates } from '../scopaTakeCandidates';

/**
 * Scopa hint heuristic. Recommends the action with the highest expected point
 * gain for the current human hand.
 *
 * Priority order:
 *   1. Capture that grabs the settebello (7♦) — worth a dedicated scoring point.
 *   2. Capture that grabs at least one diamond (the diamonds / primiera race).
 *   3. Any capture (a card in hand can take table cards).
 *   4. Lay (no capture is possible).
 * Returns null when there is no state, the game has ended, or the hand is empty.
 */
export function getScopaHint(state: ScopaResponse): HintResult | null {
  if (!state || state.gameEndFlag) return null;
  const human = state.players.find((p) => p.isHuman);
  if (!human || human.cards.length === 0) return null;

  let canCapture = false;
  let capturesSetteBello = false;
  let maxDiamonds = 0;

  for (const c of human.cards) {
    const target = scopaCaptureValue(c.value);
    const { indices } = scopaTakeCandidates(state.tableCards, target);
    if (indices.size === 0) continue;
    canCapture = true;

    let diamonds = 0;
    for (const i of indices) {
      const cardOnTable = state.tableCards[i];
      if (!cardOnTable) continue;
      if (cardOnTable.design === 'DIAMOND') {
        diamonds += 1;
        if (cardOnTable.value === 7) capturesSetteBello = true;
      }
    }
    if (diamonds > maxDiamonds) maxDiamonds = diamonds;
  }

  if (capturesSetteBello) {
    return { targetAction: 'take', reason: 'hint.take.settebello', confidence: 'strong' };
  }
  if (maxDiamonds > 0) {
    return { targetAction: 'take', reason: 'hint.take.diamonds', confidence: 'strong' };
  }
  if (canCapture) {
    return { targetAction: 'take', reason: 'hint.take.capture', confidence: 'moderate' };
  }
  return { targetAction: 'lay', reason: 'hint.lay.safe', confidence: 'moderate' };
}
