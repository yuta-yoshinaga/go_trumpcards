import type { fourcardpokerApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type FourCardPokerArgs = Parameters<typeof fourcardpokerApi.exec>;

const VALID_COMMANDS = ['b', 'bet', 'p', 'play', 'f', 'fold', 'r', 'reset', 'log', 'l'];

/** Parse a Four Card Poker CLI command into API exec arguments. */
export function parseFourCardPokerCommand(input: string): CliParseResult<FourCardPokerArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'bet': {
      const ante = parseIntArg(args, 0);
      if ('error' in ante) return { error: 'Usage: b <ante> [acesUp]' };
      let acesUp = 0;
      if (args.length > 1) {
        const parsed = parseIntArg(args, 1);
        if ('error' in parsed) return { error: 'Usage: b <ante> [acesUp]' };
        acesUp = parsed.value;
      }
      return { args: ['bet', ante.value, acesUp] as FourCardPokerArgs };
    }
    case 'p':
    case 'play': {
      const mult = parseIntArg(args, 0);
      if ('error' in mult || mult.value < 1 || mult.value > 3) {
        return { error: 'Usage: p <1|2|3> (play bet multiplier)' };
      }
      return { args: ['play', undefined, undefined, mult.value] as FourCardPokerArgs };
    }
    case 'f':
    case 'fold':
      return { args: ['fold'] as FourCardPokerArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as FourCardPokerArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as FourCardPokerArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Four Card Poker CLI mode. */
export const FOURCARDPOKER_HELP: string[] = [
  'b <ante> [acesUp]  - Place the ante (and optional Aces Up side bet)',
  'p <1|2|3>          - Make the play bet at the given ante multiplier',
  'f/fold             - Fold and forfeit the ante',
  'log                - Show the action log',
  'r/reset            - Reset game',
];
