import { describe, expect, it } from 'vitest';
import { BASEBALLPOKER_CLI_HELP, parseBaseballPokerCommand } from './baseballpokerCommands';

describe('parseBaseballPokerCommand', () => {
  it('額なしの手を解釈する', () => {
    expect(parseBaseballPokerCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseBaseballPokerCommand('f')).toEqual({ args: ['fold'] });
    expect(parseBaseballPokerCommand('check')).toEqual({ args: ['check'] });
    expect(parseBaseballPokerCommand('k')).toEqual({ args: ['check'] });
    expect(parseBaseballPokerCommand('call')).toEqual({ args: ['call'] });
    expect(parseBaseballPokerCommand('c')).toEqual({ args: ['call'] });
  });

  it('額が要る手を解釈する', () => {
    expect(parseBaseballPokerCommand('bet 20')).toEqual({ args: ['bet', { amount: 20 }] });
    expect(parseBaseballPokerCommand('b 30')).toEqual({ args: ['bet', { amount: 30 }] });
    expect(parseBaseballPokerCommand('raise 40')).toEqual({ args: ['raise', { amount: 40 }] });
  });

  it('額の無いベット / レイズを拒む', () => {
    expect(parseBaseballPokerCommand('bet')).toHaveProperty('error');
    expect(parseBaseballPokerCommand('raise')).toHaveProperty('error');
    expect(parseBaseballPokerCommand('bet xyz')).toHaveProperty('error');
  });

  // **買い増しの返事は別々のコマンド。** 数値を添えさせると、添え忘れが
  // 「0 番の返事」= 支払いに化ける。
  it('買い増しの返事を別々のコマンドに解釈する', () => {
    expect(parseBaseballPokerCommand('pay')).toEqual({ args: ['pay'] });
    expect(parseBaseballPokerCommand('p')).toEqual({ args: ['pay'] });
    expect(parseBaseballPokerCommand('buyfold')).toEqual({ args: ['buyfold'] });
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseBaseballPokerCommand('next')).toEqual({ args: ['next'] });
    expect(parseBaseballPokerCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseBaseballPokerCommand('log')).toEqual({ args: ['log'] });
    expect(parseBaseballPokerCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseBaseballPokerCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseBaseballPokerCommand('bufold');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('buyfold');
    expect(parseBaseballPokerCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = BASEBALLPOKER_CLI_HELP.join('\n');
    for (const cmd of ['fold', 'check', 'call', 'bet', 'raise', 'pay', 'buyfold', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
