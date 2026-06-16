import type { beziqueApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by beziqueApi.exec. */
export type BeziqueCliArgs = Parameters<typeof beziqueApi.exec>;

const VALID_COMMANDS = [
  'p',
  'play',
  'm',
  'meld',
  's',
  'skip',
  'n',
  'next',
  'sd',
  'setdifficulty',
  'st',
  'settarget',
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
 * Parses a single CLI command line for the Bezique game into
 * {@link beziqueApi}.exec arguments.
 *
 * Bezique alternates a Play phase (`play <idx>`) with a Meld phase where the
 * trick winner declares a meld (`meld <idx>`) or skips it (`skip`); `next`
 * advances to the following deal. `sd <0-2>` and `st <n>` reset the game with a
 * new CPU difficulty / target score because config is only accepted on reset.
 */
export function parseBeziqueCommand(input: string): CliParseResult<BeziqueCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: idx }] };
    }
    case 'm':
    case 'meld': {
      const idx = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(idx)) return { error: 'Usage: m <idx>' };
      return { args: ['meld', { meldIndex: idx }] };
    }
    case 's':
    case 'skip':
      return { args: ['skip'] };
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'sd':
    case 'setdifficulty': {
      const level = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(level) || level < 0 || level > 2) return { error: 'Usage: sd <0-2> (0=Easy 1=Normal 2=Hard)' };
      return { args: ['reset', { config: { cpuDifficulty: level } }] };
    }
    case 'st':
    case 'settarget': {
      const target = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(target) || target < 1) return { error: 'Usage: st <n> (target score, >= 1)' };
      return { args: ['reset', { config: { targetScore: target } }] };
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

/** Help text shown in the CLI terminal for Bezique. */
export const BEZIQUE_HELP: string[] = [
  'p <idx>             - Play a card (Play phase)',
  'm <idx>             - Declare a meld (Meld phase, trick winner only)',
  's / skip            - Skip declaring a meld',
  'n / next            - Next deal',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'st <n>              - Set target score (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
