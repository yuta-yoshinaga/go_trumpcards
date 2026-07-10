import type { kingAlbertApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type KingAlbertArgs = Parameters<typeof kingAlbertApi.exec>;

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

/** Parse a King Albert CLI command into API exec arguments. */
export function parseKingAlbertCommand(input: string): CliParseResult<KingAlbertArgs> {
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

/** Parse the source token into a move-zone. Supports t<col> (tableau) and r<idx> (reserve). */
function parseSource(tok: string): { zone: 'tableau' | 'reserve'; col: number } | { error: string } {
  if (tok.startsWith('t')) {
    const col = Number(tok.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m t<col> ...' };
    return { zone: 'tableau', col };
  }
  if (tok.startsWith('r')) {
    const col = Number(tok.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m r<idx> ...' };
    return { zone: 'reserve', col };
  }
  return { error: 'Invalid source: use t<col> (tableau) or r<idx> (reserve)' };
}

function parseMoveCommand(args: string[]): CliParseResult<KingAlbertArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m t0 t1, m t0 f, m r0 t1, m r0 f)' };
  }
  const fromTok = args[0].toLowerCase();
  const src = parseSource(fromTok);
  if ('error' in src) return { error: src.error };

  // Optional cardIndex on a tableau source: m t<col> <i> t<col>|f
  let cardIndex: number | undefined;
  let toIdx = 1;
  if (src.zone === 'tableau' && args.length >= 3 && /^\d+$/.test(args[1])) {
    const ci = parseIntArg(args, 1);
    if ('error' in ci) return { error: 'Invalid card index' };
    cardIndex = ci.value;
    toIdx = 2;
  }

  const from =
    src.zone === 'tableau'
      ? { zone: 'tableau' as const, col: src.col, cardIndex }
      : { zone: 'reserve' as const, col: src.col };

  const target = args[toIdx]?.toLowerCase() ?? '';
  if (target.startsWith('t')) {
    const toCol = Number(target.slice(1));
    if (Number.isNaN(toCol)) return { error: 'Usage: m <from> t<col>' };
    return { args: ['move', from, { zone: 'tableau', col: toCol }] };
  }
  if (target === 'f') {
    return { args: ['move', from, { zone: 'foundation' }] };
  }
  if (target.startsWith('f')) {
    const fCol = Number(target.slice(1));
    if (Number.isNaN(fCol)) return { error: 'Usage: m <from> f<idx>' };
    return { args: ['move', from, { zone: 'foundation', col: fCol }] };
  }
  return { error: 'Invalid target: use t<col> (tableau) or f / f<idx> (foundation)' };
}

/** Help text for King Albert CLI mode. */
export const KINGALBERT_HELP: string[] = [
  'm t<c> t<c>     - Move tableau top to column',
  'm t<c> <i> t<c> - Move card at index to column',
  'm t<c> f        - Move to any foundation',
  'm t<c> f<i>     - Move to specific foundation',
  'm r<i> t<c>     - Move reserve card to column',
  'm r<i> f        - Move reserve card to foundation',
  'ac/autocomplete - Auto-complete to foundation',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
