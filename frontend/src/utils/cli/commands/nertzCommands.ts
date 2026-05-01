import type { nertzApi } from '../../../api/gameApi';
import type { NertzMoveZone } from '../../../types/card';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type NertzArgs = Parameters<typeof nertzApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'm',
  'move',
  't',
  'tick',
  'nr',
  'nextround',
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

/** Parse a Nertz CLI command into API exec arguments. */
export function parseNertzCommand(input: string): CliParseResult<NertzArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['d', { playerIdx: 0 }] };
    case 't':
    case 'tick':
      return { args: ['tick'] };
    case 'nr':
    case 'nextround':
      return { args: ['nr'] };
    case 'u':
    case 'undo':
      return { args: ['u'] };
    case 'h':
    case 'hint':
      return { args: ['h'] };
    case 'log':
      return { args: ['log'] };
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'm':
    case 'move':
      return parseMoveCommand(args);
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

function parseSource(tok: string): NertzMoveZone | { error: string } {
  const t = tok.toLowerCase();
  if (t === 'n' || t === 'nertz') return { zone: 'nertz' };
  if (t === 'w' || t === 'waste') return { zone: 'waste' };
  if (t.startsWith('t')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m t<col> ...' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid source: use n (nertz), w (waste), or t<col> (tableau)' };
}

function parseTarget(tok: string): NertzMoveZone | { error: string } {
  const t = tok.toLowerCase();
  if (t === 'f' || t === 'foundation') return { zone: 'foundation' };
  if (t.startsWith('f')) {
    const idx = Number(t.slice(1));
    if (Number.isNaN(idx)) return { error: 'Usage: m ... f<idx>' };
    return { zone: 'foundation', idx };
  }
  if (t.startsWith('t')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m ... t<col>' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid target: use f / f<idx> (foundation) or t<col> (tableau)' };
}

function parseMoveCommand(args: string[]): CliParseResult<NertzArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m n f, m t0 t1, m t0 2 t1)' };
  }
  const from = parseSource(args[0]);
  if ('error' in from) return { error: from.error };

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
    args: ['m', { playerIdx: 0, from: { ...from, cardIndex }, to }],
  };
}

/** Help text for Nertz CLI mode. */
export const NERTZ_HELP: string[] = [
  'd/draw          - Draw from stock',
  't/tick          - Step CPU players',
  'm n f           - Move from nertz pile to foundation',
  'm n t<c>        - Move from nertz to tableau',
  'm w f           - Move waste top to foundation',
  'm w t<c>        - Move waste to tableau',
  'm t<c> f        - Move tableau top to foundation',
  'm t<c> t<c>     - Move tableau to tableau',
  'm t<c> <i> t<c> - Move card at index to column',
  'm t<c> f<i>     - Move to specific foundation',
  'nr/nextround    - Start the next round',
  'u/undo          - Undo last move',
  'h/hint          - Get a hint',
  'log             - Show action log',
  'r/reset         - Reset match',
];
