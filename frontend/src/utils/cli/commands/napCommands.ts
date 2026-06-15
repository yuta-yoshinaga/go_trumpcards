import type { napApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by napApi.exec. */
export type NapCliArgs = Parameters<typeof napApi.exec>;

/** Valid Nap bid values (0=Pass, 2/3/4=trick count, 5=Nap). */
const VALID_BIDS = new Set([0, 2, 3, 4, 5]);

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
 * Parses a single CLI command line for the Nap (Napoleon) game into
 * {@link napApi}.exec arguments.
 *
 * Nap has a bidding phase (`bid <0/2/3/4/5>` / `pass`) followed by 5-trick play
 * (`play <idx>`). `sd <0-2>` resets the game with a new CPU difficulty because
 * config is only accepted on the `reset` command.
 */
export function parseNapCommand(input: string): CliParseResult<NapCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(bid) || !VALID_BIDS.has(bid))
        return { error: 'Usage: bid <0/2/3/4/5> (0=Pass 2=Two 3=Three 4=Four 5=Nap)' };
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

/** Help text shown in the CLI terminal for Nap. */
export const NAP_HELP: string[] = [
  'bid <0/2/3/4/5>  - Bid (0=Pass 2=Two 3=Three 4=Four 5=Nap)',
  'pass             - Pass (bid 0)',
  'p <idx>          - Play a card (Play phase, must follow suit)',
  'n / next         - Next trick',
  'nr / nextround   - Next round',
  'sd <0-2>         - Set CPU difficulty (resets game)',
  'h / hint         - Show hint',
  'l / log          - Show action log',
  'r / reset        - Reset game',
];
