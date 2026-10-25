import type { jassApi } from '../../../api/games/jass';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by jassApi.exec. */
export type JassCliArgs = Parameters<typeof jassApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'ct',
  'calltrump',
  's',
  'schieben',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
];

/** Parse CLI input into arguments accepted by jassApi.exec. */
export function parseJassCommand(input: string): CliParseResult<JassCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number(args[0]);
      if (!Number.isInteger(idx) || idx < 0 || idx > 8) return { error: 'Usage: p <idx> (0-8)' };
      return { args: ['play', undefined, idx] };
    }
    case 'ct':
    case 'calltrump': {
      const suit = Number(args[0]);
      if (!Number.isInteger(suit) || suit < 1 || suit > 4) return { error: 'Usage: ct <suit> (1-4)' };
      return { args: ['calltrump', suit] };
    }
    case 's':
    case 'schieben':
      return { args: ['schieben'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Jass CLI terminal. */
export const JASS_HELP = [
  'p <idx>          - Play a card from your hand (0-8)',
  'ct <suit>        - Call trump (1=spades, 2=clubs, 3=hearts, 4=diamonds)',
  's / schieben     - Pass the trump call to your partner',
  'n / next         - Continue after a trick',
  'nr / nextround   - Start the next round',
  'h / hint         - Request a hint',
  'r / reset        - Reset game',
];
