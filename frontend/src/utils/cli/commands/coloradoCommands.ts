import type { coloradoApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type ColoradoArgs = Parameters<typeof coloradoApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
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
 * Parse a Colorado CLI command into API exec arguments.
 *
 * A tableau card has exactly one legal destination — a foundation — so
 * `m t <col>` needs no destination zone. The trailing `f` is still accepted so
 * that muscle memory from the other solitaires does not produce an error.
 */
export function parseColoradoCommand(input: string): CliParseResult<ColoradoArgs> {
  const { cmd, args } = splitCommand(input);
  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'd':
    case 'draw':
      return { args: ['draw'] };
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
      if (args[0] === 'w') {
        if (args[1] === 'f') {
          return { args: ['move', { zone: 'waste' }, { zone: 'foundation' }] };
        }
        if (args[1] === 't') {
          const idx = parseIntArg(args, 2);
          if ('error' in idx) return { error: idx.error };
          return { args: ['move', { zone: 'waste' }, { zone: 'tableau', idx: idx.value }] };
        }
        return { error: 'Usage: m w f | m w t <idx>' };
      }
      if (args[0] === 't') {
        const from = parseIntArg(args, 1);
        if ('error' in from) return { error: from.error };
        if (args[2] !== undefined && args[2] !== 'f') {
          return { error: 'Usage: m t <col> (a tableau card can only go to a foundation)' };
        }
        return { args: ['move', { zone: 'tableau', idx: from.value }, { zone: 'foundation' }] };
      }
      if (args[0] === 's') {
        if (args[1] !== 't') return { error: 'Usage: m s t <idx>' };
        const idx = parseIntArg(args, 2);
        if ('error' in idx) return { error: idx.error };
        return { args: ['move', { zone: 'stock' }, { zone: 'tableau', idx: idx.value }] };
      }
      return { error: 'Usage: m w|t|s ...' };
    }
    default:
      return { error: suggestCommand(cmd, VALID_COMMANDS) ?? `Unknown command: ${cmd}` };
  }
}
