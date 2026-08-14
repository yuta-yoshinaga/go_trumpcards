import type { crazyfourpokerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CrazyFourPokerArgs = Parameters<typeof crazyfourpokerApi.exec>;

const VALID_COMMANDS = ['bet', 'b', 'play', 'p', 'fold', 'f', 'next', 'hint', 'h', 'log', 'r', 'reset', 'help', '?'];

const BET_USAGE = 'Usage: bet <ante> [queensUp]';
const PLAY_USAGE = 'Usage: play <multiplier>';

/** CLI help text for Crazy 4 Poker. */
export const CRAZYFOURPOKER_CLI_HELP = [
  'bet <ante> [qu]  place the ante (an equal Super Bonus comes with it)',
  'play <mult>      place the play bet; 3x needs a pair of aces or better',
  'fold (f)         fold, losing the ante and the Super Bonus',
  'next             deal the next hand',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Crazy 4 Poker CLI command into API exec arguments. */
export function parseCrazyFourPokerCommand(input: string): CliParseResult<CrazyFourPokerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const ante = parseIntArg(args, 0);
      if ('error' in ante) return { error: BET_USAGE };
      // **Queens Up は省略可。** 省略は「置かない」。
      if (args.length < 2) return { args: ['bet', { ante: ante.value, queensUp: 0 }] };
      const qu = parseIntArg(args.slice(1), 0);
      if ('error' in qu) return { error: BET_USAGE };
      return { args: ['bet', { ante: ante.value, queensUp: qu.value }] };
    }
    case 'p':
    case 'play': {
      // **上限は書かない。** 手役次第なのでサーバが弾く。
      const mult = parseIntArg(args, 0);
      if ('error' in mult) return { error: PLAY_USAGE };
      return { args: ['play', { multiplier: mult.value }] };
    }
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'next':
      return { args: ['next'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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
