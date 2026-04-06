import type { PineappleResponse } from '../../../types/card';
import { formatHoldemState } from './holdemFormatter';

/** Format a Pineapple Poker game state as terminal text. */
export function formatPineappleState(state: PineappleResponse): string {
  let output = formatHoldemState(state).replace("Texas Hold'em", 'Pineapple');
  if (state.isDiscardPhase) {
    output = output.replace(/phase: .*/, 'phase: DISCARD');
  }
  return output;
}
