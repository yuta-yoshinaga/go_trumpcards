import type { tonkApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TonkArgs = Parameters<typeof tonkApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'dis',
  'discard',
  'kn',
  'knock',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Tonk CLI command into API exec arguments. */
export function parseTonkCommand(input: string): CliParseResult<TonkArgs> {
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

/** Help text for Tonk CLI mode. */
export const TONK_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'dis <idx>      - Discard a card',
  'kn <idx>       - Knock (deadwood <= 5, discard idx)',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
