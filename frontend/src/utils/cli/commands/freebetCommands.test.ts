import { describe, expect, it } from 'vitest';
import { FREEBET_CLI_HELP, parseFreeBetCommand } from './freebetCommands';

describe('parseFreeBetCommand', () => {
  it('アンティを解釈する', () => {
    expect(parseFreeBetCommand('bet 50')).toEqual({ args: ['bet', { ante: 50 }] });
    expect(parseFreeBetCommand('b 10')).toEqual({ args: ['bet', { ante: 10 }] });
  });

  it('額が無い / 数でないアンティを拒む', () => {
    expect(parseFreeBetCommand('bet')).toHaveProperty('error');
    expect(parseFreeBetCommand('bet xyz')).toHaveProperty('error');
  });

  // **無料の操作に額は無い。** 余計な引数は無視され、コマンドだけが送られる。
  it('無料ダブル / 無料スプリットは引数を取らない', () => {
    expect(parseFreeBetCommand('freedouble')).toEqual({ args: ['freedouble'] });
    expect(parseFreeBetCommand('fd')).toEqual({ args: ['freedouble'] });
    expect(parseFreeBetCommand('freesplit')).toEqual({ args: ['freesplit'] });
    expect(parseFreeBetCommand('fs')).toEqual({ args: ['freesplit'] });
    expect(parseFreeBetCommand('fd 100')).toEqual({ args: ['freedouble'] });
  });

  it('残りのコマンドと短縮形を解釈する', () => {
    expect(parseFreeBetCommand('hit')).toEqual({ args: ['hit'] });
    expect(parseFreeBetCommand('h')).toEqual({ args: ['hit'] });
    expect(parseFreeBetCommand('stand')).toEqual({ args: ['stand'] });
    expect(parseFreeBetCommand('s')).toEqual({ args: ['stand'] });
    expect(parseFreeBetCommand('next')).toEqual({ args: ['next'] });
    expect(parseFreeBetCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseFreeBetCommand('log')).toEqual({ args: ['log'] });
    expect(parseFreeBetCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseFreeBetCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseFreeBetCommand('freedoubl');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('freedouble');
    expect(parseFreeBetCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = FREEBET_CLI_HELP.join('\n');
    for (const cmd of ['bet', 'hit', 'stand', 'freedouble', 'freesplit', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
