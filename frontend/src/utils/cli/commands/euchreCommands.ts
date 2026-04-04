import type { euchreApi } from '../../../api/gameApi';
import { parseIntArg } from '../commandParserBase';
import type { CliParseResult } from '../types';
import { parseTrickCommand, TRICK_HELP } from './sharedTrickCommands';

type EuchreArgs = Parameters<typeof euchreApi.exec>;

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

const EXTRA_COMMANDS = ['ou', 'orderup', 'pass', 'ct', 'calltrump', 'dis', 'discard', 'alone'];

/** Parse a Euchre CLI command into API exec arguments. */
export function parseEuchreCommand(input: string): CliParseResult<EuchreArgs> {
  const result = parseTrickCommand(input, EXTRA_COMMANDS, (cmd, args) => {
    switch (cmd) {
      case 'ou':
      case 'orderup':
        return { command: 'orderup' };
      case 'pass':
        return { command: 'pass' };
      case 'ct':
      case 'calltrump': {
        if (args.length === 0) return { error: 'Usage: ct <suit>' };
        const suit = SUIT_MAP[args[0].toLowerCase()];
        if (suit === undefined) return { error: 'Invalid suit. Use: spade/clover/heart/diamond' };
        return { command: `calltrump:${suit}` };
      }
      case 'dis':
      case 'discard': {
        const parsed = parseIntArg(args, 0);
        if ('error' in parsed) return { error: 'Usage: dis <idx>' };
        return { command: `discard:${parsed.value}` };
      }
      case 'alone':
        return { command: 'alone' };
      default:
        return null;
    }
  });

  if ('error' in result) return { error: result.error };

  const cmd = result.command;
  if (cmd === 'orderup') return { args: ['orderup'] };
  if (cmd === 'pass') return { args: ['pass'] };
  if (cmd === 'alone') return { args: ['orderup', undefined, undefined, true] };
  if (cmd.startsWith('calltrump:')) {
    const suit = Number(cmd.split(':')[1]);
    return { args: ['calltrump', undefined, suit] };
  }
  if (cmd.startsWith('discard:')) {
    const idx = Number(cmd.split(':')[1]);
    return { args: ['discard', idx] };
  }
  return { args: [cmd as EuchreArgs[0], result.cardIndex] };
}

/** Help text for Euchre CLI mode. */
export const EUCHRE_HELP: string[] = [
  'ou/orderup  - Order up face card',
  'pass        - Pass on bid',
  'ct <suit>   - Call trump (spade/clover/heart/diamond)',
  'alone       - Order up going alone',
  'dis <idx>   - Discard a card',
  ...TRICK_HELP,
];
