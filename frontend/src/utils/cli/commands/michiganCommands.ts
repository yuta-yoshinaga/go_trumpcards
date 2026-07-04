import type { michiganApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by michiganApi.exec. */
export type MichiganCliArgs = Parameters<typeof michiganApi.exec>;

const VALID_COMMANDS = [
  'b',
  'bet',
  'p',
  'play',
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
 * Parses a single CLI command line for the Michigan (Newmarket) game into
 * {@link michiganApi}.exec arguments.
 *
 * On the human's Bet turn: `bet <h> <c> <d> <s>` (or `b …`) distributes chips
 * across the four boodles (A♥, K♣, Q♦, J♠). During the Play phase: `play <n>`
 * (or `p <n>`) plays hand index n. `next` deals the next round. `sp <3-8>`,
 * `sa <n>`, `sc <n>`, and `st <n>` reset the game with a new player count /
 * ante / starting chips / target rounds because config is only accepted on
 * reset.
 */
export function parseMichiganCommand(input: string): CliParseResult<MichiganCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      if (args.length < 4) return { error: 'Usage: b <h> <c> <d> <s> (four boodle bets, e.g. b 2 2 2 2)' };
      const bets = args.slice(0, 4).map((a) => Number.parseInt(a, 10));
      if (bets.some((n) => Number.isNaN(n) || n < 0)) return { error: 'Bets must be four non-negative integers' };
      return { args: ['bet', bets] };
    }
    case 'p':
    case 'play': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 0) return { error: 'Usage: p <n> (hand index to play)' };
      return { args: ['play', undefined, n] };
    }
    case 'n':
    case 'nr':
    case 'next':
    case 'nextround':
      return { args: ['nextround'] };
    case 'sp':
    case 'setplayers': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 3 || n > 8) return { error: 'Usage: sp <3-8> (player count)' };
      return { args: ['reset', undefined, undefined, { playerCount: n }] };
    }
    case 'sa':
    case 'setante': {
      const n = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(n) || n < 4) return { error: 'Usage: sa <n> (ante, >= 4)' };
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

/** Help text shown in the CLI terminal for Michigan. */
export const MICHIGAN_HELP: string[] = [
  'b <h> <c> <d> <s>   - Bet on the boodles (A♥ K♣ Q♦ J♠, total = ante)',
  'p <n>               - Play hand index n',
  'n / next            - Next round',
  'sp <3-8>            - Set player count (resets game)',
  'sa <n>              - Set ante (resets game)',
  'sc <n>              - Set starting chips (resets game)',
  'st <n>              - Set target rounds (resets game)',
  'h / hint            - Show hint',
  'l / log             - Show action log',
  'r / reset           - Reset game',
];
