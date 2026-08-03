import { describe, expect, it } from 'vitest';
import { KLABERJASS_HELP, parseKlaberjassCommand } from './klaberjassCommands';

/** The error message, or '' when the input parsed successfully. */
function parseError(input: string): string {
  const result = parseKlaberjassCommand(input);
  return 'error' in result ? result.error : '';
}

describe('parseKlaberjassCommand', () => {
  it('parses the parameterless actions and their aliases', () => {
    for (const [input, want] of [
      ['a', 'accept'],
      ['accept', 'accept'],
      ['ps', 'pass'],
      ['pass', 'pass'],
      ['sm', 'schmeiss'],
      ['schmeiss', 'schmeiss'],
      ['n', 'next'],
      ['next', 'next'],
      ['l', 'log'],
      ['log', 'log'],
      ['r', 'reset'],
      ['reset', 'reset'],
    ] as const) {
      expect(parseKlaberjassCommand(input)).toEqual({ args: [want] });
    }
  });

  it('parses a trump call', () => {
    expect(parseKlaberjassCommand('c 3')).toEqual({ args: ['call', { suit: 3 }] });
    expect(parseKlaberjassCommand('call 1')).toEqual({ args: ['call', { suit: 1 }] });
  });

  it('rejects a suit outside 1-4', () => {
    for (const input of ['c', 'c abc', 'c 0', 'c 5']) {
      expect(parseError(input)).toContain('Usage: c');
    }
  });

  // **拒否は「同意しない」ではなく「相手を宣言側にする」。**別コマンドで区別する。
  it('tells agreeing to a schmeiss apart from refusing it', () => {
    expect(parseKlaberjassCommand('y')).toEqual({ args: ['answerschmeiss', { accept: true }] });
    expect(parseKlaberjassCommand('yes')).toEqual({ args: ['answerschmeiss', { accept: true }] });
    expect(parseKlaberjassCommand('no')).toEqual({ args: ['answerschmeiss', { accept: false }] });
  });

  it('parses a play and rejects an index outside the hand', () => {
    expect(parseKlaberjassCommand('p 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
    expect(parseKlaberjassCommand('play 8')).toEqual({ args: ['play', { cardIndex: 8 }] });
    for (const input of ['p', 'p abc', 'p -1', 'p 9']) {
      expect(parseError(input)).toContain('Usage: p');
    }
  });

  // **config は reset でしか受け付けない。**目標点の変更はリセットを伴う。
  it('turns a target change into a reset carrying the config', () => {
    expect(parseKlaberjassCommand('st 300')).toEqual({ args: ['reset', { config: { targetScore: 300 } }] });
    for (const input of ['st', 'st abc', 'st 99', 'st 1001']) {
      expect(parseError(input)).toContain('Usage: st');
    }
  });

  it('suggests a near miss and reports an unknown command', () => {
    expect(parseError('schmeis')).toContain('Did you mean');
    expect(parseError('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    for (const cmd of ['a / accept', 'c <1-4>', 'ps / pass', 'sm / schmeiss', 'y / no', 'p <0-8>', 'st <100-1000>']) {
      expect(KLABERJASS_HELP.some((line) => line.startsWith(cmd))).toBe(true);
    }
  });
});
