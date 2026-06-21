import type { chinchonApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ChinchonArgs = Parameters<typeof chinchonApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'dis',
  'discard',
  'kn',
  'knock',
  'lo',
  'layoff',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Chinchón CLI command into API exec arguments. */
export function parseChinchonCommand(input: string): CliParseResult<ChinchonArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'dis':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: dis <idx>' };
      return { args: ['discard', parsed.value] };
    }
    case 'kn':
    case 'knock': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: kn <idx> (discard card idx)' };
      return { args: ['knock', parsed.value] };
    }
    case 'lo':
    case 'layoff': {
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: lo <idx...>' };
      return { args: ['layoff', undefined, undefined, parsed.values] };
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

/** Help text for Chinchón CLI mode. */
export const CHINCHON_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'dis <idx>      - Discard a card',
  'kn <idx>       - Knock (discard idx)',
  'lo <idx...>    - Layoff cards',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
