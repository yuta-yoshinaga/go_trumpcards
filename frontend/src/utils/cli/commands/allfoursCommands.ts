import type { allfoursApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type AllFoursArgs = Parameters<typeof allfoursApi.exec>;

const VALID_COMMANDS = [
  'stand',
  'st',
  'beg',
  'bg',
  'gift',
  'g',
  'run',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse an All Fours CLI command into API exec arguments. */
export function parseAllFoursCommand(input: string): CliParseResult<AllFoursArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'stand':
    case 'st':
      return { args: ['beg', false] };
    case 'beg':
    case 'bg':
      return { args: ['beg', true] };
    case 'gift':
    case 'g':
      return { args: ['respond', undefined, false] };
    case 'run':
      return { args: ['respond', undefined, true] };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: play <index>' };
      return { args: ['play', undefined, undefined, idx.value] };
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

/** Help text for All Fours CLI mode. */
export const ALLFOURS_HELP: string[] = [
  'stand / st   - Stand (non-dealer)',
  'beg / bg     - Beg for a new trump (non-dealer)',
  'gift / g     - Gift 1 point (dealer reply to beg)',
  'run          - Run the cards / re-deal (dealer reply)',
  'p <i>        - Play card by index',
  'n / next     - Next trick',
  'nr           - Next deal (score the deal first)',
  'h / hint     - Show hint',
  'r / reset    - Reset game',
];
