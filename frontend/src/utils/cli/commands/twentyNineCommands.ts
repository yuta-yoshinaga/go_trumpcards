import type { twentyNineApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by twentyNineApi.exec. */
export type TwentyNineCliArgs = Parameters<typeof twentyNineApi.exec>;

/** Legal bid values for Twenty-Nine (0=Pass, 16, 20, 24, 28). */
const VALID_BIDS = new Set([0, 16, 20, 24, 28]);

const VALID_COMMANDS = [
  'b',
  'bid',
  'pass',
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
 * Parses a single CLI command line for the Twenty-Nine (29) game into
 * {@link twentyNineApi}.exec arguments.
 *
 * Twenty-Nine has a bidding phase (`bid <0/16/20/24/28>` / `pass`) followed by
 * eight tricks of trick play (`play <idx>`). `sd <0-2>` resets the game with a
 * new CPU difficulty because config is only accepted on the `reset` command.
 */
export function parseTwentyNineCommand(input: string): CliParseResult<TwentyNineCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(bid) || !VALID_BIDS.has(bid)) return { error: 'Usage: bid <0/16/20/24/28> (0=Pass)' };
      return { args: ['bid', { bid }] };
    }
    case 'pass':
      return { args: ['bid', { bid: 0 }] };
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

/** Help text shown in the CLI terminal for Twenty-Nine (29). */
export const TWENTY_NINE_HELP: string[] = [
  'bid <0/16/20/24/28> - Bid (0=Pass)',
  'pass                - Pass (bid 0)',
  'p <idx>             - Play a card (Play phase, must follow suit)',
  'n / next            - Next trick',
  'nr / nextround      - Next round',
  'sd <0-2>            - Set CPU difficulty (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
