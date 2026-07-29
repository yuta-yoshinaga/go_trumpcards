import type { duchessApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type DuchessArgs = Parameters<typeof duchessApi.exec>;

const VALID_COMMANDS = [
  'b',
  'base',
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

/** Parse a Duchess CLI command into API exec arguments. */
export function parseDuchessCommand(input: string): CliParseResult<DuchessArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'b':
    case 'base':
      return parseBaseCommand(args);
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

/** `b <fan>` — take the base rank off the top of a reserve fan. */
function parseBaseCommand(args: string[]): CliParseResult<DuchessArgs> {
  if (args.length === 0) return { error: 'Usage: b <fan>' };
  const fan = Number(args[0]);
  if (Number.isNaN(fan)) return { error: 'Usage: b <fan>' };
  return { args: ['base', { zone: 'reserve', col: fan }] };
}

function parseTableauToken(tok: string): { col: number; cardIndex?: number } | { error: string } {
  const lower = tok.toLowerCase();
  if (!lower.startsWith('t')) return { error: 'Invalid source: use t<col> (tableau)' };
  const [colTok, idxTok] = lower.slice(1).split('.');
  const col = Number(colTok);
  if (Number.isNaN(col) || colTok === '') return { error: 'Usage: t<col> ...' };
  if (idxTok === undefined) return { col };
  const cardIndex = Number(idxTok);
  if (Number.isNaN(cardIndex)) return { error: 'Usage: t<col>.<idx> ...' };
  return { col, cardIndex };
}

function parseTarget(target: string): { zone: string; col?: number } | { error: string } {
  if (target === 'f') return { zone: 'foundation' };
  if (target.startsWith('t')) {
    const col = Number(target.slice(1));
    if (Number.isNaN(col) || target.slice(1) === '') return { error: 'Usage: ... t<col>' };
    return { zone: 'tableau', col };
  }
  return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
}

function parseMoveCommand(args: string[]): CliParseResult<DuchessArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m r0 f, m r0 t1, m w f, m w t1, m t0 f, m t0.2 t1)' };
  }
  const fromTok = args[0].toLowerCase();
  const target = parseTarget(args[1].toLowerCase());
  if ('error' in target) return target;

  if (fromTok === 'w') {
    return { args: ['move', { zone: 'waste' }, target] };
  }
  if (fromTok.startsWith('r')) {
    const fan = Number(fromTok.slice(1));
    if (Number.isNaN(fan) || fromTok.slice(1) === '') return { error: 'Usage: m r<fan> <to>' };
    return { args: ['move', { zone: 'reserve', col: fan }, target] };
  }

  const parsed = parseTableauToken(fromTok);
  if ('error' in parsed) return parsed;
  return { args: ['move', { zone: 'tableau', col: parsed.col, cardIndex: parsed.cardIndex }, target] };
}

/** Help text for Duchess CLI mode. */
export const DUCHESS_HELP: string[] = [
  'b <fan>         - Take the base rank off a reserve fan (0-3)',
  'd/draw          - Turn one card from the stock',
  'm r<f> f        - Reserve fan top to a foundation',
  'm r<f> t<c>     - Reserve fan top to a tableau column',
  'm w f           - Waste to a foundation',
  'm w t<c>        - Waste to a tableau column',
  'm t<c> f        - Tableau top to a foundation',
  'm t<c> t<c>     - Move the top card between columns',
  'm t<c>.<i> t<c> - Move the run starting at index i',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
