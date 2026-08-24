import type { quodlibetApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type QuodlibetArgs = Parameters<typeof quodlibetApi.exec>;

const VALID_COMMANDS = ['c', 'contract', 'p', 'play', 'pass', 'nd', 'nextdeal', 'h', 'hint', 'r', 'reset', 'help', '?'];

/** Parse a Quodlibet CLI command into API exec arguments. */
export function parseQuodlibetCommand(input: string): CliParseResult<QuodlibetArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'contract': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: c <0-11>' };
      return { args: ['contract', { contract: parsed.value }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'nd':
    case 'nextdeal':
      return { args: ['nextdeal'] };
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

/** Help text for Quodlibet CLI mode. */
export const QUODLIBET_HELP: string[] = [
  "c <0-11>                         - Choose this wheel's contract (dealer only)",
  'p <idx>                          - Play a card',
  'pass                             - Pass (Quadrature/Snack only, when nothing is playable)',
  'nd/nextdeal                      - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
