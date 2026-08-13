import { describe, expect, it } from 'vitest';
import { IRONCROSS_CLI_HELP, parseIronCrossCommand } from './ironcrossCommands';

describe('parseIronCrossCommand', () => {
  // **額が要らない手は引数を取らない。**
  it('額なしの手を解釈する', () => {
    expect(parseIronCrossCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseIronCrossCommand('f')).toEqual({ args: ['fold'] });
    expect(parseIronCrossCommand('check')).toEqual({ args: ['check'] });
    expect(parseIronCrossCommand('k')).toEqual({ args: ['check'] });
    expect(parseIronCrossCommand('call')).toEqual({ args: ['call'] });
    expect(parseIronCrossCommand('c')).toEqual({ args: ['call'] });
  });

  it('額が要る手を解釈する', () => {
    expect(parseIronCrossCommand('bet 20')).toEqual({ args: ['bet', { amount: 20 }] });
    expect(parseIronCrossCommand('b 30')).toEqual({ args: ['bet', { amount: 30 }] });
    expect(parseIronCrossCommand('raise 40')).toEqual({ args: ['raise', { amount: 40 }] });
  });

  it('額の無いベット / レイズを拒む', () => {
    expect(parseIronCrossCommand('bet')).toHaveProperty('error');
    expect(parseIronCrossCommand('raise')).toHaveProperty('error');
    expect(parseIronCrossCommand('bet xyz')).toHaveProperty('error');
  });

  // **列は名前で送る。** 打ち間違いで一度きりの選択が潰れないように、
  // 縦と横が別々のコマンドになっていることを確かめる。
  it('縦と横をそれぞれのコマンドに解釈する', () => {
    expect(parseIronCrossCommand('vertical')).toEqual({ args: ['vertical'] });
    expect(parseIronCrossCommand('v')).toEqual({ args: ['vertical'] });
    expect(parseIronCrossCommand('horizontal')).toEqual({ args: ['horizontal'] });
    expect(parseIronCrossCommand('h')).toEqual({ args: ['horizontal'] });
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseIronCrossCommand('next')).toEqual({ args: ['next'] });
    expect(parseIronCrossCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseIronCrossCommand('log')).toEqual({ args: ['log'] });
    expect(parseIronCrossCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseIronCrossCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseIronCrossCommand('vertcal');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('vertical');
    expect(parseIronCrossCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = IRONCROSS_CLI_HELP.join('\n');
    for (const cmd of [
      'fold',
      'check',
      'call',
      'bet',
      'raise',
      'vertical',
      'horizontal',
      'next',
      'hint',
      'log',
      'reset',
    ]) {
      expect(help).toContain(cmd);
    }
  });
});
