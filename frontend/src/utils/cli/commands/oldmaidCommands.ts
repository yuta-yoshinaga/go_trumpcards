import type { oldmaidApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type OldMaidArgs = Parameters<typeof oldmaidApi.exec>;

const VALID_COMMANDS = ['d', 'draw', 'sh', 'shuffle', 'ro', 'reorder', 'r', 'reset', 'help', '?'];

/** Parse an Old Maid CLI command into API exec arguments. */
export function parseOldmaidCommand(input: string): CliParseResult<OldMaidArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['draw', parsed.value] };
    }
    case 'sh':
    case 'shuffle':
      return { args: ['shuffle'] };
    case 'ro':
    case 'reorder': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: ro <idx...>' };
      return { args: ['reorder', undefined, undefined, undefined, parsed.values] };
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

/** Help text for Old Maid CLI mode. */
export const OLDMAID_HELP: string[] = [
  'd <idx>     - Draw a card from opponent',
  'sh/shuffle  - Shuffle your hand',
  'ro <idx...> - Reorder hand (e.g., ro 2 0 1 3)',
  'r/reset     - Reset game',
];
