import { describe, expect, it } from 'vitest';
import { BANLUCK_CLI_HELP, parseBanLuckCommand } from './banluckCommands';

describe('parseBanLuckCommand', () => {
  it('賭け金を解釈する', () => {
    expect(parseBanLuckCommand('bet 50')).toEqual({ args: ['bet', { bet: 50 }] });
    expect(parseBanLuckCommand('b 10')).toEqual({ args: ['bet', { bet: 10 }] });
  });

  // **0 は「親なので賭けない」で、正当な入力。**
  it('賭け金 0 を通す', () => {
    expect(parseBanLuckCommand('bet 0')).toEqual({ args: ['bet', { bet: 0 }] });
  });

  it('額が無い / 数でない賭け金を拒む', () => {
    expect(parseBanLuckCommand('bet')).toHaveProperty('error');
    expect(parseBanLuckCommand('bet xyz')).toHaveProperty('error');
  });

  it('残りのコマンドと短縮形を解釈する', () => {
    expect(parseBanLuckCommand('hit')).toEqual({ args: ['hit'] });
    expect(parseBanLuckCommand('h')).toEqual({ args: ['hit'] });
    expect(parseBanLuckCommand('stand')).toEqual({ args: ['stand'] });
    expect(parseBanLuckCommand('s')).toEqual({ args: ['stand'] });
    expect(parseBanLuckCommand('next')).toEqual({ args: ['next'] });
    expect(parseBanLuckCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseBanLuckCommand('log')).toEqual({ args: ['log'] });
    expect(parseBanLuckCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseBanLuckCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseBanLuckCommand('stnd');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('stand');
    expect(parseBanLuckCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = BANLUCK_CLI_HELP.join('\n');
    for (const cmd of ['bet', 'hit', 'stand', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
