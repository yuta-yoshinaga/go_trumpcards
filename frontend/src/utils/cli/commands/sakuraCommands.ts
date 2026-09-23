import type { sakuraApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SakuraArgs = Parameters<typeof sakuraApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'ss',
  'setseats',
  'sr',
  'setrounds',
  'h',
  'hint',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Sakura (さくら) CLI command into API exec arguments. */
export function parseSakuraCommand(input: string): CliParseResult<SakuraArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const hand = parseIntArg(args, 0);
      if ('error' in hand) return { error: 'Usage: p <handIdx> [fieldIdx]' };
      if (args.length > 1) {
        const field = Number.parseInt(args[1], 10);
        if (Number.isNaN(field)) return { error: `Invalid field index: ${args[1]}` };
        return { args: ['play', { cardIndex: hand.value, fieldIndex: field }] };
      }
      return { args: ['play', { cardIndex: hand.value }] };
    }
    case 'n':
    case 'next':
    case 'nr':
    case 'nextround':
      return { args: ['next'] };
    case 'ss':
    case 'setseats': {
      const value = parseIntArg(args, 0);
      return 'error' in value ? { error: 'Usage: ss <2-4>' } : { args: ['reset', { config: { seats: value.value } }] };
    }
    case 'sr':
    case 'setrounds': {
      const value = parseIntArg(args, 0);
      return 'error' in value
        ? { error: 'Usage: sr <1-12>' }
        : { args: ['reset', { config: { rounds: value.value } }] };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'log':
    case 'l':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      return {
        error: suggestion ? `Unknown command: ${cmd}. Did you mean: ${suggestion}?` : `Unknown command: ${cmd}`,
      };
    }
  }
}

/** Help text for Sakura CLI mode. */
export const SAKURA_HELP: string[] = [
  'p <h> [f]    - Play hand card h, choosing field card f when needed',
  'n/nr/next    - Deal the next round',
  'h/hint       - Show hint',
  'l/log         - Show action log',
  'ss <2-4>      - Set player count and reset',
  'sr <1-12>     - Set round count and reset',
  'r/reset       - Reset game',
];
