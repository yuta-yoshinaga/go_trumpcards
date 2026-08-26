import type { sevenTwentySevenApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by sevenTwentySevenApi.exec. */
export type SevenTwentySevenCliArgs = Parameters<typeof sevenTwentySevenApi.exec>;

const VALID_COMMANDS = [
  'c',
  'card',
  's',
  'stand',
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
 * Parses a single CLI command line for the SevenTwentySeven game into
 * {@link sevenTwentySevenApi}.exec arguments.
 *
 * Each pass: `card` (or `c`) takes one more card and `stand` (or `s`) ends the
 * human's drawing for the round. `next` deals the next round. `sp <2-7>`,
 * `sa <n>`, `sc <n>`, and `st <n>` reset the game with a new player count /
 * ante / starting chips / target rounds because config is only accepted on
 * reset.
 */
export function parseSevenTwentySevenCommand(input: string): CliParseResult<SevenTwentySevenCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'card':
      return { args: ['card'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'n':
    case 'nr':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sp':
    case 'setplayers': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 2 || n > 7) return { error: 'Usage: sp <2-7> (player count)' };
      return { args: ['reset', { playerCount: n }] };
    }
    case 'sa':
    case 'setante': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: sa <n> (ante, >= 1)' };
      return { args: ['reset', { ante: n }] };
    }
    case 'sc':
    case 'setchips': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 10) return { error: 'Usage: sc <n> (starting chips, >= 10)' };
      return { args: ['reset', { startingChips: n }] };
    }
    case 'st':
    case 'setrounds': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 1) return { error: 'Usage: st <n> (target rounds, >= 1)' };
      return { args: ['reset', { targetRounds: n }] };
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

/** Help text shown in the CLI terminal for SevenTwentySeven. */
export const SEVENTWENTYSEVEN_HELP: string[] = [
  'c / card            - Take one more card',
  's / stand           - Stand pat (no more cards this round)',
  'n / next            - Next round',
  'sp <2-7>            - Set player count (resets game)',
  'sa <n>              - Set ante (resets game)',
  'sc <n>              - Set starting chips (resets game)',
  'st <n>              - Set target rounds (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
