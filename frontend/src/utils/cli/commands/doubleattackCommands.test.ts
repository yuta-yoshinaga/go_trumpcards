import { describe, expect, it } from 'vitest';
import { DOUBLEATTACK_CLI_HELP, parseDoubleAttackCommand } from './doubleattackCommands';

describe('parseDoubleAttackCommand', () => {
  it('bet はアンティと任意の Bust It を渡す', () => {
    expect(parseDoubleAttackCommand('bet 50 20')).toEqual({ args: ['bet', { ante: 50, bustIt: 20 }] });
    expect(parseDoubleAttackCommand('bet 50')).toEqual({ args: ['bet', { ante: 50, bustIt: 0 }] });
    expect(parseDoubleAttackCommand('b 10')).toEqual({ args: ['bet', { ante: 10, bustIt: 0 }] });
  });

  it('壊れた bet は使い方を返す', () => {
    expect(parseDoubleAttackCommand('bet')).toHaveProperty('error');
    expect(parseDoubleAttackCommand('bet xyz')).toHaveProperty('error');
    expect(parseDoubleAttackCommand('bet 50 xyz')).toHaveProperty('error');
  });

  // **attack 0 は「見送り」で、正当な入力。**
  it('attack 0 を見送りとして通す', () => {
    expect(parseDoubleAttackCommand('attack 0')).toEqual({ args: ['attack', { amount: 0 }] });
    expect(parseDoubleAttackCommand('a 50')).toEqual({ args: ['attack', { amount: 50 }] });
    // 上限はサーバが持つので、クライアントでは丸めない。
    expect(parseDoubleAttackCommand('attack 99999')).toEqual({ args: ['attack', { amount: 99999 }] });
  });

  it('金額の無い attack は使い方を返す', () => {
    expect(parseDoubleAttackCommand('attack')).toHaveProperty('error');
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseDoubleAttackCommand('hit')).toEqual({ args: ['hit'] });
    expect(parseDoubleAttackCommand('h')).toEqual({ args: ['hit'] });
    expect(parseDoubleAttackCommand('stand')).toEqual({ args: ['stand'] });
    expect(parseDoubleAttackCommand('s')).toEqual({ args: ['stand'] });
    expect(parseDoubleAttackCommand('double')).toEqual({ args: ['double'] });
    expect(parseDoubleAttackCommand('d')).toEqual({ args: ['double'] });
    expect(parseDoubleAttackCommand('split')).toEqual({ args: ['split'] });
    expect(parseDoubleAttackCommand('sp')).toEqual({ args: ['split'] });
    expect(parseDoubleAttackCommand('next')).toEqual({ args: ['next'] });
    expect(parseDoubleAttackCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseDoubleAttackCommand('log')).toEqual({ args: ['log'] });
    expect(parseDoubleAttackCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseDoubleAttackCommand('r')).toEqual({ args: ['reset'] });
  });

  it('未知のコマンドは候補を添えて返す', () => {
    const result = parseDoubleAttackCommand('attac');
    expect(result).toHaveProperty('error');
    expect((result as { error: string }).error).toContain('attack');
    expect(parseDoubleAttackCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプに主要コマンドが載っている', () => {
    const help = DOUBLEATTACK_CLI_HELP.join('\n');
    for (const cmd of ['bet', 'attack', 'hit', 'stand', 'double', 'split']) {
      expect(help).toContain(cmd);
    }
  });
});
