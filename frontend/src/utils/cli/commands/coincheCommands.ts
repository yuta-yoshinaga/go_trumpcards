import type { coincheApi } from '../../../api/games/coinche';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

export type CoincheCliArgs = Parameters<typeof coincheApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bid',
  'p',
  'play',
  'pass',
  'c',
  'coinche',
  'sc',
  'surcoinche',
  'd',
  'decline',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
];

/** Parse Coinche terminal input into the endpoint's positional exec arguments. */
export function parseCoincheCommand(input: string): CliParseResult<CoincheCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'b':
    case 'bid': {
      const points = Number.parseInt(args[0] ?? '', 10);
      const suit = Number.parseInt(args[1] ?? '', 10);
      if (
        ![80, 90, 100, 110, 120, 130, 140, 150, 160, 250].includes(points) ||
        suit < 1 ||
        suit > 4 ||
        Number.isNaN(suit)
      )
        return { error: 'Usage: bid <points> <suit 1-4>' };
      return { args: ['bid', points, suit] };
    }
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx) || idx < 0 || idx > 7) return { error: 'Usage: p <idx 0-7>' };
      return { args: ['play', undefined, idx] };
    }
    case 'pass':
      return { args: ['pass'] };
    case 'c':
    case 'coinche':
      return { args: ['coinche'] };
    case 'sc':
    case 'surcoinche':
      return { args: ['surcoinche'] };
    case 'd':
    case 'decline':
      return { args: ['decline'] };
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

export const COINCHE_HELP = [
  'bid <points> <suit 1-4> - Bid a contract',
  'p <idx 0-7> - Play a card',
  'pass - Pass',
  'c / coinche - Double',
  'sc / surcoinche - Redouble',
  'd / decline - Decline',
  'n / next - Next trick',
  'nr / nextround - Next round',
  'h / hint - Show hint',
  'r / reset - Reset game',
];
