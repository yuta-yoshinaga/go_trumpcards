import type { OmahaResponse } from '../../../types/card';
import { formatHoldemState } from './holdemFormatter';

/** Format an Omaha Hold'em game state as terminal text. */
export function formatOmahaState(state: OmahaResponse): string {
  return formatHoldemState(state).replace("Texas Hold'em", 'Omaha');
}
