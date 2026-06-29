import type { agnesApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type AgnesArgs = Parameters<typeof agnesApi.exec>;

const VALID_COMMANDS = [
  'd',
  'deal',
  'm',
  'move',
  'g',
  'giveup',
  'u',
  'undo',
  'h',
  'hint',
  'l',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse an Agnes Sorel CLI command into API exec arguments. */
export function parseAgnesCommand(input: string): CliParseResult<AgnesArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'deal':
      return { args: ['deal'] };
    case 'm':
    case 'move':
      return parseMoveCommand(args);
    case 'g':
    case 'giveup':
      return { args: ['giveup'] };
    case 'u':
    case 'undo':
      return { args: ['undo'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
    case 'l':
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
  if (t.startsWith('t')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m t<col> ...' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid source: use t<col> (tableau)' };
}

function parseTarget(tok: string): { zone: string; col?: number } | { error: string } {
  const t = tok.toLowerCase();
  if (t === 'f' || t === 'foundation') return { zone: 'foundation' };
  if (t.startsWith('t')) {
    const col = Number(t.slice(1));
    if (Number.isNaN(col)) return { error: 'Usage: m ... t<col>' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
}

function parseMoveCommand(args: string[]): CliParseResult<AgnesArgs> {
  if (args.length < 2) return { error: 'Usage: m <from> <to> (e.g., m t0 f, m t0 t1, m t0 3 t1)' };
  const from = parseSource(args[0]);
  if ('error' in from) return { error: from.error };

  // Optional cardIndex on tableau source: m t0 <idx> <target>
  let cardIndex: number | undefined;
  let toIdx = 1;
  if (args.length >= 3 && /^\d+$/.test(args[1])) {
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

/** Help text for Agnes Sorel CLI mode. */
export const AGNES_HELP: string[] = [
  'd/deal          - Deal one card to each tableau column',
  'm t<c> f        - Move tableau end card to foundation',
  'm t<c> t<c>     - Move tableau end card to another column',
  'm t<c> <i> t<c> - Move card at index to column',
  'u/undo          - Undo last move',
  'h/hint          - Get a hint',
  'g/giveup        - Give up',
  'l/log           - Show action log',
  'r/reset         - Reset game',
];
