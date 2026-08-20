import type { chinesetenApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ChineseTenArgs = Parameters<typeof chinesetenApi.exec>;

const VALID_COMMANDS = ['p', 'play', 's', 'select', 'log', 'r', 'reset', 'help', '?'];

/** Parse a Chinese Ten CLI command into API exec arguments. */
export function parseChineseTenCommand(input: string): CliParseResult<ChineseTenArgs> {
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
      // `select` names a LAYOUT index, passed in the third position so it can
      // never be read as `play`'s hand index.
      const n = index('layout');
      return typeof n === 'string' ? { error: n } : { args: ['select', undefined, n] };
    }
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Chinese Ten CLI mode. */
export const CHINESETEN_HELP: string[] = [
  'p <i>           - Play the hand card at index i',
  's <i>           - Take the layout card at index i',
  'log             - Show action log',
  'r/reset         - New game',
  'h/hint      - Get a hint',
];
