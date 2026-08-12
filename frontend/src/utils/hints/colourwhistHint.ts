import type { ColourWhistResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { ColourWhistPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Colour Whist, or null when there is
 * nothing to advise.
 *
 * There is nothing to say about a Troel deal — the contract was settled before
 * anyone had a choice — so the hint only speaks during a real auction, and in
 * play it names the follow-suit rule rather than pretending to a read.
 */
export function getColourwhistHint(state: ColourWhistResponse): HintResult | null {
  if (state.gameEndFlag || !state.isHumanTurn) return null;

  if (state.phase === ColourWhistPhase.BID) {
    return { targetAction: 'bid', reason: 'frontendHint.colourWhistBidStrength', confidence: 'moderate' };
  }
  if (state.phase === ColourWhistPhase.PLAY) {
    return { targetAction: 'play', reason: 'frontendHint.colourWhistFollowSuit', confidence: 'moderate' };
  }
  return null;
}
