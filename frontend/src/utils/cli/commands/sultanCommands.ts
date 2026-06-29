import type { sultanApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SultanArgs = Parameters<typeof sultanApi.exec>;

const VALID_COMMANDS = [
  'd',
  'draw',
  'rd',
  'redeal',
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

/** Parse a Sultan of Turkey CLI command into API exec arguments. */
export function parseSultanCommand(input: string): CliParseResult<SultanArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'd':
    case 'draw':
      return { args: ['draw'] };
    case 'rd':
    case 'redeal':
      return { args: ['redeal'] };
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

/** Strict divan index token: only a non-negative integer (rejects 'd', 'd-1', '1.5'). */
const INDEX_TOKEN = /^\d+$/;

function parseMoveCommand(args: string[]): CliParseResult<SultanArgs> {
  if (args.length < 1) {
    return { error: 'Usage: m d <idx> (divan→foundation), m w (waste→foundation), or m <idx> (divan→foundation)' };
  }
  const first = args[0].toLowerCase();

  // Waste → foundation: m w
  if (first === 'w') {
    return { args: ['move', { zone: 'waste' }] };
  }

  // Divan → foundation: m d <idx>
  if (first === 'd') {
    if (args.length < 2 || !INDEX_TOKEN.test(args[1])) {
      return { error: 'Invalid divan index: use m d <idx>' };
    }
    return { args: ['move', { zone: 'divan', divanIdx: Number(args[1]) }] };
  }

  // Shorthand: m <idx> → divan→foundation
  if (INDEX_TOKEN.test(first)) {
    return { args: ['move', { zone: 'divan', divanIdx: Number(first) }] };
  }

  return { error: 'Invalid source: use d <idx> (divan), w (waste), or <idx> (divan)' };
}

/** Help text for Sultan of Turkey CLI mode. */
export const SULTAN_HELP: string[] = [
  'd/draw          - Draw a card from the stock',
  'rd/redeal       - Recycle waste into stock (up to twice)',
  'm d <idx>       - Play a divan slot onto its foundation',
  'm <idx>         - Play a divan slot onto its foundation',
  'm w             - Play the waste top onto its foundation',
  'ac/autocomplete - Auto-complete to foundation',
  'u/undo          - Undo last move',
  'h/hint          - Show suggested move',
  'g/giveup        - Give up',
  'log             - Show action log',
  'r/reset         - Reset game',
];
