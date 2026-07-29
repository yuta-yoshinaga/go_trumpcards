import type { americanToadApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type AmericanToadArgs = Parameters<typeof americanToadApi.exec>;

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

/** Parse an American Toad CLI command into API exec arguments. */
export function parseAmericanToadCommand(input: string): CliParseResult<AmericanToadArgs> {
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

function parseMoveCommand(args: string[]): CliParseResult<AmericanToadArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m r f, m r t1, m w f, m w t1, m t0 f, m t0.2 t1)' };
  }
  const fromTok = args[0].toLowerCase();
  const target = parseTarget(args[1].toLowerCase());
  if ('error' in target) return target;

  if (fromTok === 'r') return { args: ['move', { zone: 'reserve' }, target] };
  if (fromTok === 'w') return { args: ['move', { zone: 'waste' }, target] };

  const parsed = parseTableauToken(fromTok);
  if ('error' in parsed) return parsed;
  return { args: ['move', { zone: 'tableau', col: parsed.col, cardIndex: parsed.cardIndex }, target] };
}

/** Help text for American Toad CLI mode. */
export const AMERICANTOAD_HELP: string[] = [
  'd/draw          - Turn one card (redeals the waste once when the stock is out)',
  'm r f           - Reserve to a foundation',
  'm r t<c>        - Reserve to a tableau column',
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
