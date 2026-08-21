import type { stalactitesApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type StalactitesArgs = Parameters<typeof stalactitesApi.exec>;

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

/** Parse a Stalactites CLI command into API exec arguments. */
export function parseStalactitesCommand(input: string): CliParseResult<StalactitesArgs> {
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

function parseMoveCommand(args: string[]): CliParseResult<StalactitesArgs> {
  if (args.length < 2) return { error: 'Usage: m <from> <to> (e.g., m t0 f, m t0 3 t1, m t0 c0, m c0 t1)' };
  const fromZone = args[0].toLowerCase();

  // Parse "from" zone
  if (fromZone.startsWith('t')) {
    const col = Number(fromZone.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m t<col> ...' };

    // With card index: m t0 3 t1
    if (args.length >= 4 && args[2].toLowerCase().startsWith('t')) {
      const cardIdx = parseIntArg(args, 1);
      if ('error' in cardIdx) return { error: 'Usage: m t<col> <cardIdx> t<col>' };
      const toCol = Number(args[2].toLowerCase().slice(1));
      if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> <cardIdx> t<col>' };
      return { args: ['move', { zone: 'tableau', col, cardIndex: cardIdx.value }, { zone: 'tableau', col: toCol }] };
    }

    const to = args[1].toLowerCase();
    if (to === 'f') return { args: ['move', { zone: 'tableau', col }, { zone: 'foundation' }] };
    if (to.startsWith('c')) {
      const cell = Number(to.slice(1));
      if (Number.isNaN(cell)) return { error: 'Usage: m t<col> c<cell>' };
      return { args: ['move', { zone: 'tableau', col }, { zone: 'cell', cell }] };
    }
    if (to.startsWith('t')) {
      const toCol = Number(to.slice(1));
      if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> t<col>' };
      return { args: ['move', { zone: 'tableau', col }, { zone: 'tableau', col: toCol }] };
    }
    return { error: 'Invalid target: use f (foundation), t<col> (tableau), or c<cell> (free cell)' };
  }

  if (fromZone.startsWith('c')) {
    const cell = Number(fromZone.slice(1));
    if (Number.isNaN(cell)) return { error: 'Usage: m c<cell> ...' };
    const to = args[1].toLowerCase();
    if (to === 'f') return { args: ['move', { zone: 'cell', cell }, { zone: 'foundation' }] };
    if (to.startsWith('t')) {
      const toCol = Number(to.slice(1));
      if (Number.isNaN(toCol)) return { error: 'Usage: m c<cell> t<col>' };
      return { args: ['move', { zone: 'cell', cell }, { zone: 'tableau', col: toCol }] };
    }
    return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
  }

  return { error: 'Invalid source: use t<col> (tableau) or c<cell> (free cell)' };
}

/** Help text for Stalactites CLI mode. */
export const STALACTITES_HELP: string[] = [
  'm t<c> t<c> - Move tableau top to column',
  'm t<c> <i> t<c> - Move card at index to column',
  'm t<c> f    - Move to foundation',
  'm t<c> c<n> - Move to free cell',
  'm c<n> t<c> - Move from cell to tableau',
  'm c<n> f    - Move from cell to foundation',
  'ac          - Auto-complete',
  'u/undo      - Undo last move',
  'h/hint      - Get a hint',
  'g/giveup    - Give up',
  'log         - Show action log',
  'r/reset     - Reset game',
];
