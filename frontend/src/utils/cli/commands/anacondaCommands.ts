import type { anacondaApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by anacondaApi.exec. */
export type AnacondaCliArgs = Parameters<typeof anacondaApi.exec>;

const VALID_COMMANDS = [
  'p',
  'pass',
  'k',
  'keep',
  'c',
  'call',
  'check',
  'ra',
  'raise',
  'f',
  'fold',
  'n',
  'nr',
  'next',
  'nextround',
  'sp',
  'setplayers',
  'sa',
  'setante',
  'sc',
  'setchips',
  'st',
  'setrounds',
  'hint',
  'h',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/** Parses a list of card-index arguments into unique non-negative integers. */
function parseIndices(args: string[]): number[] | null {
  if (args.length === 0) return null;
  const indices: number[] = [];
  for (const raw of args) {
    const n = Number.parseInt(raw, 10);
    if (Number.isNaN(n) || n < 0) return null;
    indices.push(n);
  }
  return indices;
}

/**
 * Parses a single CLI command line for the Anaconda game into
 * {@link anacondaApi}.exec arguments.
 *
 * During the Pass phase `p <i...>` passes the selected cards left; during the
 * Set phase `k <i...>` keeps exactly 5 cards; during the Roll phase `c`/`call`
 * calls (or checks), `ra`/`raise` raises, and `f`/`fold` folds. `n` deals the
 * next round. `sp <3-7>`, `sa <n>`, `sc <n>`, and `st <n>` reset the game with
 * a new player count / ante / starting chips / target rounds because config is
 * only accepted on reset.
 */
export function parseAnacondaCommand(input: string): CliParseResult<AnacondaCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'p':
    case 'pass': {
      const indices = parseIndices(args);
      if (!indices || indices.length < 1 || indices.length > 3) {
        return { error: 'Usage: p <i...> (pass 1-3 card indices)' };
      }
      return { args: ['pass', indices] };
    }
    case 'k':
    case 'keep': {
      const indices = parseIndices(args);
      if (indices?.length !== 5) {
        return { error: 'Usage: k <i...> (keep exactly 5 card indices)' };
      }
      return { args: ['keep', indices] };
    }
    case 'c':
    case 'call':
    case 'check':
      return { args: ['bet', undefined, 'call'] };
    case 'ra':
    case 'raise':
      return { args: ['bet', undefined, 'raise'] };
    case 'f':
    case 'fold':
      return { args: ['bet', undefined, 'fold'] };
    case 'n':
    case 'nr':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sp':
    case 'setplayers': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 3 || n > 7) return { error: 'Usage: sp <3-7> (player count)' };
      return { args: ['reset', undefined, undefined, { playerCount: n }] };
    }
    case 'sa':
    case 'setante': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: sa <n> (ante, >= 1)' };
      return { args: ['reset', undefined, undefined, { ante: n }] };
    }
    case 'sc':
    case 'setchips': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 10) return { error: 'Usage: sc <n> (starting chips, >= 10)' };
      return { args: ['reset', undefined, undefined, { startingChips: n }] };
    }
    case 'st':
    case 'setrounds': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: st <n> (target rounds, >= 1)' };
      return { args: ['reset', undefined, undefined, { targetRounds: n }] };
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

/** Help text shown in the CLI terminal for Anaconda. */
export const ANACONDA_HELP: string[] = [
  'p <i...>            - Pass selected cards left (3->2->1)',
  'k <i...>            - Keep the best 5 cards (discard 2)',
  'c / call            - Call / check',
  'ra / raise          - Raise',
  'f / fold            - Fold',
  'n / next            - Next round',
  'sp <3-7>            - Set player count (resets game)',
  'sa <n>              - Set ante (resets game)',
  'sc <n>              - Set starting chips (resets game)',
  'st <n>              - Set target rounds (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
