import { describe, expect, it } from 'vitest';
import { CRAZYFOURPOKER_CLI_HELP, parseCrazyFourPokerCommand } from './crazyfourpokerCommands';

describe('parseCrazyFourPokerCommand', () => {
  // **Queens Up は省略できる。** 省略は「置かない」= 0。
  it('bet はアンティと任意の Queens Up を渡す', () => {
    expect(parseCrazyFourPokerCommand('bet 50 20')).toEqual({
      args: ['bet', { ante: 50, queensUp: 20 }],
    });
    expect(parseCrazyFourPokerCommand('bet 50')).toEqual({
      args: ['bet', { ante: 50, queensUp: 0 }],
    });
    expect(parseCrazyFourPokerCommand('b 10')).toEqual({
      args: ['bet', { ante: 10, queensUp: 0 }],
    });
  });

  it('壊れた bet は使い方を返す', () => {
    expect(parseCrazyFourPokerCommand('bet')).toHaveProperty('error');
    expect(parseCrazyFourPokerCommand('bet xyz')).toHaveProperty('error');
    expect(parseCrazyFourPokerCommand('bet 50 xyz')).toHaveProperty('error');
  });

  // **上限はクライアントで判定しない。** 手役次第なのでサーバが弾く。
  it('play は倍率をそのまま渡す', () => {
    expect(parseCrazyFourPokerCommand('play 1')).toEqual({ args: ['play', { multiplier: 1 }] });
    expect(parseCrazyFourPokerCommand('play 3')).toEqual({ args: ['play', { multiplier: 3 }] });
    expect(parseCrazyFourPokerCommand('p 2')).toEqual({ args: ['play', { multiplier: 2 }] });
    // 明らかに過大な値もクライアントでは丸めない (サーバの拒否を見せるため)。
    expect(parseCrazyFourPokerCommand('play 99')).toEqual({ args: ['play', { multiplier: 99 }] });
  });

  it('倍率の無い play は使い方を返す', () => {
    expect(parseCrazyFourPokerCommand('play')).toHaveProperty('error');
    expect(parseCrazyFourPokerCommand('play xyz')).toHaveProperty('error');
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseCrazyFourPokerCommand('fold')).toEqual({ args: ['fold'] });
    expect(parseCrazyFourPokerCommand('f')).toEqual({ args: ['fold'] });
    expect(parseCrazyFourPokerCommand('next')).toEqual({ args: ['next'] });
    expect(parseCrazyFourPokerCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseCrazyFourPokerCommand('h')).toEqual({ args: ['hint'] });
    expect(parseCrazyFourPokerCommand('log')).toEqual({ args: ['log'] });
    expect(parseCrazyFourPokerCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseCrazyFourPokerCommand('r')).toEqual({ args: ['reset'] });
  });

  it('未知のコマンドは候補を添えて返す', () => {
    const result = parseCrazyFourPokerCommand('bett');
    expect(result).toHaveProperty('error');
    expect((result as { error: string }).error).toContain('bet');
    expect(parseCrazyFourPokerCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプに主要コマンドが載っている', () => {
    const help = CRAZYFOURPOKER_CLI_HELP.join('\n');
    for (const cmd of ['bet', 'play', 'fold', 'next']) {
      expect(help).toContain(cmd);
    }
  });
});
