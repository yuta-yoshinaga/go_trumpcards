import type { slyFoxApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SlyFoxArgs = Parameters<typeof slyFoxApi.exec>;

const VALID_COMMANDS = [
  'd',
  'deal',
  'm',
  'move',
  'g',
  'giveup',
  'ac',
  'autocomplete',
  'u',
  'undo',
  'h',
  'hint',
  'log',
  'l',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parse a Sly Fox CLI command into API exec arguments.
 *
 * A reserve card has exactly one legal destination — a foundation — so
 * `m t <slot>` needs no destination zone. The trailing `f` is still accepted so
 * that muscle memory from the other solitaires does not produce an error.
 *
 * **There is no waste, and the stock is not a move source.** A dealt card is
 * placed by naming its destination: `d <slot>` or `d f <foundation>`.
 */
export function parseSlyFoxCommand(input: string): CliParseResult<SlyFoxArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'deal': {
      if (args[0] === 'f') {
        const fIdx = parseIntArg(args, 1);
        if ('error' in fIdx) return { error: fIdx.error };
        return { args: ['deal', undefined, { zone: 'foundation', idx: fIdx.value }] };
      }
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: d <slot> | d f <foundation>' };
      return { args: ['deal', undefined, { zone: 'tableau', idx: idx.value }] };
    }
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'l':
    case 'log':
      return { args: ['log'] };
    case 'm':
    case 'move': {
      if (args[0] === 't') {
        const from = parseIntArg(args, 1);
        if ('error' in from) return { error: from.error };
        if (args[2] !== undefined && args[2] !== 'f') {
          return { error: 'Usage: m t <slot> (a reserve card can only go to a foundation)' };
        }
        return { args: ['move', { zone: 'tableau', idx: from.value }, { zone: 'foundation' }] };
      }
      // **捨て札も山札も移動元ではない。**リネームだけして残すと、サーバが
      // 400 で弾くまで分からない構文が CLI で通ってしまう。
      return { error: 'Usage: m t <slot> (there is no waste, and the stock is not a move source)' };
    }
    default:
      return { error: suggestCommand(cmd, VALID_COMMANDS) ?? `Unknown command: ${cmd}` };
  }
}
