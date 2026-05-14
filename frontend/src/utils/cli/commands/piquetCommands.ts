import type { piquetApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PiquetArgs = Parameters<typeof piquetApi.exec>;

const VALID_COMMANDS = [
  'e',
  'elder',
  'y',
  'younger',
  'd',
  'declare',
  'p',
  'play',
  'nd',
  'nextdeal',
  'h',
  'hint',
  'log',
  'r',
  'reset',
  'help',
  '?',
];

/** Parse a Piquet CLI command into API exec arguments. */
export function parsePiquetCommand(input: string): CliParseResult<PiquetArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 'r':
    case 'reset':
      return { args: ['reset'] };
    case 'e':
    case 'elder':
      return parseExchange('e', args);
    case 'y':
    case 'younger':
      return parseExchange('y', args);
    case 'd':
    case 'declare':
      return { args: ['d'] };
    case 'p':
    case 'play': {
      const idx = parseIntArg(args, 0);
      if ('error' in idx) return { error: 'Usage: p <cardIndex>' };
      return { args: ['p', idx.value] };
    }
    case 'nd':
    case 'nextdeal':
      return { args: ['nd'] };
    case 'h':
    case 'hint':
      return { args: ['h'] };
    case 'log':
      return { args: ['log'] };
    default: {
      const suggestion = suggestCommand(cmd, VALID_COMMANDS);
      if (suggestion) return { error: `Unknown command: ${cmd}. Did you mean: ${suggestion}?` };
      return { error: `Unknown command: ${cmd}` };
    }
  }
}

function parseExchange(cmd: 'e' | 'y', args: string[]): CliParseResult<PiquetArgs> {
  if (args.length === 0) {
    // Younger may pass with 0 discards
    if (cmd === 'y') return { args: [cmd, undefined, []] };
    return { error: 'Usage: e <i,j,k> (1..5 indices)' };
  }
  const joined = args.join('').replace(/\s+/g, '');
  const parts = joined.split(',').filter((p) => p.length > 0);
  const indices: number[] = [];
  for (const p of parts) {
    const n = Number(p);
    if (!Number.isInteger(n) || n < 0) return { error: `Invalid index: ${p}` };
    indices.push(n);
  }
  return { args: [cmd, undefined, indices] };
}

/** Help text for Piquet CLI mode. */
export const PIQUET_HELP: string[] = [
  'e <i,j,k>       - Elder exchange (1..5 card indices)',
  'y <i,j,k>       - Younger exchange (0..3 indices); "y" alone = pass',
  'd/declare       - Advance one declaration step (Point→Sequence→Set)',
  'p <i>           - Play card at index',
  'nd/nextdeal     - Advance to next deal after score phase',
  'h/hint          - Show suggested move',
  'log             - Show action log',
  'r/reset         - Reset game',
];
