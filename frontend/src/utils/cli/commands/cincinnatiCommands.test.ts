import { describe, expect, it } from 'vitest';
import { CINCINNATI_CLI_HELP, parseCincinnatiCommand } from './cincinnatiCommands';

describe('parseCincinnatiCommand', () => {
  // **額が要らない手は引数を取らない。**
  it('額なしの手を解釈する', () => {
    expect(parseCincinnatiCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseCincinnatiCommand('f')).toEqual({ args: ['fold'] });
    expect(parseCincinnatiCommand('check')).toEqual({ args: ['check'] });
    expect(parseCincinnatiCommand('k')).toEqual({ args: ['check'] });
    expect(parseCincinnatiCommand('call')).toEqual({ args: ['call'] });
    expect(parseCincinnatiCommand('c')).toEqual({ args: ['call'] });
  });

  it('額が要る手を解釈する', () => {
    expect(parseCincinnatiCommand('bet 20')).toEqual({ args: ['bet', { amount: 20 }] });
    expect(parseCincinnatiCommand('b 30')).toEqual({ args: ['bet', { amount: 30 }] });
    expect(parseCincinnatiCommand('raise 40')).toEqual({ args: ['raise', { amount: 40 }] });
  });

  it('額の無いベット / レイズを拒む', () => {
    expect(parseCincinnatiCommand('bet')).toHaveProperty('error');
    expect(parseCincinnatiCommand('raise')).toHaveProperty('error');
    expect(parseCincinnatiCommand('bet xyz')).toHaveProperty('error');
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseCincinnatiCommand('next')).toEqual({ args: ['next'] });
    expect(parseCincinnatiCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseCincinnatiCommand('log')).toEqual({ args: ['log'] });
    expect(parseCincinnatiCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseCincinnatiCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseCincinnatiCommand('chek');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('check');
    expect(parseCincinnatiCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = CINCINNATI_CLI_HELP.join('\n');
    for (const cmd of ['fold', 'check', 'call', 'bet', 'raise', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
