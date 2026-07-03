import type { gutsApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by gutsApi.exec. */
export type GutsCliArgs = Parameters<typeof gutsApi.exec>;

const VALID_COMMANDS = [
  'i',
  'in',
  'o',
  'out',
  'n',
  'nr',
  'next',
  'nextround',
  'sp',
  'setplayers',
  'sa',
  'setante',
  'sc',
  'setchips',
  'st',
  'setrounds',
  'hint',
  'h',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Guts game into
 * {@link gutsApi}.exec arguments.
 *
 * On the human's Declare turn: `in` (or `i`) stays in the round and `out` (or
 * `o`) folds — both resolve the round via `declare`. `next` deals the next
 * round. `sp <2-7>`, `sa <n>`, `sc <n>`, and `st <n>` reset the game with a new
 * player count / ante / starting chips / target rounds because config is only
 * accepted on reset.
 */
export function parseGutsCommand(input: string): CliParseResult<GutsCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'i':
    case 'in':
      return { args: ['declare', 1] };
    case 'o':
    case 'out':
      return { args: ['declare', 0] };
    case 'n':
    case 'nr':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sp':
    case 'setplayers': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 2 || n > 7) return { error: 'Usage: sp <2-7> (player count)' };
      return { args: ['reset', undefined, { playerCount: n }] };
    }
    case 'sa':
    case 'setante': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: sa <n> (ante, >= 1)' };
      return { args: ['reset', undefined, { ante: n }] };
    }
    case 'sc':
    case 'setchips': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 10) return { error: 'Usage: sc <n> (starting chips, >= 10)' };
      return { args: ['reset', undefined, { startingChips: n }] };
    }
    case 'st':
    case 'setrounds': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: st <n> (target rounds, >= 1)' };
      return { args: ['reset', undefined, { targetRounds: n }] };
    }
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text shown in the CLI terminal for Guts. */
export const GUTS_HELP: string[] = [
  'i / in              - Declare in (stay)',
  'o / out             - Declare out (fold)',
  'n / next            - Next round',
  'sp <2-7>            - Set player count (resets game)',
  'sa <n>              - Set ante (resets game)',
  'sc <n>              - Set starting chips (resets game)',
  'st <n>              - Set target rounds (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
