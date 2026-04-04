import type { daifugoApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DaifugoArgs = Parameters<typeof daifugoApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'sort', 'r', 'reset', 'help', '?'];

/** Parse a Daifugo CLI command into API exec arguments. */
export function parseDaifugoCommand(input: string): CliParseResult<DaifugoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { args: ['play'] }; // pass
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: `Invalid indices: ${parsed.error}` };
      return { args: ['play', parsed.values] };
    }
    case 'sort': {
      if (args.length === 0) return { args: ['sort', undefined, undefined, 0] };
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: sort [0|1]' };
      return { args: ['sort', undefined, undefined, parsed.value] };
    }
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Daifugo CLI mode. */
export const DAIFUGO_HELP: string[] = [
  'p <idx...>  - Play cards (e.g., p 0 2)',
  'p           - Pass',
  'sort [0|1]  - Sort hand (0=strength, 1=number)',
  'r/reset     - Reset game',
];
