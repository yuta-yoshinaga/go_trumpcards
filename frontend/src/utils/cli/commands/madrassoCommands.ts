import type { madrassoApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MadrassoArgs = Parameters<typeof madrassoApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'nr', 'nextround', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Madrasso CLI command into API exec arguments. */
export function parseMadrassoCommand(input: string): CliParseResult<MadrassoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
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

/** Help text for Madrasso CLI mode. */
export const MADRASSO_HELP: string[] = [
  'p <idx>          - Play a card (must follow the led suit)',
  'n/next           - Next trick',
  'nr/nextround     - Next round',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
