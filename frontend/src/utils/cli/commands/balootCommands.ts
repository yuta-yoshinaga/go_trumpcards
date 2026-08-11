import type { balootApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type BalootArgs = Parameters<typeof balootApi.exec>;

const VALID_COMMANDS = [
  'sun',
  'hokom',
  'pass',
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

/** Suit codes accepted by `hokom` (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_MIN = 1;
const SUIT_MAX = 4;

/** Parse a Baloot CLI command into API exec arguments (indices are 0-based). */
export function parseBalootCommand(input: string): CliParseResult<BalootArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'sun':
      return { args: ['sun'] as BalootArgs };
    case 'hokom': {
      const suit = parseIntArg(args, 0);
      if ('error' in suit) return { error: 'Usage: hokom <suit 1-4>' };
      // **スート無しの Hokom を通さない。** 既定値で埋めると、選んでいない
      // スートが切り札になる。
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: hokom <suit 1-4>' };
      return { args: ['hokom', undefined, undefined, suit.value] as BalootArgs };
    }
    case 'pass':
      return { args: ['pass'] as BalootArgs };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as BalootArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as BalootArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as BalootArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as BalootArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as BalootArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as BalootArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Baloot CLI mode. */
export const BALOOT_HELP: string[] = [
  'sun          - Declare Sun (no trump)',
  'hokom <suit> - Declare Hokom (trump; 1=S 2=C 3=H 4=D)',
  'pass         - Pass on declaring (the dealer cannot)',
  'p <cardIdx>  - Play a card (marked * in your hand)',
  'n/next       - Start the next round',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
