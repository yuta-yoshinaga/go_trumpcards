import type { boliviaApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BoliviaArgs = Parameters<typeof boliviaApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'meld',
  'sm',
  'skipmeld',
  'dis',
  'discard',
  'go',
  'goout',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Bolivia CLI command into API exec arguments. */
export function parseBoliviaCommand(input: string): CliParseResult<BoliviaArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard': {
      // drawdiscard <idx,idx> — natural pair indices matching the discard top
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: dd <idx,idx>' };
      return { args: ['drawdiscard', undefined, undefined, parsed.values] };
    }
    case 'meld': {
      // meld <idx...> — indices forming one meld (set or sequence)
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: meld <idx...>' };
      return { args: ['meld', undefined, undefined, undefined, [parsed.values]] };
    }
    case 'sm':
    case 'skipmeld':
      return { args: ['skipmeld'] };
    case 'dis':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: dis <idx>' };
      return { args: ['discard', parsed.value] };
    }
    case 'go':
    case 'goout':
      return { args: ['goout'] };
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

/** Help text for Bolivia CLI mode. */
export const BOLIVIA_HELP: string[] = [
  'ds/drawstock    - Draw from stock',
  'dd <idx,idx>    - Take the discard pile (natural pair indices)',
  'meld <idx...>   - Meld cards (set or sequence)',
  'sm/skipmeld     - Skip meld phase',
  'dis <idx>       - Discard a card',
  'go/goout        - Go out',
  'nr/nextround    - Next round',
  'log             - Show action log',
  'r/reset         - Reset game',
];
