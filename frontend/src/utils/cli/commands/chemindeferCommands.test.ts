import { describe, expect, it } from 'vitest';
import { CHEMINDEFER_CLI_HELP, parseChemindeFerCommand } from './chemindeferCommands';

describe('parseChemindeFerCommand', () => {
  it('stake は金額を渡す', () => {
    expect(parseChemindeFerCommand('stake 250')).toEqual({ args: ['stake', { stake: 250 }] });
    expect(parseChemindeFerCommand('s 10')).toEqual({ args: ['stake', { stake: 10 }] });
  });

  it('金額の無い stake は使い方を返す', () => {
    expect(parseChemindeFerCommand('stake')).toHaveProperty('error');
    expect(parseChemindeFerCommand('stake abc')).toHaveProperty('error');
  });

  // **bet 0 は「降りる」で、正当な入力。** エラーにしてはいけない。
  it('bet 0 は降りるとして通す', () => {
    expect(parseChemindeFerCommand('bet 0')).toEqual({ args: ['bet', { amount: 0 }] });
    expect(parseChemindeFerCommand('b 75')).toEqual({ args: ['bet', { amount: 75 }] });
  });

  it('金額の無い bet は使い方を返す', () => {
    expect(parseChemindeFerCommand('bet')).toHaveProperty('error');
  });

  // **側を書けばその側へ。** 書かなければサーバがフェーズから決める。
  it('draw / stand は側の指定を尊重する', () => {
    expect(parseChemindeFerCommand('draw punter')).toEqual({ args: ['pd'] });
    expect(parseChemindeFerCommand('draw p')).toEqual({ args: ['pd'] });
    expect(parseChemindeFerCommand('stand punter')).toEqual({ args: ['ps'] });
    expect(parseChemindeFerCommand('draw banker')).toEqual({ args: ['bd'] });
    expect(parseChemindeFerCommand('draw b')).toEqual({ args: ['bd'] });
    expect(parseChemindeFerCommand('stand banker')).toEqual({ args: ['bs'] });
  });

  it('側を省くとフェーズ解決のコマンドを送る', () => {
    expect(parseChemindeFerCommand('draw')).toEqual({ args: ['d'] });
    expect(parseChemindeFerCommand('stand')).toEqual({ args: ['st'] });
  });

  it('知らない側は使い方を返す', () => {
    expect(parseChemindeFerCommand('draw dealer')).toHaveProperty('error');
    expect(parseChemindeFerCommand('stand xyz')).toHaveProperty('error');
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseChemindeFerCommand('pass')).toEqual({ args: ['pb'] });
    expect(parseChemindeFerCommand('next')).toEqual({ args: ['next'] });
    expect(parseChemindeFerCommand('giveup')).toEqual({ args: ['giveup'] });
    expect(parseChemindeFerCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseChemindeFerCommand('h')).toEqual({ args: ['hint'] });
    expect(parseChemindeFerCommand('log')).toEqual({ args: ['log'] });
    expect(parseChemindeFerCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseChemindeFerCommand('r')).toEqual({ args: ['reset'] });
  });

  it('未知のコマンドは候補を添えて返す', () => {
    const result = parseChemindeFerCommand('stak');
    expect(result).toHaveProperty('error');
    expect((result as { error: string }).error).toContain('stake');

    expect(parseChemindeFerCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプに主要コマンドが載っている', () => {
    const help = CHEMINDEFER_CLI_HELP.join('\n');
    for (const cmd of ['stake', 'bet', 'draw', 'stand', 'pass', 'next']) {
      expect(help).toContain(cmd);
    }
  });
});
