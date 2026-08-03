import type { minchiateApi } from '../../../api/gameApi';
import { MINCHIATE_SURPLUS } from '../../../types/card';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by minchiateApi.exec. */
export type MinchiateCliArgs = Parameters<typeof minchiateApi.exec>;

const VALID_COMMANDS = [
  'scarto',
  'discard',
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
 * Parses a single CLI command line for Minchiate into
 * {@link minchiateApi}.exec arguments.
 *
 * Minchiate has no bidding phase: the dealer buries the surplus (`scarto`)
 * and then it is trick play (`play <idx>`). `sd <0-2>` resets the game with a
 * new CPU difficulty because config is only accepted on `reset`.
 */
export function parseMinchiateCommand(input: string): CliParseResult<MinchiateCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'scarto':
    case 'discard': {
      // **枚数は定数から出す。**メッセージに数字を直接書くと、余剰枚数が変わった
      // ときに案内だけ古くなる。
      if (args.length < MINCHIATE_SURPLUS) {
        return { error: `Usage: scarto <i0> ... (${MINCHIATE_SURPLUS} card indices)` };
      }
      const indices = args.slice(0, MINCHIATE_SURPLUS).map((a) => Number.parseInt(a, 10));
      if (indices.some(Number.isNaN)) return { error: 'Usage: scarto <i0> ... (numeric indices)' };
      return { args: ['scarto', { cardIndices: indices }] };
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

/** Help text shown in the CLI terminal for Minchiate. */
export const MINCHIATE_HELP: string[] = [
  'scarto <i0> <i1> - Bury the 2 surplus cards (dealer only; no trumps or Matto)',
  'p <idx>          - Play a card (must follow, or trump when void)',
  'n / next         - Next trick',
  'nr / nextround   - Next round',
  'sd <0-2>         - Set CPU difficulty (resets game)',
  'h / hint         - Show hint',
  'l / log          - Show action log',
  'r / reset        - Reset game',
];
