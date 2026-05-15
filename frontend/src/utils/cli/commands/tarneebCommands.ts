import type { tarneebApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type TarneebArgs = Parameters<typeof tarneebApi.exec>;

const EXTRA_COMMANDS = ['bid', 'trump'];

/** Parse a Tarneeb CLI command into API exec arguments. */
export function parseTarneebCommand(input: string): CliParseResult<TarneebArgs> {
  // `parseTrickCommand` is shared across trick-taking games and only carries
  // `bid` / `cardIndex`. Tarneeb's `trump` argument piggybacks on `bid`: the
  // tarneebApi.exec helper interprets `arg1` based on the command name, so a
  // single int slot is sufficient.
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    if (cmd === 'bid') {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: bid <n>  (0=pass, 7-13=bid)' };
      return { command: 'bid', bid: parsed.value };
    }
    if (cmd === 'trump') {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: trump <suit>  (1=Spade 2=Club 3=Heart 4=Diamond)' };
      return { command: 'trump', bid: parsed.value };
    }
    return null;
  });

  if ('error' in result) return { error: result.error };
  if (result.command === 'bid' || result.command === 'trump') {
    return { args: [result.command as TarneebArgs[0], result.bid] };
  }
  return { args: [result.command as TarneebArgs[0], undefined, result.cardIndex] };
}

/** Help text for Tarneeb CLI mode. */
export const TARNEEB_HELP: string[] = [
  'bid <n>     - Place bid (0=pass, 7-13=bid)',
  'trump <s>   - Declare trump suit (1=Spade 2=Club 3=Heart 4=Diamond)',
  ...TRICK_HELP,
];
