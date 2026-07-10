import type { indianRummyApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type IndianRummyArgs = Parameters<typeof indianRummyApi.exec>;

const VALID_COMMANDS = [
  'ds',
  'drawstock',
  'dd',
  'drawdiscard',
  'dis',
  'discard',
  'de',
  'declare',
  'nr',
  'nextround',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse an Indian Rummy CLI command into API exec arguments. */
export function parseIndianrummyCommand(input: string): CliParseResult<IndianRummyArgs> {
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
    case 'de':
    case 'declare': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: de <idx> (discard 14th card to declare)' };
      return { args: ['declare', parsed.value] };
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

/** Help text for Indian Rummy CLI mode. */
export const INDIANRUMMY_HELP: string[] = [
  'ds/drawstock   - Draw from stock',
  'dd/drawdiscard - Draw from discard',
  'dis <idx>      - Discard a card',
  'de <idx>       - Declare (discard 14th and go out)',
  'nr/nextround   - Next round',
  'log            - Show action log',
  'r/reset        - Reset game',
];
