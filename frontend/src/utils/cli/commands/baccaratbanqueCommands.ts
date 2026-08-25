import type { baccaratbanqueApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BaccaratBanqueArgs = Parameters<typeof baccaratbanqueApi.exec>;

const VALID_COMMANDS = [
  'draw',
  'd',
  'stand',
  's',
  'nextcoup',
  'nc',
  'retire',
  'hint',
  'h',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/** CLI help text for Baccarat Banque. */
export const BACCARATBANQUE_CLI_HELP = [
  'draw (d)         take a third card (free at any total)',
  'stand (s)        stand pat and settle both tableaux',
  'nextcoup (nc)    deal the next coup; the bank stays with you',
  'retire           give up the bank and keep the chips',
  'hint (h)         show a hint',
  'log (l)          show the action log',
  'reset (r)        restart',
];

/** Parse a Baccarat Banque CLI command into API exec arguments. */
export function parseBaccaratBanqueCommand(input: string): CliParseResult<BaccaratBanqueArgs> {
  const { cmd } = splitCommand(input);

  switch (cmd) {
    // **引くと止まるは別の命令。** 真偽値ひとつに畳むと、付け忘れた要求が
    // 黙ってどちらかに倒れる。
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 's':
    case 'stand':
      return { args: ['stand'] };
    case 'nc':
    case 'nextcoup':
      return { args: ['nextcoup'] };
    case 'retire':
      return { args: ['retire'] };
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
