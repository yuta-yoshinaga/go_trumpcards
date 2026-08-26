import type { catchtenApi } from '../../../api/gameApi';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type CatchTenArgs = Parameters<typeof catchtenApi.exec>;

/** Parse a Catch the Ten CLI command into API exec arguments. */
export function parseCatchTenCommand(input: string): CliParseResult<CatchTenArgs> {
  const result = parseTrickCommand(input);

  if ('error' in result) return { error: result.error };
  // **札のインデックスは 2 番目のスロット。** クライアントは
  // `useTrickGameBase` と同じ並び `(command, arg1, arg2, config)` を取る (#6227)。
  return { args: [result.command as CatchTenArgs[0], undefined, result.cardIndex] };
}

/** Help text for Catch the Ten CLI mode. */
export const CATCHTEN_HELP: string[] = [...TRICK_HELP];
