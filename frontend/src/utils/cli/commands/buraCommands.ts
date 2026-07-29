import type { buraApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BuraArgs = Parameters<typeof buraApi.exec>;

const VALID_COMMANDS = ['p', 'play', 'c', 'claim', 'd', 'declare', 'log', 'r', 'reset', 'help', '?'];

/**
 * Parse a Bura CLI command into API exec arguments.
 *
 * `play` takes one to three indices, because a Bura lead may be several cards
 * of one suit. Duplicates are rejected here rather than passed through: a
 * repeated index would read as a longer play than the hand can support.
 */
export function parseBuraCommand(input: string): CliParseResult<BuraArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'play': {
      if (args.length === 0) return { error: 'Usage: p <i> [i] [i]' };
      if (args.length > 3) return { error: 'A play is at most 3 cards' };
      const indices: number[] = [];
      for (const a of args) {
        const n = Number(a);
        if (!Number.isInteger(n) || n < 0) return { error: `Invalid card index: ${a}` };
        if (indices.includes(n)) return { error: `Duplicate card index: ${a}` };
        indices.push(n);
      }
      return { args: ['play', indices] };
    }
    case 'c':
    case 'claim':
      return { args: ['claim'] };
    case 'd':
    case 'declare':
      return { args: ['declare'] };
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

/** Help text for Bura CLI mode. */
export const BURA_HELP: string[] = [
  'p <i> [i] [i]   - Play cards (a lead may be up to 3 of one suit)',
  'c/claim         - Claim 31 points (claiming short forfeits)',
  'd/declare       - Declare a combination (three trumps, etc.)',
  'log             - Show action log',
  'r/reset         - New round',
];
