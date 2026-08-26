import type { sutdaApi } from '../../../api/gameApi';
import { splitCommand, suggestCommand } from '../commandParserBase';
import type { CliParseResult } from '../types';

type SutdaArgs = Parameters<typeof sutdaApi.exec>;

const VALID_COMMANDS = [
  'c',
  'call',
  'b',
  'raise',
  'f',
  'fold',
  'nh',
  'nexthand',
  'h',
  'hint',
  'r',
  'reset',
  'help',
  '?',
];

/**
 * Parse a Sutda CLI command into API exec arguments.
 *
 * **Every command here is an action, not a card.** Sutda hands are dealt whole
 * and never changed, so nothing takes an index.
 */
export function parseSutdaCommand(input: string): CliParseResult<SutdaArgs> {
  const { cmd } = splitCommand(input);

  switch (cmd) {
    case 'c':
    case 'call':
      return { args: ['call'] };
    case 'b':
    case 'raise':
      return { args: ['raise'] };
    case 'f':
    case 'fold':
      return { args: ['fold'] };
    case 'nh':
    case 'nexthand':
      return { args: ['nexthand'] };
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

/** Help text for Sutda CLI mode. */
export const SUTDA_HELP: string[] = [
  'call/c                           - Call (check when nothing is owed)',
  'raise/b                          - Raise by one unit',
  'fold/f                           - Fold',
  'nh/nexthand                      - Next hand',
  'h/hint                           - Show hint',
  'r/reset                          - Reset game',
];
