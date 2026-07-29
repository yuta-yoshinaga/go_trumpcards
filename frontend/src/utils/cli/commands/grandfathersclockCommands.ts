import type { grandfathersClockApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type GrandfathersClockArgs = Parameters<typeof grandfathersClockApi.exec>;

const VALID_COMMANDS = [
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
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Grandfather's Clock CLI command into API exec arguments. */
export function parseGrandfathersClockCommand(input: string): CliParseResult<GrandfathersClockArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'm':
    case 'move':
      return parseMoveCommand(args);
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'ac':
    case 'autocomplete':
      return { args: ['autocomplete'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'h':
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

function parseMoveCommand(args: string[]): CliParseResult<GrandfathersClockArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m t0 f7)' };
  }
  const fromTok = args[0].toLowerCase();
  if (!fromTok.startsWith('t')) {
    return { error: 'Invalid source: use t<col> (tableau)' };
  }
  const fromCol = Number(fromTok.slice(1));
  if (Number.isNaN(fromCol)) return { error: 'Usage: m t<col> ...' };

  const target = args[1].toLowerCase();
  if (target.startsWith('f')) {
    // Twelve faces can hold the same suit, so the index is required — unlike
    // the one-per-suit foundations of the other solitaires.
    const fIdx = Number(target.slice(1));
    if (Number.isNaN(fIdx) || target.length === 1) {
      return { error: 'Usage: m t<col> f<idx> — the clock face index is required' };
    }
    return { args: ['move', { zone: 'tableau', col: fromCol }, { zone: 'foundation', col: fIdx }] };
  }
  if (target.startsWith('t')) {
    const toCol = Number(target.slice(1));
    if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> t<col>' };
    return { args: ['move', { zone: 'tableau', col: fromCol }, { zone: 'tableau', col: toCol }] };
  }
  return { error: 'Invalid target: use t<col> (tableau) or f<idx> (clock face)' };
}

/** Help text for Grandfather's Clock CLI mode. */
export const GRANDFATHERSCLOCK_HELP: string[] = [
  'm t<c> f<i>     - Move to clock face i (0..11)',
  'm t<c> t<c>     - Move the top card between columns',
  'ac/autocomplete - Auto-complete to the clock faces',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
