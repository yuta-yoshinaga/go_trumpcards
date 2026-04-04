import type { napoleonApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type NapoleonArgs = Parameters<typeof napoleonApi.exec>;

const SUIT_MAP: Record<string, number> = {
  spade: 1,
  spades: 1,
  s: 1,
  clover: 2,
  clubs: 2,
  c: 2,
  heart: 3,
  hearts: 3,
  h: 3,
  diamond: 4,
  diamonds: 4,
  d: 4,
};

const EXTRA_COMMANDS = ['bid', 'pass', 'trump', 'adj', 'adjutant', 'ex', 'exchange', 'log'];

/** Parse a Napoleon CLI command into API exec arguments. */
export function parseNapoleonCommand(input: string): CliParseResult<NapoleonArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'bid': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: bid <n>' };
        return { command: `bid:${parsed.value}` };
      }
      case 'pass':
        return { command: 'pass' };
      case 'trump': {
        if (args.length < 1) return { error: 'Usage: trump <suit>' };
        const suit = SUIT_MAP[args[0].toLowerCase()];
        if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
        return { command: `trump:${suit}` };
      }
      case 'adj':
      case 'adjutant': {
        if (args.length < 2) return { error: 'Usage: adj <suit> <value>' };
        const suit = SUIT_MAP[args[0].toLowerCase()];
        if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
        const val = parseIntArg(args, 1);
        if ('error' in val) return { error: 'Usage: adj <suit> <value>' };
        return { command: `adj:${suit}:${val.value}` };
      }
      case 'ex':
      case 'exchange': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: ex <idx>' };
        return { command: `exchange:${parsed.value}` };
      }
      case 'log':
        return { command: 'log' };
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };

  const cmd = result.command;
  if (cmd === 'pass') return { args: ['bid', 0] };
  if (cmd === 'log') return { args: ['log'] };
  if (cmd.startsWith('bid:')) {
    return { args: ['bid', Number(cmd.split(':')[1])] };
  }
  if (cmd.startsWith('trump:')) {
    const suit = Number(cmd.split(':')[1]);
    return { args: ['trump', undefined, suit] };
  }
  if (cmd.startsWith('adj:')) {
    const parts = cmd.split(':');
    return { args: ['trump', undefined, undefined, Number(parts[1]), Number(parts[2])] };
  }
  if (cmd.startsWith('exchange:')) {
    const idx = Number(cmd.split(':')[1]);
    return { args: ['exchange', undefined, undefined, undefined, undefined, idx] };
  }
  if (result.command === 'play') {
    return { args: ['play', undefined, undefined, undefined, undefined, undefined, result.cardIndex] };
  }
  return { args: [cmd as NapoleonArgs[0]] };
}

/** Help text for Napoleon CLI mode. */
export const NAPOLEON_HELP: string[] = [
  'bid <n>     - Place bid (0 to pass)',
  'pass        - Pass on bidding',
  'trump <suit>- Declare trump',
  'adj <s> <v> - Declare adjutant card',
  'ex <idx>    - Exchange card with kitty',
  'log         - Show action log',
  ...TRICK_HELP,
];
