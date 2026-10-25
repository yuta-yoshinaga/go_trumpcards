import type { ultimatetexasholdemApi } from '../../../api/gameApi';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args accepted by ultimatetexasholdemApi.exec. */
export type UltimateTexasHoldemCliArgs = Parameters<typeof ultimatetexasholdemApi.exec>;

const VALID_COMMANDS = ['b', 'bet', '4', '3', '2', '1', 'play', 'c', 'check', 'f', 'fold', 'r', 'reset', 'help', '?'];

/** Parse a CLI command into Ultimate Texas Hold'em API arguments. */
export function parseUltimateTexasHoldemCommand(input: string): CliParseResult<UltimateTexasHoldemCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'b':
    case 'bet': {
      const ante = Number(args[0]);
      const trips = args[1] === undefined ? 0 : Number(args[1]);
      if (!Number.isInteger(ante) || ante <= 0 || !Number.isInteger(trips) || trips < 0) {
        return { error: 'Usage: bet <ante> [trips]' };
      }
      return { args: ['bet', ante, trips] };
    }
    case 'play': {
      const multiplier = Number(args[0]);
      if (!Number.isInteger(multiplier) || multiplier < 1 || multiplier > 4) return { error: 'Usage: play <1|2|3|4>' };
      return { args: ['play', undefined, undefined, multiplier] };
    }
    case '1':
    case '2':
    case '3':
    case '4':
      return { args: ['play', undefined, undefined, Number(cmd)] };
    case 'c':
    case 'check':
      return { args: ['check'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default:
      return unknownCommand(cmd, VALID_COMMANDS);
  }
}

/** Help text shown in the Ultimate Texas Hold'em terminal. */
export const ULTIMATE_TEXAS_HOLDEM_HELP = [
  'bet <ante> [trips] - Place Ante and optional Trips bets',
  '4 / 3 / 2 / 1      - Play the available multiplier',
  'c / check          - Check',
  'f / fold           - Fold',
  'h / hint           - Request a hint',
  'r / reset          - Reset game',
];
