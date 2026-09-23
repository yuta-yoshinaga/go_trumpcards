import type { binokelApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import { STANDARD_SUIT_MAP as SUIT_MAP } from '../suitMaps';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type BinokelArgs = Parameters<typeof binokelApi.exec>;

const EXTRA_COMMANDS = ['b', 'bid', 'pa', 'pass', 'd', 'discard', 't', 'trump', 'm', 'meld', 'l', 'log'];

/** Parse a Binokel CLI command into API exec arguments. */
export function parseBinokelCommand(input: string): CliParseResult<BinokelArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'b':
      case 'bid': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: bid <amount>' };
        if (parsed.value < 150 || (parsed.value - 150) % 10 !== 0) {
          return { error: 'Invalid bid amount. Must be 150 or more in steps of 10' };
        }
        return { command: `bid:${parsed.value}` };
      }
      case 'pa':
      case 'pass':
        return { command: 'pass' };
      case 'd':
      case 'discard': {
        if (args.length < 3) return { error: 'Usage: discard <i> <j> <k>' };
        const a = parseIntArg(args, 0);
        const b = parseIntArg(args, 1);
        const c = parseIntArg(args, 2);
        if ('error' in a || 'error' in b || 'error' in c) return { error: 'Usage: discard <i> <j> <k>' };
        return { command: `discard:${a.value},${b.value},${c.value}` };
      }
      case 't':
      case 'trump': {
        if (args.length === 0) return { error: 'Usage: trump <suit>' };
        const suitArg = args[0].toLowerCase();
        let suit = SUIT_MAP[suitArg];
        if (suit === undefined) {
          const num = Number(suitArg);
          if (num >= 1 && num <= 4) {
            suit = num;
          }
        }
        if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
        return { command: `trump:${suit}` };
      }
      case 'm':
      case 'meld':
        return { command: 'meld' };
      case 'l':
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
  if (cmd.startsWith('discard:')) {
    const indices = cmd.split(':')[1].split(',').map(Number);
    return { args: ['discard', undefined, undefined, undefined, undefined, indices] };
  }
  if (cmd.startsWith('trump:')) {
    return { args: ['trump', undefined, undefined, undefined, Number(cmd.split(':')[1])] };
  }
  return { args: [cmd as BinokelArgs[0], result.cardIndex] };
}

/** Help text for Binokel CLI mode. */
export const BINOKEL_HELP: string[] = [
  'b/bid <amt>      - Place bid (150+, step 10)',
  'pa/pass          - Pass on bidding',
  'd <i> <j> <k>    - Discard 3 cards to Dabb',
  't/trump <suit>   - Declare trump (1-4 or suit name)',
  'm/meld           - Confirm melds',
  'l/log            - Show action log',
  ...TRICK_HELP,
];
