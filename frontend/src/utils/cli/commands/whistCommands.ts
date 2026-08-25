import type { whistApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type WhistArgs = Parameters<typeof whistApi.exec>;

/** Parse a Whist CLI command into API exec arguments. */
export function parseWhistCommand(input: string): CliParseResult<WhistArgs> {
  const result = parseTrickCommand(input);

  if ('error' in result) return { error: result.error };
  // **札のインデックスは 2 番目のスロット。** クライアントは
  // `useTrickGameBase` と同じ並び `(command, arg1, arg2, config)` を取る (#6227)。
  return { args: [result.command as WhistArgs[0], undefined, result.cardIndex] };
}

/** Help text for Whist CLI mode. */
export const WHIST_HELP: string[] = [...TRICK_HELP];
