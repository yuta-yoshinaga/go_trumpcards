import type { tusacApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TuSacArgs = Parameters<typeof tusacApi.exec>;

const VALID_COMMANDS = [
  'draw',
  'd',
  'take',
  't',
  'meld',
  'm',
  'discard',
  'x',
  'next',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** CLI help text for Tu Sac. */
export const TUSAC_CLI_HELP = [
  'draw (d)         draw from the stock',
  'take (t)         take the top discard',
  'meld <n...> (m)  lay a combination (e.g. meld 1 4 7)',
  'discard <n> (x)  discard one card',
  'next             start the next round',
  'hint             show a hint',
  'log              show the action log',
  'reset (r)        restart',
];

/** Parse a Tu Sac CLI command into API exec arguments. */
export function parseTuSacCommand(input: string): CliParseResult<TuSacArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    // **山と捨て札は別のコマンド。** 引き先を引数にすると、付け忘れが
    // 「山から」に化けて、狙って拾った札が黙って流れる。
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 't':
    case 'take':
      return { args: ['take'] };
    case 'm':
    case 'meld': {
      const indexes = parseTuSacIndexes(args);
      if (indexes === null) return { error: 'Usage: meld <n...> (e.g. meld 1 4 7)' };
      return { args: ['meld', { indexes }] };
    }
    case 'x':
    case 'discard': {
      const index = parseIntArg(args, 0);
      if ('error' in index || index.value < 1) return { error: 'Usage: discard <n> (from 1)' };
      // **画面は 1 始まり、ワイヤは 0 始まり。**
      return { args: ['discard', { index: index.value - 1 }] };
    }
    case 'next':
      return { args: ['next'] };
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

/** Parses "1 4 7" into 0-based indexes, or null when any entry is invalid. */
function parseTuSacIndexes(args: string[]): number[] | null {
  if (args.length === 0) return null;
  const out: number[] = [];
  for (const a of args) {
    const n = Number.parseInt(a.trim(), 10);
    if (Number.isNaN(n) || n < 1) return null;
    out.push(n - 1);
  }
  return out;
}
