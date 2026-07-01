import type { kingApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KingArgs = Parameters<typeof kingApi.exec>;

const VALID_COMMANDS = ['c', 'contract', 'p', 'play', 'n', 'next', 'h', 'hint', 'r', 'reset', 'help', '?'];

/**
 * Parse a King CLI command into API exec arguments.
 *
 * `contract <idx> [trump]` selects the deal's contract (0..6). For contract 6
 * ("King (Trump)") a trump suit (1..4) may follow; otherwise trumpSuit defaults
 * to -1.
 */
export function parseKingCommand(input: string): CliParseResult<KingArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'contract': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: c <contract 0-6> [trump 1-4]' };
      const trump = parseIntArg(args, 1);
      const trumpSuit = 'error' in trump ? -1 : trump.value;
      return { args: ['contract', { contract: parsed.value, trumpSuit }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { handIndex: parsed.value }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for King CLI mode. */
export const KING_HELP: string[] = [
  'c <n> [t]        - Choose the deal contract 0-6 (t=trump 1-4 for contract 6)',
  'p <idx>          - Play a card (Play phase, must follow suit)',
  'n/next           - Next deal',
  'h/hint           - Show hint',
  'r/reset          - Reset game',
];
