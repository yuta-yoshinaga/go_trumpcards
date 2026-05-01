import type { canfieldApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type CanfieldArgs = Parameters<typeof canfieldApi.exec>;

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
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Canfield CLI command into API exec arguments. */
export function parseCanfieldCommand(input: string): CliParseResult<CanfieldArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
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

function parseSource(tok: string): { zone: string; col?: number } | { error: string } {
  const t = tok.toLowerCase();
  if (t === 'w' || t === 'waste') return { zone: 'waste' };
  if (t === 'rs' || t === 'reserve') return { zone: 'reserve' };
  if (t.startsWith('t')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m t<col> ...' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid source: use w (waste), rs (reserve), or t<col> (tableau)' };
}

function parseTarget(tok: string): { zone: string; col?: number } | { error: string } {
  const t = tok.toLowerCase();
  if (t === 'f' || t === 'foundation') return { zone: 'foundation' };
  if (t.startsWith('f')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m ... f<idx>' };
    return { zone: 'foundation', col };
  }
  if (t.startsWith('t')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m ... t<col>' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid target: use f / f<idx> (foundation) or t<col> (tableau)' };
}

function parseMoveCommand(args: string[]): CliParseResult<CanfieldArgs> {
  if (args.length < 2) return { error: 'Usage: m <from> <to> (e.g., m w f, m t0 t1, m t0 2 t1)' };
  const from = parseSource(args[0]);
  if ('error' in from) return { error: from.error };

  // Optional cardIndex on tableau source: m t0 <idx> <target>
  let cardIndex: number | undefined;
  let toIdx = 1;
  if (from.zone === 'tableau' && args.length >= 3 && /^\d+$/.test(args[1])) {
    const ci = parseIntArg(args, 1);
    if ('error' in ci) return { error: 'Invalid card index' };
    cardIndex = ci.value;
    toIdx = 2;
  }

  const target = args[toIdx];
  if (!target) return { error: 'Missing target zone' };
  const to = parseTarget(target);
  if ('error' in to) return { error: to.error };

  return {
    args: ['move', { zone: from.zone, col: from.col, cardIndex }, { zone: to.zone, col: to.col }],
  };
}

/** Help text for Canfield CLI mode. */
export const CANFIELD_HELP: string[] = [
  'd/draw          - Draw from stock',
  'm w f           - Move waste to foundation',
  'm w t<c>        - Move waste to tableau column',
  'm rs f          - Move reserve to foundation',
  'm rs t<c>       - Move reserve to tableau column',
  'm t<c> f        - Move tableau top to foundation',
  'm t<c> t<c>     - Move tableau to tableau',
  'm t<c> <i> t<c> - Move card at index to column',
  'm t<c> f<i>     - Move to specific foundation',
  'ac/autocomplete - Auto-complete to foundation',
  'u/undo          - Undo last move',
  'h/hint          - Get a hint',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
