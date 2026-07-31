import { describe, expect, it } from 'vitest';
import { parseTrexCommand, TREX_HELP } from './trexCommands';

/** Narrow a parse result to its error branch -- CliParseResult is a union. */
function errorOf(input: string): string {
  const result = parseTrexCommand(input);
  if (!('error' in result)) throw new Error(`expected ${input} to fail, got args`);
  return result.error;
}

describe('parseTrexCommand', () => {
  it('parses a contract choice', () => {
    expect(parseTrexCommand('c 4')).toEqual({ args: ['choose', 4] });
    expect(parseTrexCommand('choose 2')).toEqual({ args: ['choose', 2] });
  });

  it('accepts contract zero, which is the king of hearts', () => {
    // 0 を「省略」と同じ扱いにすると、♥K 契約だけ選べなくなる。
    expect(parseTrexCommand('c 0')).toEqual({ args: ['choose', 0] });
  });

  it('rejects contracts outside the five', () => {
    expect(errorOf('c 5')).toBe('Invalid contract: 5');
    expect(errorOf('c -1')).toBe('Invalid contract: -1');
    expect(errorOf('c x')).toBe('Invalid contract: x');
  });

  it('sends the hand index in the PLAY position, not the contract position', () => {
    // どちらも裸の数字なので、位置を取り違えると別の意味になる。
    expect(parseTrexCommand('p 3')).toEqual({ args: ['play', undefined, 3] });
  });

  it('rejects missing or malformed values', () => {
    expect(errorOf('c')).toBe('Usage: c <contract>');
    expect(errorOf('p')).toBe('Usage: p <index>');
    expect(errorOf('p y')).toBe('Invalid card index: y');
  });

  it('parses pass, next, log and reset', () => {
    expect(parseTrexCommand('s')).toEqual({ args: ['pass'] });
    expect(parseTrexCommand('pass')).toEqual({ args: ['pass'] });
    expect(parseTrexCommand('n')).toEqual({ args: ['next'] });
    expect(parseTrexCommand('log')).toEqual({ args: ['log'] });
    expect(parseTrexCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a near miss and reports an unknown command otherwise', () => {
    expect(errorOf('chose')).toContain('Did you mean');
    expect(errorOf('zzzz')).toBe('Unknown command: zzzz');
  });

  it('documents every command it accepts', () => {
    const help = TREX_HELP.join('\n');
    for (const cmd of ['c <n>', 'p <i>', 's/pass', 'n/next', 'log', 'r/reset']) {
      expect(help).toContain(cmd);
    }
  });
});
