import type { fourseasonsApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type FourSeasonsArgs = Parameters<typeof fourseasonsApi.exec>;

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

/** Zone letters accepted as a move destination. */
function destZone(letter: string): 'foundation' | 'tableau' | null {
  if (letter === 'f') return 'foundation';
  if (letter === 't') return 'tableau';
  return null;
}

/**
 * Parse a Four Seasons CLI command into API exec arguments.
 *
 * A foundation destination always needs an index: which corner opens is decided
 * by the order they are started, so it cannot be implied the way a single-pile
 * game's can.
 */
export function parseFourSeasonsCommand(input: string): CliParseResult<FourSeasonsArgs> {
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
        const zone = destZone(args[1] ?? '');
        if (!zone) return { error: 'Usage: m w f|t <idx>' };
        const idx = parseIntArg(args, 2);
        if ('error' in idx) return { error: idx.error };
        return { args: ['move', { zone: 'waste' }, { zone, idx: idx.value }] };
      }
      if (args[0] === 't') {
        const from = parseIntArg(args, 1);
        if ('error' in from) return { error: from.error };
        const zone = destZone(args[2] ?? '');
        if (!zone) return { error: 'Usage: m t <col> f|t <idx>' };
        const idx = parseIntArg(args, 3);
        if ('error' in idx) return { error: idx.error };
        return { args: ['move', { zone: 'tableau', idx: from.value }, { zone, idx: idx.value }] };
      }
      return { error: 'Usage: m w|t ...' };
    }
    default:
      return { error: suggestCommand(cmd, VALID_COMMANDS) ?? `Unknown command: ${cmd}` };
  }
}
