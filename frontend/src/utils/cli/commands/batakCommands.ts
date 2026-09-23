import type { batakApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type BatakArgs = Parameters<typeof batakApi.exec>;

const EXTRA_COMMANDS = ['bid', 'pass'];

/** Parse a Batak CLI command into API exec arguments. */
export function parseBatakCommand(input: string): CliParseResult<BatakArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    if (cmd === 'bid') {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: bid <n> (5-13, or 0 to pass)' };
      if (parsed.value !== 0 && (parsed.value < 5 || parsed.value > 13)) {
        return { error: 'Invalid bid: must be 5-13 or 0 (pass)' };
      }
      return { command: 'bid', bid: parsed.value };
    }
    if (cmd === 'pass') {
      return { command: 'bid', bid: 0 };
    }
    return null;
  });

  if ('error' in result) return { error: result.error };
  if (result.command === 'bid') {
    return { args: ['bid', result.bid] };
  }
  return { args: [result.command as BatakArgs[0], undefined, result.cardIndex] };
}

/** Help text for Batak CLI mode. */
export const BATAK_HELP: string[] = [
  'bid <n>     - Place bid (5-13, or 0 to pass)',
  'pass        - Pass on bidding',
  ...TRICK_HELP,
];
