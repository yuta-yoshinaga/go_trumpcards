import type { VideoPokerResponse } from '../../../types/card';
import { formatVideopokerState } from './videopokerFormatter';

/** Format a Deuces Wild game state as terminal text. */
export function formatDeuceswildState(state: VideoPokerResponse): string {
  return formatVideopokerState(state);
}
