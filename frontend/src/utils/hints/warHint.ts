import type { WarResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';

/** Returns a War frontend hint or null. */
export function getWarHint(state: WarResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  const hasRevealed = state.playerRevealed !== null && state.cpuRevealed !== null;
  return {
    targetAction: 'step',
    reason: hasRevealed ? 'hint.resolveRound' : 'hint.flipCard',
    confidence: 'strong',
  };
}
