import type { wattenApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type WattenArgs = Parameters<typeof wattenApi.exec>;

const VALID_COMMANDS = [
  'd',
  'declare',
  'p',
  'play',
  'rz',
  'raise',
  'hold',
  'fold',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/** Maps a Schlag-rank letter/number to its numeric card value (1=A, 7..10, 11=J, 12=Q, 13=K). */
const RANK_CODES: Readonly<Record<string, number>> = {
  a: 1,
  '7': 7,
  '8': 8,
  '9': 9,
  '10': 10,
  t: 10,
  j: 11,
  q: 12,
  k: 13,
};

/** Maps a critical-suit letter to its numeric code (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_CODES: Readonly<Record<string, number>> = { s: 1, c: 2, h: 3, d: 4 };

/** Parse a Watten (ヴァッテン) CLI command into API exec arguments. */
export function parseWattenCommand(input: string): CliParseResult<WattenArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'declare': {
      const rank = RANK_CODES[args[0]?.toLowerCase() ?? ''];
      const suit = SUIT_CODES[args[1]?.toLowerCase() ?? ''];
      if (!rank || !suit) return { error: 'Usage: d <a|7|8|9|10|j|q|k> <s|c|h|d>' };
      return { args: ['declare', rank, suit] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', undefined, undefined, parsed.value] };
    }
    case 'rz':
    case 'raise':
      return { args: ['raise'] };
    case 'hold':
      return { args: ['respond', undefined, undefined, undefined, true] };
    case 'fold':
      return { args: ['respond', undefined, undefined, undefined, false] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Watten (ヴァッテン) CLI mode. */
export const WATTEN_HELP: string[] = [
  'd <rank> <s|c|h|d>           - Declare the Schlag rank + critical suit (dealer, Declare phase)',
  'p <idx>                      - Play a card (Play phase)',
  'rz/raise                     - Raise the stake (as lead)',
  'hold                         - Hold (accept) a pending raise',
  'fold                         - Fold (concede) a pending raise',
  'nr/nextround                 - Next deal',
  'h/hint                       - Show hint',
  'r/reset                      - Reset game',
];
