import type { ganjifaApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by ganjifaApi.exec. */
export type GanjifaCliArgs = Parameters<typeof ganjifaApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'sd',
  'h',
  'hint',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parses a single CLI command line for the Ganjifa game into
 * {@link ganjifaApi}.exec arguments.
 *
 * Ganjifa has no bidding phase, so trick play (`play <idx>`) is the only
 * decision command. `sd <0-2>` resets the game with a new CPU difficulty
 * because config is only accepted on the `reset` command.
 */
export function parseGanjifaCommand(input: string): CliParseResult<GanjifaCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: idx }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sd': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
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

/** Help text shown in the CLI terminal for Ganjifa. */
export const GANJIFA_HELP: string[] = [
  'p <idx>          - Play a card (must follow the led suit)',
  'n / next         - Next trick',
  'nr / nextround   - Next round',
  'sd <0-2>         - Set CPU difficulty (resets game)',
  'h / hint         - Show hint',
  'l / log          - Show action log',
  'r / reset        - Reset game',
];
