import type { terraceApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type TerraceArgs = Parameters<typeof terraceApi.exec>;

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

/** Parse a Terrace CLI command into API exec arguments. */
export function parseTerraceCommand(input: string): CliParseResult<TerraceArgs> {
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

function parsePileToken(tok: string): number | { error: string } {
  const body = tok.slice(1);
  const n = Number(body);
  if (body === '' || Number.isNaN(n)) return { error: 'Usage: t<pile>' };
  return n;
}

function parseMoveCommand(args: string[]): CliParseResult<TerraceArgs> {
  if (args.length < 2) {
    return { error: 'Usage: m <from> <to> (e.g., m r f, m w f, m w t2, m t0 f, m t0 t5)' };
  }
  const fromTok = args[0].toLowerCase();
  const toTok = args[1].toLowerCase();

  // The terrace feeds the foundations and nothing else.
  if (fromTok === 'r') {
    if (toTok !== 'f') return { error: 'Invalid move: a terrace card can only go to a foundation' };
    return { args: ['move', { zone: 'reserve' }, { zone: 'foundation' }] };
  }

  let from: { zone: string; col?: number };
  if (fromTok === 'w') {
    from = { zone: 'waste' };
  } else if (fromTok.startsWith('t')) {
    const pile = parsePileToken(fromTok);
    if (typeof pile !== 'number') return pile;
    from = { zone: 'tableau', col: pile };
  } else {
    return { error: 'Invalid source: use r (terrace), w (waste) or t<pile> (tableau)' };
  }

  if (toTok === 'f') return { args: ['move', from, { zone: 'foundation' }] };
  if (toTok.startsWith('t')) {
    const pile = parsePileToken(toTok);
    if (typeof pile !== 'number') return pile;
    return { args: ['move', from, { zone: 'tableau', col: pile }] };
  }
  return { error: 'Invalid target: use f (foundation) or t<pile> (tableau)' };
}

/** Help text for Terrace CLI mode. */
export const TERRACE_HELP: string[] = [
  'd/draw          - Turn one card from the stock (no redeal)',
  'm r f           - Terrace top to a foundation (its only destination)',
  'm w f           - Waste to a foundation',
  'm w t<p>        - Waste to a tableau pile',
  'm t<p> f        - Tableau top to a foundation',
  'm t<p> t<p>     - Move one card between piles',
  'ac/autocomplete - Auto-complete to the foundations',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
