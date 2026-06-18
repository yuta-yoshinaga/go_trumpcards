import type { conquianApi } from '../../../api/gameApi';
import { parseIntArg, parseIntSlice, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ConquianArgs = Parameters<typeof conquianApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'meld',
  'm',
  'dis',
  'discard',
  'd',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Conquian CLI command into API exec arguments. */
export function parseConquianCommand(input: string): CliParseResult<ConquianArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'ds':
    case 'drawstock':
      return { args: ['drawstock'] };
    case 'dd':
    case 'drawdiscard':
      return { args: ['drawdiscard'] };
    case 'm':
    case 'meld': {
      // meld <idx...> — all indices form one meld group
      const parsed = parseIntSlice(args);
      if ('error' in parsed) return { error: 'Usage: meld <idx...>' };
      return { args: ['meld', undefined, undefined, [parsed.values]] };
    }
    case 'd':
    case 'dis':
    case 'discard': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: d <idx>' };
      return { args: ['discard', parsed.value] };
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

/** Help text for Conquian CLI mode. */
export const CONQUIAN_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'meld <idx...>  - Lay down a meld',
  'd <idx>        - Discard a card',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
