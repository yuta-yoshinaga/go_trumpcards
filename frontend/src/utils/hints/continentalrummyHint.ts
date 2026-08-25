import type { ContinentalRummyResponse } from '../../types/card';
import { CONTINENTAL_RUMMY_PHASE } from '../../types/games/continentalrummy';
import type { HintResult } from '../../types/hint';

/**
 * Returns a frontend {@link HintResult} for Continental Rummy, or null when
 * there is nothing to advise.
 *
 * **The recommendation comes from the server.** Whether the hand can go out is
 * a fifteen-card partition problem the domain already solves, and which card
 * is loose depends on the same sequence logic — re-deriving either here would
 * put the rule in a second place and let the two disagree.
 */
export function getContinentalrummyHint(state: ContinentalRummyResponse): HintResult | null {
  if (state.gameEndFlag || !state.isHumanTurn) return null;
  if (!state.hintReason) return null;

  if (state.phase === CONTINENTAL_RUMMY_PHASE.DRAW) {
    return {
      targetAction: state.hintReason === 'take_discard' ? 'take' : 'stock',
      reason: `frontendHint.continentalrummy_${state.hintReason}`,
      confidence: state.hintReason === 'take_discard' ? 'strong' : 'moderate',
    };
  }
  if (state.phase === CONTINENTAL_RUMMY_PHASE.DISCARD) {
    // **上がれるときは札ではなく「上がる」を指す。**
    if (state.goOutIdx >= 0) {
      return {
        targetAction: 'goout',
        reason: 'frontendHint.continentalrummy_go_out',
        confidence: 'strong',
      };
    }
    if (state.hintDiscardIdx < 0) return null;
    return {
      // **どの札を捨てるかまで言う。** 「緩い札を捨てて」だけだと、
      // どれが緩いのかを player がもう一度目で探すことになる (#4887)。
      targetAction: 'discard',
      targetPos: state.hintDiscardIdx,
      reason: `frontendHint.continentalrummy_${state.hintReason}`,
      confidence: 'moderate',
    };
  }
  return null;
}
