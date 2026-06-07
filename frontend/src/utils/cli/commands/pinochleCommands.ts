import type { pinochleApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import { STANDARD_SUIT_MAP as SUIT_MAP } from '../suitMaps';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type PinochleArgs = Parameters<typeof pinochleApi.exec>;

const EXTRA_COMMANDS = ['bid', 'pass', 'trump', 'meld', 'log'];

/** Parse a Pinochle CLI command into API exec arguments. */
export function parsePinochleCommand(input: string): CliParseResult<PinochleArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'bid': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: bid <amount>' };
        return { command: `bid:${parsed.value}` };
      }
      case 'pass':
        return { command: 'pass' };
      case 'trump': {
        if (args.length === 0) return { error: 'Usage: trump <suit>' };
        const suit = SUIT_MAP[args[0].toLowerCase()];
        if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
        return { command: `trump:${suit}` };
      }
      case 'meld':
        return { command: 'meld' };
      case 'log':
        return { command: 'log' };
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };

  const cmd = result.command;
  if (cmd === 'pass') return { args: ['pass'] };
  if (cmd === 'meld') return { args: ['meld'] };
  if (cmd === 'log') return { args: ['log'] };
  if (cmd.startsWith('bid:')) {
    return { args: ['bid', undefined, undefined, Number(cmd.split(':')[1])] };
  }
  if (cmd.startsWith('trump:')) {
    return { args: ['trump', undefined, undefined, undefined, Number(cmd.split(':')[1])] };
  }
  return { args: [cmd as PinochleArgs[0], result.cardIndex] };
}

/** Help text for Pinochle CLI mode. */
export const PINOCHLE_HELP: string[] = [
  'bid <amt>   - Place bid',
  'pass        - Pass on bidding',
  'trump <suit>- Declare trump',
  'meld        - Show melds',
  'log         - Show action log',
  ...TRICK_HELP,
];
