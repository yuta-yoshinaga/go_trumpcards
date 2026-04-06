import type { jokerpokerApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseVideoPokerCommand, VIDEO_POKER_HELP } from './sharedVideoPokerCommands';

type JokerpokerArgs = Parameters<typeof jokerpokerApi.exec>;

/** Parse a Joker Poker CLI command into API exec arguments. */
export function parseJokerpokerCommand(input: string): CliParseResult<JokerpokerArgs> {
  const result = parseVideoPokerCommand(input);
  if ('error' in result) return { error: result.error };
  return { args: [result.command as JokerpokerArgs[0], result.amount, result.indices] };
}

/** Help text for Joker Poker CLI mode. */
export const JOKERPOKER_HELP: string[] = [...VIDEO_POKER_HELP, 'log         - Show action log'];
