import type { IronCrossResponse } from '../../types/card';
import type { HintResult } from '../../types/hint';
import { IronCrossPhase } from '../../types/phases';

/** Rank thresholds, matching the Go domain's advice. */
const TWO_PAIR = 3;
const TRIPS = 4;
const ONE_PAIR = 2;

/** Line values, matching the Go domain. */
const VERTICAL = 1;
const HORIZONTAL = 2;

/**
 * Returns a frontend {@link HintResult} for Iron Cross, or null when there is
 * nothing to advise.
 *
 * **The one advice that matters is which arm to take**, and it is the only
 * irreversible decision in the game. The server publishes the recommendation
 * through the hint endpoint; this frontend hint covers the betting rounds and
 * points at the choose buttons once the cross is complete.
 *
 * The page cannot see the seat's rank before showdown (the server withholds
 * it), so the betting advice leans on the pot price the server does publish.
 */
export function getIroncrossHint(state: IronCrossResponse): HintResult | null {
  if (state.gameEndFlag) return null;

  // **選ぶ場面は最優先。** ここだけは取り返しがつかない。
  if (state.isChoosing || state.phase === IronCrossPhase.CHOOSE_LINE) {
    const seat = state.seats[state.humanSeat];
    if (seat && seat.line !== 0) return null;
    return { targetAction: 'line', reason: 'frontendHint.ironcrossPickTheStrongerLine', confidence: 'strong' };
  }

  if (state.phase !== IronCrossPhase.BETTING || !state.isHumanTurn) return null;

  const seat = state.seats[state.humanSeat];
  if (!seat) return null;

  // handRank is only populated at showdown, so treat 0 as "not known yet".
  const rank = seat.handRank;

  if (state.toCall <= 0) {
    if (rank >= TWO_PAIR && state.canRaise) {
      return { targetAction: 'bet', reason: 'frontendHint.ironcrossStrongEnoughToBet', confidence: 'moderate' };
    }
    // **中央が最後に開く。** 上下左右が出そろっても勝負は決まっていない。
    return { targetAction: 'check', reason: 'frontendHint.ironcrossSeeAnotherCard', confidence: 'strong' };
  }
  if (rank >= TRIPS && state.canRaise) {
    return { targetAction: 'raise', reason: 'frontendHint.ironcrossStrongEnoughToRaise', confidence: 'moderate' };
  }
  if (rank >= ONE_PAIR) {
    return { targetAction: 'call', reason: 'frontendHint.ironcrossWorthACall', confidence: 'moderate' };
  }
  const ante = state.config?.ante ?? 0;
  if (state.toCall <= ante) {
    return { targetAction: 'call', reason: 'frontendHint.ironcrossCheapToStay', confidence: 'moderate' };
  }
  return { targetAction: 'fold', reason: 'frontendHint.ironcrossNotWorthIt', confidence: 'moderate' };
}

/** Names the arm the server recommends, for display next to the hint. */
export function ironcrossLineName(line: number): 'vertical' | 'horizontal' | null {
  if (line === VERTICAL) return 'vertical';
  if (line === HORIZONTAL) return 'horizontal';
  return null;
}
