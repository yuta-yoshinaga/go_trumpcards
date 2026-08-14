import type { TrogguResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { TrogguPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Troggu, or null when there is
 * nothing to advise.
 *
 * **The contract decides which way the advice points.** Under Piccolo and
 * Misère the declarer must *not* take tricks, so telling them to win one would
 * be advice for losing the deal. Which card comes from the server; what is
 * added here is the shape of the contract.
 */
export function getTrogguHint(state: TrogguResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === TrogguPhase.TRICK_END) {
    return { targetAction: 'next', reason: 'frontendHint.trogguNextTrick', confidence: 'strong' };
  }
  if (state.phase === TrogguPhase.ROUND_END) {
    return { targetAction: 'nextround', reason: 'frontendHint.trogguNextDeal', confidence: 'strong' };
  }
  if (!state.isHumanTurn) return null;

  if (state.phase === TrogguPhase.BID) {
    return { targetAction: 'bid', reason: 'frontendHint.trogguBidPickTheContract', confidence: 'moderate' };
  }

  const human = state.players.find((p) => p.isHuman);
  const avoids = human?.isDeclarer && (state.contractName === 'misere' || state.contractName === 'piccolo');
  if (avoids) {
    return { targetAction: 'play', reason: 'frontendHint.trogguAvoidTricks', confidence: 'strong' };
  }
  return { targetAction: 'play', reason: 'frontendHint.trogguFollowSuit', confidence: 'moderate' };
}
