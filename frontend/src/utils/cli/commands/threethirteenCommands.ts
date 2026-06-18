import type { threethirteenApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ThreeThirteenArgs = Parameters<typeof threethirteenApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'd',
  'discard',
  'k',
  'knock',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Three Thirteen CLI command into API exec arguments. */
export function parseThreeThirteenCommand(input: string): CliParseResult<ThreeThirteenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'd':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['discard', parsed.value] };
    }
    case 'k':
    case 'knock': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: k <idx> (discard card idx)' };
      return { args: ['knock', parsed.value] };
    }
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
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

/** Help text for Three Thirteen CLI mode. */
export const THREETHIRTEEN_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'd <idx>        - Discard a card',
  'k <idx>        - Knock (discard idx)',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
