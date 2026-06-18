import type { spoonsApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by spoonsApi.exec. */
export type SpoonsCliArgs = Parameters<typeof spoonsApi.exec>;

const VALID_COMMANDS = [
  'p',
  'pass',
  'g',
  'grab',
  'n',
  'next',
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
 * Parses a single CLI command line for the Spoons game into
 * {@link spoonsApi}.exec arguments.
 *
 * On the Pass phase `pass <i>` passes the card at index `i` (0-based) to the
 * next player. When the grab window is open `grab` races to grab a spoon.
 * `next` advances to the following round. `sd <0-2>` resets the game with a new
 * CPU difficulty because config is only accepted on reset.
 */
export function parseSpoonsCommand(input: string): CliParseResult<SpoonsCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'pass': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx) || idx < 0) return { error: 'Usage: p <i> (pass card at index i, >= 0)' };
      return { args: ['pass', { cardIndex: idx }] };
    }
    case 'g':
    case 'grab':
      return { args: ['grab'] };
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

/** Help text shown in the CLI terminal for Spoons. */
export const SPOONS_HELP: string[] = [
  'p / pass <i>        - Pass the card at index i to the next player',
  'g / grab            - Grab a spoon (when the grab window is open)',
  'n / next            - Next round',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
