import type { AcesUpResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Aces Up phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns an Aces Up frontend hint derived from the backend hint, or null.
 *
 * The suggestion is computed by the Go backend and rides along with every
 * state response (see AcesUpWebPresenter.Output, #4483), so this only has to
 * map the recommended action onto the button that performs it.
 */
export function getAcesUpHint(state: AcesUpResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint) return null;

  switch (hint.type) {
    case 'remove':
      return { targetAction: 'remove', reason: 'frontendHint.acesupRemove', confidence: 'strong' };
    case 'move':
      return { targetAction: 'move', reason: 'frontendHint.acesupMove', confidence: 'moderate' };
    case 'draw':
      return { targetAction: 'deal', reason: 'frontendHint.acesupDeal', confidence: 'moderate' };
    default:
      return null;
  }
}
