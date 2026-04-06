import type { ShortDeckResponse } from '../../../types/card';
import { formatHoldemState } from './holdemFormatter';

/** Format a Short Deck Hold'em game state as terminal text. */
export function formatShortdeckState(state: ShortDeckResponse): string {
  return formatHoldemState(state).replace("Texas Hold'em", 'Short Deck');
}
