import type { doubtApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DoubtArgs = Parameters<typeof doubtApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'doubt', 'skip', 'r', 'reset', 'help', '?'];

/** Parse a Doubt CLI command into API exec arguments. */
export function parseDoubtCommand(input: string): CliParseResult<DoubtArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length < 2) return { error: 'Usage: p <claimedValue> <idx...> (e.g., p 5 0 1)' };
      const claimed = parseIntArg(args, 0);
      if ('error' in claimed) return { error: 'Usage: p <claimedValue> <idx...>' };
      const indices = parseIntSlice(args.slice(1));
      if ('error' in indices) return { error: `Invalid indices: ${indices.error}` };
      return { args: ['play', indices.values, claimed.value] };
    }
    case 'doubt':
      return { args: ['doubt'] };
    case 'skip':
      return { args: ['skip'] };
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

/** Help text for Doubt CLI mode. */
export const DOUBT_HELP: string[] = [
  'p <val> <idx...> - Play cards claiming value',
  'doubt       - Call doubt',
  'skip        - Skip doubt',
  'r/reset     - Reset game',
];
