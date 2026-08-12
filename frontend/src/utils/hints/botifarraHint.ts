import type { BotifarraResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { BotifarraPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Botifarra, or null when there is
 * nothing to advise.
 *
 * Two decisions are worth advising on. In the declaration phase the usual
 * choice is your longest suit. In play there is often no choice at all — you
 * must win the trick if you can — so the hint names the rule rather than
 * pretending to a read.
 */
export function getBotifarraHint(state: BotifarraResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === BotifarraPhase.DECLARE || state.phase === BotifarraPhase.DELEGATED) {
    return { targetAction: 'declare', reason: 'frontendHint.botifarraDeclareLongest', confidence: 'moderate' };
  }
  if (state.phase === BotifarraPhase.PLAY && state.isHumanTurn) {
    return { targetAction: 'play', reason: 'frontendHint.botifarraMustWin', confidence: 'moderate' };
  }
  return null;
}
