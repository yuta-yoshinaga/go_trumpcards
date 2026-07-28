import type { napoleonsSquareApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type NapoleonsSquareArgs = Parameters<typeof napoleonsSquareApi.exec>;

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

/** Parse a Napoleon's Square CLI command into API exec arguments. */
export function parseNapoleonsSquareCommand(input: string): CliParseResult<NapoleonsSquareArgs> {
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

function parseMoveCommand(args: string[]): CliParseResult<NapoleonsSquareArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m w f, m w t3, m t0 f, m t0 t5, m t0.2 t5)' };
  }
  const fromTok = args[0].toLowerCase();
  const target = args[1].toLowerCase();

  if (fromTok === 'w') {
    if (target === 'f') {
      return { args: ['move', { zone: 'waste' }, { zone: 'foundation' }] };
    }
    if (target.startsWith('t')) {
      const col = Number(target.slice(1));
      if (Number.isNaN(col)) return { error: 'Usage: m w t<col>' };
      return { args: ['move', { zone: 'waste' }, { zone: 'tableau', col }] };
    }
    return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
  }

  if (!fromTok.startsWith('t')) {
    return { error: 'Invalid source: use w (waste) or t<col> (tableau)' };
  }
  // t<col> optionally carries the run head as t<col>.<idx>; omitted means the top card.
  const [colTok, idxTok] = fromTok.slice(1).split('.');
  const fromCol = Number(colTok);
  if (Number.isNaN(fromCol)) return { error: 'Usage: m t<col> ...' };
  let cardIndex: number | undefined;
  if (idxTok !== undefined) {
    const idx = Number(idxTok);
    if (Number.isNaN(idx)) return { error: 'Usage: m t<col>.<idx> ...' };
    cardIndex = idx;
  }

  if (target === 'f') {
    return { args: ['move', { zone: 'tableau', col: fromCol }, { zone: 'foundation' }] };
  }
  if (target.startsWith('t')) {
    const toCol = Number(target.slice(1));
    if (Number.isNaN(toCol)) return { error: 'Usage: m t<col> t<col>' };
    return { args: ['move', { zone: 'tableau', col: fromCol, cardIndex }, { zone: 'tableau', col: toCol }] };
  }
  return { error: 'Invalid target: use f (foundation) or t<col> (tableau)' };
}

/** Help text for Napoleon's Square CLI mode. */
export const NAPOLEONSSQUARE_HELP: string[] = [
  'd/draw          - Turn one card from the stock',
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
