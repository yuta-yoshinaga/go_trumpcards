import type { bauernschnapsenApi } from '../../../api/gameApi';
import { splitCommand, unknownCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Arguments accepted by the Bauernschnapsen exec API. */
export type BauernschnapsenCliArgs = Parameters<typeof bauernschnapsenApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'c',
  'contract',
  'm',
  'marriage',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
];

/** Parse CLI input into arguments accepted by bauernschnapsenApi.exec. */
export function parseBauernschnapsenCommand(input: string): CliParseResult<BauernschnapsenCliArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number(args[0]);
      if (!Number.isInteger(idx) || idx < 0 || idx > 4) return { error: 'Usage: p <idx> (0-4)' };
      return { args: ['play', undefined, idx] };
    }
    case 'c':
    case 'contract': {
      const contract = Number(args[0]);
      const suit = contract === 0 || contract === 3 ? 1 : args[1] === undefined ? 1 : Number(args[1]);
      if (
        !Number.isInteger(contract) ||
        contract < 0 ||
        contract > 3 ||
        !Number.isInteger(suit) ||
        suit < 1 ||
        suit > 4 ||
        ((contract === 1 || contract === 2) && args[1] === undefined)
      ) {
        return { error: 'Usage: c <contract> [suit] (0=Pass, 1=Rufer, 2=Farbenzwang, 3=Bettel; suit 1-4)' };
      }
      return { args: ['contract', contract, undefined, undefined, suit] };
    }
    case 'm':
    case 'marriage': {
      const idx = Number(args[0]);
      if (!Number.isInteger(idx) || idx < 0 || idx > 4) return { error: 'Usage: m <idx> (0-4)' };
      return { args: ['marriage', undefined, idx] };
    }
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

/** Help text shown in the Bauernschnapsen CLI terminal. */
export const BAUERNSCHNAPSEN_HELP = [
  'p <idx>            - Play a card from your hand (0-4)',
  'c <0-3> [suit]     - Declare a contract (0=Pass, 1=Rufer, 2=Farbenzwang, 3=Bettel; suit 1-4)',
  'm <idx>            - Declare a marriage with a card (0-4)',
  'n / next           - Continue after a trick',
  'nr / nextround     - Start the next round',
  'h / hint           - Request a hint',
  'r / reset          - Reset game',
];
