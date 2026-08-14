import { describe, expect, it } from 'vitest';
import { HORSE_HELP, parseHorseCommand } from './horseCommands';

describe('parseHorseCommand', () => {
  it.each([
    ['f', 'fold'],
    ['fold', 'fold'],
    ['x', 'check'],
    ['check', 'check'],
    ['c', 'call'],
    ['call', 'call'],
    ['allin', 'allin'],
  ])('parses %s', (input, action) => {
    expect(parseHorseCommand(input)).toEqual({ args: ['action', { action }] });
  });

  it('parses bet and raise with an amount', () => {
    expect(parseHorseCommand('b 50')).toEqual({ args: ['action', { action: 'bet', amount: 50 }] });
    expect(parseHorseCommand('raise 120')).toEqual({ args: ['action', { action: 'raise', amount: 120 }] });
  });

  // **額の無いベットは送らない。** 送ってもサーバに断られるだけで理由が残らない。
  it.each(['b', 'bet', 'raise', 'b x'])('rejects %s without a usable amount', (input) => {
    const result = parseHorseCommand(input);
    expect(result).toHaveProperty('error');
  });

  it('parses next, hint and reset', () => {
    expect(parseHorseCommand('n')).toEqual({ args: ['next'] });
    expect(parseHorseCommand('h')).toEqual({ args: ['hint'] });
    expect(parseHorseCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command', () => {
    const result = parseHorseCommand('foldd');
    expect(result).toHaveProperty('error');
    expect('error' in result && result.error).toContain('fold');
  });

  it('rejects a command with nothing close to it', () => {
    expect(parseHorseCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  it('documents every action in the help text', () => {
    for (const token of ['fold', 'check', 'call', 'b <amount>', 'raise <amount>', 'allin', 'next']) {
      expect(HORSE_HELP.join('\n')).toContain(token);
    }
  });
});
