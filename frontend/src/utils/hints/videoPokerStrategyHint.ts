import type { VideoPokerResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { VideoPokerPhase } from '../../types/phases';

/** Builds a video poker hint from the strategy advice returned by the server. */
export function getVideoPokerStrategyHint(state: VideoPokerResponse): HintResult | null {
  if (state.phase !== VideoPokerPhase.DRAW || !state.recommendedHoldRule) return null;
  if (state.recommendedHold.length === 0) {
    return { targetAction: 'draw-all', reason: 'hint.strategy.drawAll', confidence: 'moderate' };
  }
  return {
    targetAction: `hold:${state.recommendedHold.join(',')}`,
    reason: `hint.strategy.${state.recommendedHoldRule}`,
    confidence: 'strong',
  };
}
