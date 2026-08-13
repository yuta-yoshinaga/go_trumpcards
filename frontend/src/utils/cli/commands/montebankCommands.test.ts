import { describe, expect, it } from 'vitest';
import { MONTEBANK_CLI_HELP, parseMonteBankCommand } from './montebankCommands';

describe('parseMonteBankCommand', () => {
  // **画面は 1 始まり、ワイヤは 0 始まり。** 変換はここ 1 か所だけ。
  it('場札の番号を0始まりに直して送る', () => {
    expect(parseMonteBankCommand('bet 1 50')).toEqual({ args: ['bet', { idx: 0, bet: 50 }] });
    expect(parseMonteBankCommand('bet 4 10')).toEqual({ args: ['bet', { idx: 3, bet: 10 }] });
    expect(parseMonteBankCommand('b 2 20')).toEqual({ args: ['bet', { idx: 1, bet: 20 }] });
  });

  it('範囲外の番号を拒む', () => {
    expect(parseMonteBankCommand('bet 0 50')).toHaveProperty('error');
    expect(parseMonteBankCommand('bet 5 50')).toHaveProperty('error');
    expect(parseMonteBankCommand('bet -1 50')).toHaveProperty('error');
  });

  it('引数が足りない / 数でない入力を拒む', () => {
    expect(parseMonteBankCommand('bet')).toHaveProperty('error');
    expect(parseMonteBankCommand('bet 1')).toHaveProperty('error');
    expect(parseMonteBankCommand('bet xyz 50')).toHaveProperty('error');
    expect(parseMonteBankCommand('bet 1 xyz')).toHaveProperty('error');
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseMonteBankCommand('next')).toEqual({ args: ['next'] });
    expect(parseMonteBankCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseMonteBankCommand('log')).toEqual({ args: ['log'] });
    expect(parseMonteBankCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseMonteBankCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseMonteBankCommand('nxt');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('next');
    expect(parseMonteBankCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = MONTEBANK_CLI_HELP.join('\n');
    for (const cmd of ['bet', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
