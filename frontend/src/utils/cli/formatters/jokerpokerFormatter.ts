import type { VideoPokerResponse } from '../../../types/card';
import { formatVideopokerState } from './videopokerFormatter';

/** Format a Joker Poker game state as terminal text. */
export function formatJokerpokerState(state: VideoPokerResponse): string {
  return formatVideopokerState(state);
}
