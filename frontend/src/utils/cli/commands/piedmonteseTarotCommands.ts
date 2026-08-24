import type { piedmonteseTarotApi } from '../../../api/gameApi';
import { parseIntArg, splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type PiedmonteseTarotArgs = Parameters<typeof piedmonteseTarotApi.exec>;

const VALID_COMMANDS = [
  's',
  'scarto',
  'd',
  'discard',
  'p',
  'play',
  'n',
  'next',
  'nr',
  'nextround',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parse a Tarocco Piemontese CLI command into API exec arguments.
 *
 * **The scarto takes as many indices as you type.** Four seats bury two cards
 * and three seats bury three, so a parser that insists on a fixed count would
 * refuse a legal move at one of the two table sizes; the server checks the
 * count against the talon it actually dealt.
 */
export function parsePiedmonteseTarotCommand(input: string): CliParseResult<PiedmonteseTarotArgs> {
  const { cmd, args } = splitCommand(input);

  switch (cmd) {
    case 's':
    case 'scarto':
    case 'd':
    case 'discard': {
      if (args.length === 0) return { error: 'Usage: scarto <i1> <i2> [i3]' };
      const indices: number[] = [];
      for (let i = 0; i < args.length; i++) {
        const parsed = parseIntArg(args, i);
        if ('error' in parsed) return { error: 'Usage: scarto <i1> <i2> [i3]' };
        indices.push(parsed.value);
      }
      return { args: ['scarto', { cardIndices: indices }] };
    }
    case 'p':
    case 'play': {
      const parsed = parseIntArg(args, 0);
      if ('error' in parsed) return { error: 'Usage: p <idx>' };
      return { args: ['play', { cardIndex: parsed.value }] };
    }
    case 'n':
    case 'next':
      return { args: ['next'] };
    case 'nr':
    case 'nextround':
      return { args: ['nextround'] };
    case 'h':
    case 'hint':
      return { args: ['hint'] };
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

/** Help text for Tarocco Piemontese CLI mode. */
export const PIEDMONTESE_TAROT_HELP: string[] = [
  's/scarto <i1> <i2> [i3]          - Bury the talon (dealer only; 2 cards at four seats, 3 at three)',
  'p <idx>                          - Play a card (must follow suit / trump / overtrump)',
  'n/next                           - Next trick',
  'nr/nextround                     - Next deal',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
