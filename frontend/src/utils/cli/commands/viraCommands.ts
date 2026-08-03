import type { viraApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by viraApi.exec. */
export type ViraCliArgs = Parameters<typeof viraApi.exec>;

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
 * Parses a single CLI command line for the Vira game into
 * {@link viraApi}.exec arguments.
 *
 * Vira has a bidding phase (`bid <0-4>` / `pass`) followed by trick play
 * (`play <idx>`). `sd <0-2>` resets the game with a new CPU difficulty because
 * config is only accepted on the `reset` command.
 */
export function parseViraCommand(input: string): CliParseResult<ViraCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bid': {
      const bid = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(bid) || bid < 0 || bid > 4)
        return { error: 'Usage: bid <0-4> (0=Pass 1=Six 2=Misère 3=Seven 4=Eight)' };
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

/** Help text shown in the CLI terminal for Vira. */
export const PREFERENCE_HELP: string[] = [
  'bid <0-4>        - Bid (0=Pass 1=Six 2=Misère 3=Seven 4=Eight)',
  'pass             - Pass (bid 0)',
  'p <idx>          - Play a card (Play phase, must follow suit)',
  'n / next         - Next trick',
  'nr / nextround   - Next round',
  'sd <0-2>         - Set CPU difficulty (resets game)',
  'h / hint         - Show hint',
  'l / log          - Show action log',
  'r / reset        - Reset game',
];
