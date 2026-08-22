import type { ristikontraApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by ristikontraApi.exec. */
export type RistikontraCliArgs = Parameters<typeof ristikontraApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'n', 'next', 'sd', 'setdifficulty', 'l', 'log', 'r', 'reset', 'help', '?'];

/**
 * Parses a single CLI command line for the Ristikontra game into
 * {@link ristikontraApi}.exec arguments.
 *
 * `play <h>` plays hand card `h` (1-based for the user, converted to a 0-based
 * `handIndex`) onto the pile; **matching the pile top's rank** captures it —
 * a Jack is an ordinary card here, unlike in the clone source (Pişti).
 * `next` starts the following game. `sd <0-2>` resets the game with a new CPU
 * difficulty, because config is only accepted on reset.
 *
 * **There is no `sp`/`setplayers`.** Ristikontra is always a fixed 2-vs-2
 * partnership, so a table of any other size cannot form teams; the backend
 * rejects it, and advertising the command here would leave a verb that parses
 * and then fails.
 */
export function parseRistikontraCommand(input: string): CliParseResult<RistikontraCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: p <hand#> (1-based)' };
      return { args: ['play', { handIndex: n - 1 }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text shown in the CLI terminal for Ristikontra. */
export const RISTIKONTRA_HELP: string[] = [
  'p <hand#>           - Play hand card # onto the pile (capture by matching rank or with a Jack)',
  'n / next            - Start the next game',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
