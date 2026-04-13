import type { whistApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type WhistArgs = Parameters<typeof whistApi.exec>;

/** Parse a Whist CLI command into API exec arguments. */
export function parseWhistCommand(input: string): CliParseResult<WhistArgs> {
  const result = parseTrickCommand(input);

  if ('error' in result) return { error: result.error };
  return { args: [result.command as WhistArgs[0], result.cardIndex] };
}

/** Help text for Whist CLI mode. */
export const WHIST_HELP: string[] = [...TRICK_HELP];
