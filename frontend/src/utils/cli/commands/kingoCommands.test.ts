import { describe, expect, it } from 'vitest';
import { KINGO_CLI_HELP, parseKingoCommand } from './kingoCommands';

describe('parseKingoCommand', () => {
  it('張りを解釈する', () => {
    expect(parseKingoCommand('bet 20')).toEqual({ args: ['bet', { amount: 20 }] });
    expect(parseKingoCommand('b 50')).toEqual({ args: ['bet', { amount: 50 }] });
  });

  it('額の無い張りを拒む', () => {
    expect(parseKingoCommand('bet')).toHaveProperty('error');
    expect(parseKingoCommand('bet xyz')).toHaveProperty('error');
  });

  // **張ると配るは別のコマンド。** 親と子で求められる手が違う。
  it('配るを別のコマンドとして解釈する', () => {
    expect(parseKingoCommand('deal')).toEqual({ args: ['deal'] });
    expect(parseKingoCommand('d')).toEqual({ args: ['deal'] });
    // 配るのに額は要らない。
    expect(parseKingoCommand('deal 20')).toEqual({ args: ['deal'] });
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseKingoCommand('next')).toEqual({ args: ['next'] });
    expect(parseKingoCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseKingoCommand('log')).toEqual({ args: ['log'] });
    expect(parseKingoCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseKingoCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseKingoCommand('dael');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('deal');
    expect(parseKingoCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = KINGO_CLI_HELP.join('\n');
    for (const cmd of ['bet', 'deal', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
