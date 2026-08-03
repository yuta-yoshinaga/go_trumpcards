import type { shengjiApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

/** Args tuple accepted by shengjiApi.exec. */
export type ShengJiCliArgs = Parameters<typeof shengjiApi.exec>;

const VALID_COMMANDS = ['d', 'declare', 'b', 'bury', 'p', 'play', 'n', 'next', 'l', 'log', 'r', 'reset', 'help', '?'];

/** Cards dealt to each seat, plus the kitty the declarer picks up. */
const HAND_MAX = 32;

/** Cards buried back into the kitty (sync: `ShengJiKittySize`). */
const KITTY_SIZE = 8;

/** Suit bounds. **Zero is a pass, not an omission.** */
const SUIT_MIN = 0;
const SUIT_MAX = 4;

/** Parses `<cmd> <idx...>`, optionally insisting on an exact count. */
function parseIndexes(args: string[], want: number): number[] | string {
  if (args.length === 0) {
    return 'Usage: card indexes are required';
  }
  if (want > 0 && args.length !== want) {
    return `Usage: give exactly ${want} card indexes`;
  }
  const out: number[] = [];
  for (const a of args) {
    const v = Number.parseInt(a, 10);
    if (Number.isNaN(v) || v < 0 || v > HAND_MAX) {
      return `Usage: every index is 0-${HAND_MAX}`;
    }
    // **同じ札を 2 回数えられない。**通すと 1 枚から対子が作れてしまう。
    if (out.includes(v)) {
      return `Usage: index ${v} was given twice`;
    }
    out.push(v);
  }
  return out;
}

/**
 * Parses a single CLI command line for the Sheng Ji game into
 * {@link shengjiApi}.exec arguments.
 *
 * `declare` takes **0 through 4**, where **0 passes** — declaring is a race to
 * show a level card, not an auction, and passing has to be expressible. Which
 * showings override which, and whether you actually hold the level card, are
 * settled on the server where the hand is known.
 */
export function parseShengJiCommand(input: string): CliParseResult<ShengJiCliArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'declare': {
      const suit = Number.parseInt(args[0] ?? '', 10);
      if (Number.isNaN(suit) || suit < SUIT_MIN || suit > SUIT_MAX) {
        return { error: 'Usage: d <0-4> (0 passes, 1=S 2=C 3=H 4=D)' };
      }
      return { args: ['declare', { suit }] };
    }
    case 'b':
    case 'bury': {
      const idxs = parseIndexes(args, KITTY_SIZE);
      if (typeof idxs === 'string') return { error: idxs };
      return { args: ['bury', { cardIndexes: idxs }] };
    }
    case 'p':
    case 'play': {
      const idxs = parseIndexes(args, 0);
      if (typeof idxs === 'string') return { error: idxs };
      return { args: ['play', { cardIndexes: idxs }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
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

/** Help text shown in the CLI terminal for Sheng Ji. */
export const SHENGJI_HELP: string[] = [
  'd <0-4>               - Declare trump (0 passes, 1=S 2=C 3=H 4=D)',
  'b <idx x8>            - Bury eight cards back into the kitty',
  'p <idx...>            - Play (e.g. p 0 1 for a pair)',
  'n / next              - Deal the next hand',
  'l / log               - Show action log',
  'r / reset             - Reset game',
];
