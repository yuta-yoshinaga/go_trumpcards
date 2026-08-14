import type { minibridgeApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type MinibridgeArgs = Parameters<typeof minibridgeApi.exec>;

// **競りは無い。** デクレアラーは公開申告した HCP から決まる。
const VALID_COMMANDS = [
  'c',
  'contract',
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

/** Contract levels accepted by `contract`. Level n needs 6 + n tricks. */
const LEVEL_MIN = 1;
const LEVEL_MAX = 7;

/** Denominations accepted by `contract`. **0 is no-trump, not "unset".** */
const SUIT_MIN = 0;
const SUIT_MAX = 4;

/** Parse a Minibridge CLI command into API exec arguments (indices are 0-based). */
export function parseMinibridgeCommand(input: string): CliParseResult<MinibridgeArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'contract': {
      const level = parseIntArg(args, 0);
      if ('error' in level) return { error: 'Usage: c <level 1-7> <suit 0-4>' };
      if (level.value < LEVEL_MIN || level.value > LEVEL_MAX) return { error: 'Usage: c <level 1-7> <suit 0-4>' };
      // **スートを既定値で埋めない。** 0 はノートランプという明示の選択肢。
      const suit = parseIntArg(args, 1);
      if ('error' in suit) return { error: 'Usage: c <level 1-7> <suit 0-4>' };
      if (suit.value < SUIT_MIN || suit.value > SUIT_MAX) return { error: 'Usage: c <level 1-7> <suit 0-4>' };
      return { args: ['contract', undefined, undefined, level.value, suit.value] as MinibridgeArgs };
    }
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIdx>' };
      return { args: ['play', idx.value] as MinibridgeArgs };
    }
    case 'n':
    case 'next':
      return { args: ['next'] as MinibridgeArgs };
    case 'h':
    case 'hint':
      return { args: ['hint'] as MinibridgeArgs };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] as MinibridgeArgs };
    case 'log':
    case 'l':
      return { args: ['log'] as MinibridgeArgs };
    case 'r':
    case 'reset':
      return { args: ['reset'] as MinibridgeArgs };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

/** Help text for Minibridge CLI mode. */
export const MINIBRIDGE_HELP: string[] = [
  'c <lvl> <s>  - Choose the contract (level 1-7; suit 0=NT 1=S 2=C 3=H 4=D)',
  'p <cardIdx>  - Play a card (the dummy is yours to play too)',
  'n/next       - Start the next deal',
  'h/hint       - Show a hint',
  'g/giveup     - Give up',
  'log          - Show the action log',
  'r/reset      - Reset game',
];
