import type { primeroApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by primeroApi.exec. */
export type PrimeroCliArgs = Parameters<typeof primeroApi.exec>;

const VALID_COMMANDS = [
  'c',
  'call',
  'ra',
  'raise',
  'vie',
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

/**
 * Parses a single CLI command line for the Primero game into
 * {@link primeroApi}.exec arguments.
 *
 * On the human's Betting turn: `call` (or `c`) matches the current bet,
 * `raise` (`ra` / `vie`) increases it by a fixed increment, and `fold` (`f`)
 * drops out of the round. `next` deals the next round. `sp <2-6>`, `sa <n>`,
 * `sc <n>`, and `st <n>` reset the game with a new player count / ante /
 * starting chips / target rounds because config is only accepted on reset.
 */
export function parsePrimeroCommand(input: string): CliParseResult<PrimeroCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'call':
      return { args: ['bet', 'call'] };
    case 'ra':
    case 'raise':
    case 'vie':
      return { args: ['bet', 'raise'] };
    case 'f':
    case 'fold':
      return { args: ['bet', 'fold'] };
    case 'n':
    case 'nr':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sp':
    case 'setplayers': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 2 || n > 6) return { error: 'Usage: sp <2-6> (player count)' };
      return { args: ['reset', undefined, { playerCount: n }] };
    }
    case 'sa':
    case 'setante': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: sa <n> (ante, >= 1)' };
      return { args: ['reset', undefined, { ante: n }] };
    }
    case 'sc':
    case 'setchips': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 10) return { error: 'Usage: sc <n> (starting chips, >= 10)' };
      return { args: ['reset', undefined, { startingChips: n }] };
    }
    case 'st':
    case 'setrounds': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: st <n> (target rounds, >= 1)' };
      return { args: ['reset', undefined, { targetRounds: n }] };
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

/** Help text shown in the CLI terminal for Primero. */
export const PRIMERO_HELP: string[] = [
  'c / call            - Call (match the current bet)',
  'ra / raise / vie    - Raise (vie) by a fixed increment',
  'f / fold            - Fold (drop out of the round)',
  'n / next            - Next round',
  'sp <2-6>            - Set player count (resets game)',
  'sa <n>              - Set ante (resets game)',
  'sc <n>              - Set starting chips (resets game)',
  'st <n>              - Set target rounds (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
