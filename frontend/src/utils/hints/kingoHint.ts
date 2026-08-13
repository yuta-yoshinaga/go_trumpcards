import type { KingoResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { KingoPhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for Kingo, or null when there is
 * nothing to advise.
 *
 * **There is nothing to read.** The bet is placed before anything is dealt, so
 * no hint here can speak to a hand — the only honest advice is about stake size
 * relative to the stack. A hint that implied otherwise would be pretending to
 * see cards that do not exist yet.
 */
export function getKingoHint(state: KingoResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === KingoPhase.RESULT) {
    return { targetAction: 'next', reason: 'frontendHint.kingoRoundIsOver', confidence: 'strong' };
  }
  if (state.phase !== KingoPhase.BET) return null;

  if (state.isHumanBanker) {
    return { targetAction: 'deal', reason: 'frontendHint.kingoYouAreTheBanker', confidence: 'strong' };
  }

  const seat = state.seats[state.humanSeat];
  if (!seat) return null;

  const minBet = state.config?.minBet ?? 0;
  if (minBet > 0 && seat.chips < minBet * 5) {
    return { targetAction: 'bet', reason: 'frontendHint.kingoStackIsShort', confidence: 'moderate' };
  }
  return { targetAction: 'bet', reason: 'frontendHint.kingoNoInformationYet', confidence: 'moderate' };
}
