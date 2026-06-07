import type { mightyApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import { STANDARD_SUIT_MAP } from '../suitMaps';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type MightyArgs = Parameters<typeof mightyApi.exec>;

// Mighty also accepts the joker and a no-trump declaration, layered on the
// shared standard suit aliases.
const SUIT_MAP: Record<string, number> = {
  ...STANDARD_SUIT_MAP,
  joker: 0,
  j: 0,
  notrump: -1,
  nt: -1,
};

const EXTRA_COMMANDS = ['bid', 'pass', 'trump', 'friend', 'f', 'ex', 'exchange', 'jl', 'jokerlead', 'log'];

/** Parse a Mighty CLI command into API exec arguments. */
export function parseMightyCommand(input: string): CliParseResult<MightyArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'bid': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: bid <n> [nt]' };
        const noTrump = args.length >= 2 && args[1].toLowerCase() === 'nt';
        return { command: `bid:${parsed.value}:${noTrump ? '1' : '0'}` };
      }
      case 'pass':
        return { command: 'pass' };
      case 'trump': {
        if (args.length < 3) return { error: 'Usage: trump <trumpSuit> <partnerSuit> <partnerValue>' };
        const trump = SUIT_MAP[args[0].toLowerCase()];
        const pSuit = SUIT_MAP[args[1].toLowerCase()];
        if (trump === undefined || pSuit === undefined) {
          return { error: 'Invalid suit. Use: spade/clover/heart/diamond/joker/nt' };
        }
        const pVal = parseIntArg(args, 2);
        if ('error' in pVal) return { error: 'Usage: trump <trumpSuit> <partnerSuit> <partnerValue>' };
        return { command: `trump:${trump}:${pSuit}:${pVal.value}` };
      }
      case 'ex':
      case 'exchange': {
        if (args.length < 3) return { error: 'Usage: ex <i1> <i2> <i3>' };
        const i1 = parseIntArg(args, 0);
        const i2 = parseIntArg(args, 1);
        const i3 = parseIntArg(args, 2);
        if ('error' in i1 || 'error' in i2 || 'error' in i3) {
          return { error: 'Usage: ex <i1> <i2> <i3>' };
        }
        return { command: `exchange:${i1.value}:${i2.value}:${i3.value}` };
      }
      case 'jl':
      case 'jokerlead': {
        if (args.length < 2) return { error: 'Usage: jl <cardIdx> <suit>' };
        const idx = parseIntArg(args, 0);
        if ('error' in idx) return { error: 'Usage: jl <cardIdx> <suit>' };
        const suit = SUIT_MAP[args[1].toLowerCase()];
        if (suit === undefined || suit < 1) {
          return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
        }
        return { command: `jl:${idx.value}:${suit}` };
      }
      case 'log':
        return { command: 'log' };
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };

  const cmd = result.command;
  if (cmd === 'pass') return { args: ['bid', 0, false] };
  if (cmd === 'log') return { args: ['log'] };
  if (cmd.startsWith('bid:')) {
    const parts = cmd.split(':');
    return { args: ['bid', Number(parts[1]), parts[2] === '1'] };
  }
  if (cmd.startsWith('trump:')) {
    const parts = cmd.split(':');
    return {
      args: ['trump', undefined, undefined, undefined, Number(parts[1]), Number(parts[2]), Number(parts[3])],
    };
  }
  if (cmd.startsWith('exchange:')) {
    const parts = cmd.split(':');
    return {
      args: [
        'exchange',
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        undefined,
        [Number(parts[1]), Number(parts[2]), Number(parts[3])],
      ],
    };
  }
  if (cmd.startsWith('jl:')) {
    const parts = cmd.split(':');
    return {
      args: [
        'jokerlead',
        undefined,
        undefined,
        Number(parts[1]),
        undefined,
        undefined,
        undefined,
        undefined,
        Number(parts[2]),
      ],
    };
  }
  if (result.command === 'play') {
    return { args: ['play', undefined, undefined, result.cardIndex] };
  }
  return { args: [cmd as MightyArgs[0]] };
}

/** Help text for Mighty CLI mode. */
export const MIGHTY_HELP: string[] = [
  'bid <n> [nt]      - Place bid (0 to pass; add "nt" for No-Trump)',
  'pass              - Pass on bidding',
  'trump <t> <p> <v> - Declare trump suit, partner suit and partner card value',
  'ex <i1> <i2> <i3> - Discard 3 cards from hand (after kitty merge)',
  'jl <idx> <suit>   - Lead the Joker with a demanded suit',
  'log               - Show action log',
  ...TRICK_HELP,
];
