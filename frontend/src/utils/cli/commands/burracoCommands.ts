import type { burracoApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BurracoArgs = Parameters<typeof burracoApi.exec>;

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
  'h',
  'hint',
  'help',
  '?',
];

/** Parse a Burraco CLI command into API exec arguments. */
export function parseBurracoCommand(input: string): CliParseResult<BurracoArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'meld': {
      // meld <idx,idx,idx> [<idx,idx,idx>] — groups separated by spaces, commas within
      // Simplified: meld <idx...> — all indices as one group
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

/** Help text for Burraco CLI mode. */
export const BURRACO_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'meld <idx...>  - Meld cards',
  'sm/skipmeld    - Skip meld phase',
  'dis <idx>      - Discard a card',
  'go/goout       - Go out',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
  'h/hint         - Get a hint',
];
