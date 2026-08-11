import type { estimationApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type EstimationArgs = Parameters<typeof estimationApi.exec>;

const VALID_COMMANDS = [
  't',
  'trump',
  'b',
  'bid',
  'p',
  'play',
  'n',
  'next',
  'h',
  'hint',
  'g',
  'giveup',
  'r',
  'reset',
  'log',
  'l',
];

/** Suit codes accepted by `trump` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Calls run from a Dash Call (0) up to the whole hand. */
const BID_MIN = 0;
const BID_MAX = 13;

/** Parse an Estimation CLI command into API exec arguments (indices are 0-based). */
export function parseEstimationCommand(input: string): CliParseResult<EstimationArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 't':
    case 'trump': {
      const suit = parseIntArg(args, 0);
      if ('error' in suit) return { error: 'Usage: t <suit 1-4>' };
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: t <suit 1-4>' };
      return { args: ['trump', undefined, undefined, suit.value] as EstimationArgs };
    }
    case 'b':
    case 'bid': {
      const bid = parseIntArg(args, 0);
      if ('error' in bid) return { error: 'Usage: b <0-13>' };
      // **0 は Dash Call という宣言。** 省略とは区別する。
      if (bid.value < BID_MIN || bid.value > BID_MAX) return { error: 'Usage: b <0-13>' };
      return { args: ['bid', undefined, undefined, undefined, bid.value] as EstimationArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as EstimationArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as EstimationArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as EstimationArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as EstimationArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as EstimationArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as EstimationArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Estimation CLI mode. */
export const ESTIMATION_HELP: string[] = [
  't <suit>     - Choose the trump suit (1=S 2=C 3=H 4=D; dealer only)',
  'b <0-13>     - Call your tricks (0 is a Dash Call)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
