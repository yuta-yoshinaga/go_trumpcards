import type { killeApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by killeApi.exec. */
export type KilleCliArgs = Parameters<typeof killeApi.exec>;

const VALID_COMMANDS = [
  'e',
  'exchange',
  's',
  'satisfied',
  're',
  'reenter',
  'nr',
  'nextround',
  'st',
  'setstake',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Kille game into {@link killeApi}.exec
 * arguments.
 *
 * `exchange` challenges your left neighbour — the dealer swaps with the stock
 * instead, and nobody may refuse. `satisfied` keeps your card. `reenter` buys
 * you back in after you go out (three times at most, at a rising price), and
 * `nextround` deals again. `st <1-100>` resets the game with a new stake,
 * because config is only accepted on reset.
 */
export function parseKilleCommand(input: string): CliParseResult<KilleCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'e':
    case 'exchange':
      return { args: ['exchange'] };
    case 's':
    case 'satisfied':
      return { args: ['satisfied'] };
    case 're':
    case 'reenter':
      return { args: ['reenter'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'st':
    case 'setstake': {
      const stake = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(stake) || stake < 1 || stake > 100) return { error: 'Usage: st <1-100>' };
      return { args: ['reset', { config: { stake } }] };
    }
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text shown in the CLI terminal for Kille. */
export const KILLE_HELP: string[] = [
  'e / exchange        - Challenge your left neighbour (dealer swaps with the stock)',
  's / satisfied       - Keep your card',
  're / reenter        - Buy back in after going out',
  'nr / nextround      - Deal the next round',
  'st <1-100>          - Set the stake (resets game)',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
