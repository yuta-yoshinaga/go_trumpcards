import type { CrescentMoveZone, crescentApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CrescentArgs = Parameters<typeof crescentApi.exec>;

const VALID_COMMANDS = [
  'm',
  'move',
  'd',
  'redeal',
  'u',
  'undo',
  'h',
  'hint',
  'a',
  'ac',
  'autocomplete',
  'g',
  'giveup',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

// Parse a t<col>/f<col>/f zone token into a CrescentMoveZone, or null if invalid.
function parseZone(tok: string): CrescentMoveZone | null {
  const t = tok.toLowerCase();
  if (t === 'f') return { zone: 'foundation' };
  const rest = t.slice(1);
  if (t.startsWith('t')) {
    const col = Number(rest);
    return Number.isNaN(col) ? null : { zone: 'tableau', col };
  }
  if (t.startsWith('f')) {
    const col = Number(rest);
    return Number.isNaN(col) ? null : { zone: 'foundation', col };
  }
  return null;
}

/** Parse a Crescent Solitaire CLI command into API exec arguments. */
export function parseCrescentCommand(input: string): CliParseResult<CrescentArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'm':
    case 'move': {
      if (args.length < 2) {
        return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m t0 f, m t0 f2)' };
      }
      const from = parseZone(args[0]);
      const to = parseZone(args[1]);
      if (!from) return { error: 'Invalid source: use t<col> (tableau) or f<col> (foundation)' };
      if (!to) return { error: 'Invalid target: use t<col> (tableau) or f / f<col> (foundation)' };
      return { args: ['move', from, to] };
    }
    case 'd':
    case 'redeal':
      return { args: ['redeal'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'a':
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
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

/** Help text for Crescent Solitaire CLI mode. */
export const CRESCENT_HELP: string[] = [
  'm t<c> t<c> - Move tableau top to column',
  'm t<c> f    - Move to any foundation',
  'm t<c> f<i> - Move to specific foundation',
  'd/redeal    - Redeal (shift each column)',
  'a/ac        - Auto-complete to foundations',
  'u/undo      - Undo last move',
  'h/hint      - Show suggested move',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
