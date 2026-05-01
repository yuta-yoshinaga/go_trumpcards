import type { shitheadApi } from '../../../api/gameApi';
import { parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ShitheadArgs = Parameters<typeof shitheadApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'pu', 'pickup', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Shithead CLI command into API exec arguments. */
export function parseShitheadCommand(input: string): CliParseResult<ShitheadArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: p <idx...> (e.g. p 0 2)' };
      return { args: ['play', { indices: parsed.values }] };
    }
    case 'pu':
    case 'pickup':
      // Pickup = play with no indices (server interprets empty as pickup)
      return { args: ['play', { indices: [] }] };
    case 'log':
      return { args: ['log'] };
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

/** Help text for Shithead CLI mode. */
export const SHITHEAD_HELP: string[] = [
  'p <idx...>  - Play card(s) by index (e.g. p 0 2 3)',
  'pu/pickup   - Pick up the discard pile',
  'log         - Show action log',
  'r/reset     - Reset game',
];
