import type { courtPieceApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by courtPieceApi.exec. */
export type CourtPieceCliArgs = Parameters<typeof courtPieceApi.exec>;

const VALID_COMMANDS = [
  't',
  'trump',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'sd',
  'setdifficulty',
  'sl',
  'setlimit',
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
 * Parses a single CLI command line for the Court Piece (Rang) game into
 * {@link courtPieceApi}.exec arguments.
 *
 * Court Piece has a trump-declaration phase (`trump <1-4>` where 1=♠ 2=♣ 3=♥
 * 4=♦) followed by 13 tricks of trick play (`play <idx>`). `sd <0-2>` and
 * `sl <n>` reset the game with a new CPU difficulty / point limit because
 * config is only accepted on the `reset` command.
 */
export function parseCourtPieceCommand(input: string): CliParseResult<CourtPieceCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'trump': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit) || suit < 1 || suit > 4) return { error: 'Usage: trump <1-4> (1=♠ 2=♣ 3=♥ 4=♦)' };
      return { args: ['trump', { trumpSuit: suit }] };
    }
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
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'sl':
    case 'setlimit': {
      const limit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(limit) || limit < 1) return { error: 'Usage: sl <n> (point limit, >= 1)' };
      return { args: ['reset', { config: { pointLimit: limit } }] };
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

/** Help text shown in the CLI terminal for Court Piece (Rang). */
export const COURT_PIECE_HELP: string[] = [
  'trump <1-4>         - Declare trump (1=♠ 2=♣ 3=♥ 4=♦)',
  'p <idx>             - Play a card (Play phase, must follow suit)',
  'n / next            - Next trick',
  'nr / nextround      - Next round',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'sl <n>              - Set point limit (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
