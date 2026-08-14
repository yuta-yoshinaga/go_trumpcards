import type { HorseResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { HorsePhase } from '../../types/phases';

/**
 * Returns a frontend {@link HintResult} for H.O.R.S.E., or null when there is
 * nothing to advise.
 *
 * **The advice stops at the discipline.** Which hand is worth playing belongs
 * to whichever game is running — repeating it here would mean carrying five
 * sets of poker strategy in one page. What this adds is the thing unique to a
 * mixed game: the rules just changed under you.
 */
export function getHorseHint(state: HorseResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  if (state.phase === HorsePhase.HAND_END) {
    return { targetAction: 'next', reason: 'frontendHint.horseNextHand', confidence: 'strong' };
  }
  if (!state.isHumanTurn) return null;

  // **種目が変わった直後は規則が変わっている。** 最初のハンドだけ強く告げる。
  if (state.handInDiscipline === 1) {
    return { targetAction: 'action', reason: 'frontendHint.horseDisciplineChanged', confidence: 'strong' };
  }
  if (state.communityCards.length === 0 && state.disciplineName.startsWith('omaha')) {
    return { targetAction: 'action', reason: 'frontendHint.horseOmahaUsesTwo', confidence: 'moderate' };
  }
  if (state.disciplineName === 'razz') {
    return { targetAction: 'action', reason: 'frontendHint.horseRazzWantsLow', confidence: 'moderate' };
  }
  return { targetAction: 'action', reason: 'frontendHint.horsePlayTheDiscipline', confidence: 'moderate' };
}
