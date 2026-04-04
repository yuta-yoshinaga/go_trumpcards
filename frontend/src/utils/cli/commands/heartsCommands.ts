import type { heartsApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type HeartsArgs = Parameters<typeof heartsApi.exec>;

const VALID_COMMANDS = ['pass', 'p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Hearts CLI command into API exec arguments. */
export function parseHeartsCommand(input: string): CliParseResult<HeartsArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'pass': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: `Invalid indices: ${parsed.error}` };
      if (parsed.values.length !== 3) return { error: 'Usage: pass <idx> <idx> <idx> (exactly 3 cards)' };
      return { args: ['pass', parsed.values] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', undefined, parsed.value] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Hearts CLI mode. */
export const HEARTS_HELP: string[] = [
  'pass <i> <i> <i> - Pass 3 cards',
  'p <idx>          - Play a card',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
