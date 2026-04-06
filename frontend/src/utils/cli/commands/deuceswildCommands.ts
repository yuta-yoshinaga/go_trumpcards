import type { deuceswildApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseVideoPokerCommand, VIDEO_POKER_HELP } from './sharedVideoPokerCommands';

type DeuceswildArgs = Parameters<typeof deuceswildApi.exec>;

/** Parse a Deuces Wild CLI command into API exec arguments. */
export function parseDeuceswildCommand(input: string): CliParseResult<DeuceswildArgs> {
  const result = parseVideoPokerCommand(input);
  if ('error' in result) return { error: result.error };
  return { args: [result.command as DeuceswildArgs[0], result.amount, result.indices] };
}

/** Help text for Deuces Wild CLI mode. */
export const DEUCESWILD_HELP: string[] = [...VIDEO_POKER_HELP, 'log         - Show action log'];
