import { describe, expect, it } from 'vitest';
import { KILLE_HELP, parseKilleCommand } from './killeCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseKilleCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseKilleCommand', () => {
  it('parses every action and its alias', () => {
    for (const [input, want] of [
      ['e', 'exchange'],
      ['exchange', 'exchange'],
      ['s', 'satisfied'],
      ['satisfied', 'satisfied'],
      ['re', 'reenter'],
      ['reenter', 'reenter'],
      ['nr', 'nextround'],
      ['nextround', 'nextround'],
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseKilleCommand(input)).toEqual({ args: [want] });
    }
  });

  // **config は reset でしか受け付けない。**掛け金変更はリセットを伴う。
  it('turns a stake change into a reset carrying the config', () => {
    expect(parseKilleCommand('st 10')).toEqual({ args: ['reset', { config: { stake: 10 } }] });
    expect(parseKilleCommand('setstake 1')).toEqual({ args: ['reset', { config: { stake: 1 } }] });
  });

  it('rejects a stake outside 1-100', () => {
    for (const input of ['st', 'st abc', 'st 0', 'st 101']) {
      expect(parseError(input)).toContain('Usage: st');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('exchang')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['e / exchange', 's / satisfied', 're / reenter', 'nr / nextround', 'st <1-100>']) {
      expect(KILLE_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
  });
});
