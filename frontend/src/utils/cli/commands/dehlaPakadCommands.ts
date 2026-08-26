import type { dehlaPakadApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DehlaPakadArgs = Parameters<typeof dehlaPakadApi.exec>;

const VALID_COMMANDS = ['t', 'trump', 'p', 'play', 'nh', 'nexthand', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Dehla Pakad CLI command into API exec arguments. */
export function parseDehlaPakadCommand(input: string): CliParseResult<DehlaPakadArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'trump': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: t <1-4>' };
      return { args: ['trump', { trumpSuit: parsed.value }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'nh':
    case 'nexthand':
      return { args: ['nexthand'] };
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

/** Help text for Dehla Pakad CLI mode. */
export const DEHLA_PAKAD_HELP: string[] = [
  't <1-4>                          - Call trump (1=spade 2=club 3=heart 4=diamond)',
  'p <idx>                          - Play a card',
  'nh/nexthand                      - Next hand',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
