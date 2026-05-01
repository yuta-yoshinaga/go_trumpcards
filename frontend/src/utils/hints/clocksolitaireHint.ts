import type { ClockSolitaireResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ClockSolitairePhase } from '../../types/phases';

/** Returns a Clock Solitaire frontend hint or null.
 *
 * Clock Solitaire is fully deterministic — the next move is always to flip the
 * card at the current pile. The hint just nudges the player to keep stepping
 * (or to autoplay) while the game is active. */
export function getClocksolitaireHint(state: ClockSolitaireResponse): HintResult | null {
  if (state.phase !== ClockSolitairePhase.PLAYING) return null;
  if (!state.currentCard) {
    return { targetAction: 'step', reason: 'hint.firstStep', confidence: 'strong' };
  }
  return { targetAction: 'step', reason: 'hint.continueStep', confidence: 'strong' };
}
