import { describe, expect, it } from 'vitest';
import { PIEDMONTESE_TAROT_HELP, parsePiedmonteseTarotCommand } from './piedmonteseTarotCommands';

describe('parsePiedmonteseTarotCommand', () => {
  // **枚数を決め打たない。** 4 人卓は 2 枚、3 人卓は 3 枚を捨てるので、
  // 固定の枚数しか受け付けないパーサはどちらかの卓で合法手を拒む。
  it.each([
    ['scarto 0 1', [0, 1]],
    ['s 0 1 2', [0, 1, 2]],
    ['discard 4', [4]],
    ['d 0 2 5', [0, 2, 5]],
  ] as const)('parses %s', (input, expected) => {
    expect(parsePiedmonteseTarotCommand(input)).toEqual({ args: ['scarto', { cardIndices: expected }] });
  });

  it.each(['scarto', 'scarto x', 's 0 y'])('rejects %s', (input) => {
    expect(parsePiedmonteseTarotCommand(input)).toHaveProperty('error');
  });

  it('parses a play', () => {
    expect(parsePiedmonteseTarotCommand('p 3')).toEqual({ args: ['play', { cardIndex: 3 }] });
    expect(parsePiedmonteseTarotCommand('play 0')).toEqual({ args: ['play', { cardIndex: 0 }] });
  });

  it('rejects a play with no index', () => {
    expect(parsePiedmonteseTarotCommand('play')).toHaveProperty('error');
  });

  it('parses the bare commands', () => {
    expect(parsePiedmonteseTarotCommand('n')).toEqual({ args: ['next'] });
    expect(parsePiedmonteseTarotCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parsePiedmonteseTarotCommand('h')).toEqual({ args: ['hint'] });
    expect(parsePiedmonteseTarotCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command', () => {
    const result = parsePiedmonteseTarotCommand('playy');
    expect(result).toHaveProperty('error');
    expect('error' in result && result.error).toContain('play');
  });

  it('rejects a command with nothing close to it', () => {
    expect(parsePiedmonteseTarotCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // **枚数が卓で変わることをヘルプが言う。** 固定の枚数を書くと、片方の卓では
  // 嘘になる。
  it('documents the commands, including the variable talon', () => {
    const help = PIEDMONTESE_TAROT_HELP.join('\n');
    for (const token of ['scarto', 'p <idx>', 'n/next', 'nr/nextround', 'hint', 'reset']) {
      expect(help).toContain(token);
    }
    expect(help).toContain('2 cards at four seats, 3 at three');
  });
});
