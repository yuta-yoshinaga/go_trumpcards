import type { cuckooApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by cuckooApi.exec. */
export type CuckooCliArgs = Parameters<typeof cuckooApi.exec>;

const VALID_COMMANDS = [
  'k',
  'keep',
  's',
  'swap',
  'rf',
  'refuse',
  'ac',
  'accept',
  'nr',
  'nextround',
  'sd',
  'setdifficulty',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Cuckoo game into
 * {@link cuckooApi}.exec arguments.
 *
 * `keep` keeps your card; `swap` swaps with the next player (the dealer swaps
 * with the stock). When you hold a King and someone tries to swap into you,
 * `refuse` reveals the King to block it and `accept` allows the swap. `nextround`
 * advances after the lowest card loses a life. `sd <0-2>` resets the game with a
 * new CPU difficulty because config is only accepted on reset.
 */
export function parseCuckooCommand(input: string): CliParseResult<CuckooCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'k':
    case 'keep':
      return { args: ['keep'] };
    case 's':
    case 'swap':
      return { args: ['swap'] };
    case 'rf':
    case 'refuse':
      return { args: ['refuse'] };
    case 'ac':
    case 'accept':
      return { args: ['accept'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
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

/** Help text shown in the CLI terminal for Cuckoo. */
export const CUCKOO_HELP: string[] = [
  'k / keep            - Keep your card',
  's / swap            - Swap with your neighbour (dealer swaps with the stock)',
  'rf / refuse         - Reveal your King to refuse an incoming swap',
  'ac / accept         - Accept an incoming swap',
  'nr / nextround      - Advance to the next round',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
