import type { NarcoticResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Narcotic phase: playing. Later phases are game clear / game over. */
const PHASE_PLAYING = 0;

/**
 * Returns a Narcotic frontend hint derived from the backend hint, or null.
 *
 * The suggestion is computed by the Go backend and rides along with every
 * state response (see NarcoticWebPresenter.Output, #4483), so this only has to
 * map the recommended action onto the button that performs it.
 */
export function getNarcoticHint(state: NarcoticResponse): HintResult | null {
  if (state.phase !== PHASE_PLAYING) return null;
  const hint = state.hint;
  if (!hint) return null;

  switch (hint.type) {
    case 'remove':
      return { targetAction: 'remove', reason: 'frontendHint.narcoticRemove', confidence: 'strong' };
    case 'move':
      return { targetAction: 'move', reason: 'frontendHint.narcoticMove', confidence: 'moderate' };
    case 'draw':
      return { targetAction: 'deal', reason: 'frontendHint.narcoticDeal', confidence: 'moderate' };
    // **クローン元には無い手。**山札が尽きても場を集めれば続けられる。
    case 'redeal':
      return { targetAction: 'redeal', reason: 'frontendHint.narcoticRedeal', confidence: 'moderate' };
    default:
      return null;
  }
}
