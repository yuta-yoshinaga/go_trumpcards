import type { beloteApi } from '../../../api/games/belote';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by beloteApi.exec. */
export type BeloteCliArgs = Parameters<typeof beloteApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'orderup',
  'pass',
  'calltrump',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
];

/** Parse a Belote CLI command into beloteApi.exec arguments. */
export function parseBeloteCommand(input: string): CliParseResult<BeloteCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const index = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(index) || index < 0 || index > 7) return { error: 'Usage: p <idx>' };
      return { args: ['play', undefined, index] };
    }
    case 'orderup':
      return { args: ['orderup', undefined] };
    case 'pass':
      return { args: ['pass', undefined] };
    case 'calltrump': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit) || suit < 1 || suit > 4) return { error: 'Usage: calltrump <1-4>' };
      return { args: ['calltrump', undefined, suit] };
    }
    case 'n':
    case 'next':
      return { args: ['next', undefined] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround', undefined] };
    case 'h':
    case 'hint':
      return { args: ['hint', undefined] };
    case 'r':
    case 'reset':
      return { args: ['reset'] as unknown as BeloteCliArgs };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the CLI terminal for Belote. */
export const BELOTE_HELP: string[] = [
  'p <idx>          - Play a card by hand index',
  'orderup          - Take the face-up trump card',
  'pass             - Pass the bid',
  'calltrump <1-4>  - Choose trump (1=Spades 2=Clubs 3=Hearts 4=Diamonds)',
  'n / next         - Next trick',
  'nr / nextround   - Next round',
  'h / hint         - Show hint',
  'r / reset        - Reset game',
];
