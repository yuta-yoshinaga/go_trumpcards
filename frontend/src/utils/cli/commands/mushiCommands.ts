import type { mushiApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MushiArgs = Parameters<typeof mushiApi.exec>;

const VALID_COMMANDS = ['p', 'play', 's', 'select', 'n', 'next', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Mushi CLI command into API exec arguments. */
export function parseMushiCommand(input: string): CliParseResult<MushiArgs> {
  const { cmd, args } = splitCommand(input);

  const index = (label: string): number | string => {
    if (args.length === 0) return `Usage: ${cmd} <index>`;
    const n = Number(args[0]);
    if (!Number.isInteger(n) || n < 0) return `Invalid ${label} index: ${args[0]}`;
    return n;
  };

  switch (cmd) {
    case 'p':
    case 'play': {
      const n = index('card');
      return typeof n === 'string' ? { error: n } : { args: ['play', n] };
    }
    case 's':
    case 'select': {
      // `select` names a FIELD index, not a hand index -- passed in the third
      // position so it does not collide with `play`'s card index.
      const n = index('field');
      return typeof n === 'string' ? { error: n } : { args: ['select', undefined, n] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text for Mushi CLI mode. */
export const MUSHI_HELP: string[] = [
  'p <i>           - Play the hand card at index i',
  's <i>           - Take the field card at index i',
  'n/next          - Start the next round',
  'log             - Show action log',
  'r/reset         - New game',
];
