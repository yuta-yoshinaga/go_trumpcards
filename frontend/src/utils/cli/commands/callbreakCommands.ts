import type { callBreakApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type CallBreakArgs = Parameters<typeof callBreakApi.exec>;

const EXTRA_COMMANDS = ['bid'];

/** Parse a Call Break CLI command into API exec arguments. */
export function parseCallBreakCommand(input: string): CliParseResult<CallBreakArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    if (cmd === 'bid') {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: bid <n>' };
      return { command: 'bid', bid: parsed.value };
    }
    return null;
  });

  if ('error' in result) return { error: result.error };
  if (result.command === 'bid') {
    return { args: ['bid', result.bid] };
  }
  return { args: [result.command as CallBreakArgs[0], undefined, result.cardIndex] };
}

/** Help text for Call Break CLI mode. */
export const CALLBREAK_HELP: string[] = ['bid <n>     - Place bid (1-13)', ...TRICK_HELP];
