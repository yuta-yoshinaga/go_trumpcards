import type { bridgeApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type BridgeArgs = Parameters<typeof bridgeApi.exec>;

const SUIT_MAP: Record<string, number> = {
  clubs: 1,
  c: 1,
  diamonds: 2,
  d: 2,
  hearts: 3,
  h: 3,
  spades: 4,
  s: 4,
  notrump: 5,
  nt: 5,
};

const EXTRA_COMMANDS = ['bid', 'pass', 'double', 'dbl', 'redouble', 'rdbl', 'log'];

/** Parse a Bridge CLI command into API exec arguments. */
export function parseBridgeCommand(input: string): CliParseResult<BridgeArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'pass':
        return { command: 'pass' };
      case 'double':
      case 'dbl':
        return { command: 'double' };
      case 'redouble':
      case 'rdbl':
        return { command: 'redouble' };
      case 'log':
        return { command: 'log' };
      case 'bid': {
        if (args.length < 2) return { error: 'Usage: bid <level> <suit> (suit: clubs/diamonds/hearts/spades/notrump)' };
        const level = parseIntArg(args, 0);
        if ('error' in level) return { error: 'Usage: bid <level> <suit>' };
        const suit = SUIT_MAP[args[1].toLowerCase()];
        if (suit === undefined) return { error: 'Invalid suit. Use: clubs/diamonds/hearts/spades/notrump' };
        return { command: `bid:${level.value}:${suit}` };
      }
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };

  const cmd = result.command;
  if (cmd === 'pass') return { args: ['bid', undefined, 0] };
  if (cmd === 'double') return { args: ['bid', undefined, 1] };
  if (cmd === 'redouble') return { args: ['bid', undefined, 2] };
  if (cmd === 'log') return { args: ['log'] };
  if (cmd.startsWith('bid:')) {
    const parts = cmd.split(':');
    return { args: ['bid', undefined, 3, Number(parts[1]), Number(parts[2])] };
  }
  return { args: [cmd as BridgeArgs[0], result.cardIndex] };
}

/** Help text for Bridge CLI mode. */
export const BRIDGE_HELP: string[] = [
  'bid <lvl> <suit> - Bid (clubs/diamonds/hearts/spades/notrump)',
  'pass        - Pass on bidding',
  'double/dbl  - Double',
  'redouble/rdbl- Redouble',
  'log         - Show action log',
  ...TRICK_HELP,
];
