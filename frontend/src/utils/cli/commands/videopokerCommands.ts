import type { videopokerApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseVideoPokerCommand, VIDEO_POKER_HELP } from './sharedVideoPokerCommands';

type VideoPokerArgs = Parameters<typeof videopokerApi.exec>;

/** Parse a Video Poker CLI command into API exec arguments. */
export function parseVideopokerCommand(input: string): CliParseResult<VideoPokerArgs> {
  const result = parseVideoPokerCommand(input);
  if ('error' in result) return { error: result.error };
  return { args: [result.command as VideoPokerArgs[0], result.amount, result.indices] };
}

/** Help text for Video Poker CLI mode. */
export const VIDEOPOKER_HELP: string[] = [...VIDEO_POKER_HELP, 'log         - Show action log'];
