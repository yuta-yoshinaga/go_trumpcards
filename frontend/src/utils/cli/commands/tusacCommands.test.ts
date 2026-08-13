import { describe, expect, it } from 'vitest';
import { parseTuSacCommand, TUSAC_CLI_HELP } from './tusacCommands';

describe('parseTuSacCommand', () => {
  // **山と捨て札は別のコマンド。** 引き先を引数にすると、付け忘れが
  // 「山から」に化けて、狙って拾った札が黙って流れる。
  it('引き先をコマンドで分ける', () => {
    expect(parseTuSacCommand('draw')).toEqual({ args: ['draw'] });
    expect(parseTuSacCommand('d')).toEqual({ args: ['draw'] });
    expect(parseTuSacCommand('take')).toEqual({ args: ['take'] });
    expect(parseTuSacCommand('t')).toEqual({ args: ['take'] });
  });

  // **画面は 1 始まり、ワイヤは 0 始まり。** ずれると別の札が動く。
  it('番号を 0 始まりに直して送る', () => {
    expect(parseTuSacCommand('meld 1 4 7')).toEqual({ args: ['meld', { indexes: [0, 3, 6] }] });
    expect(parseTuSacCommand('m 2 3')).toEqual({ args: ['meld', { indexes: [1, 2] }] });
    expect(parseTuSacCommand('discard 5')).toEqual({ args: ['discard', { index: 4 }] });
    expect(parseTuSacCommand('x 1')).toEqual({ args: ['discard', { index: 0 }] });
  });

  it('番号の無い meld / discard を拒む', () => {
    expect(parseTuSacCommand('meld')).toHaveProperty('error');
    expect(parseTuSacCommand('discard')).toHaveProperty('error');
    expect(parseTuSacCommand('meld abc')).toHaveProperty('error');
    expect(parseTuSacCommand('discard xyz')).toHaveProperty('error');
  });

  // **0 番は受け付けない。** 画面に 0 番の札は無い。
  it('0 番を拒む', () => {
    expect(parseTuSacCommand('meld 0')).toHaveProperty('error');
    expect(parseTuSacCommand('discard 0')).toHaveProperty('error');
    expect(parseTuSacCommand('meld 1 0 3')).toHaveProperty('error');
  });

  it('残りのコマンドを解釈する', () => {
    expect(parseTuSacCommand('next')).toEqual({ args: ['next'] });
    expect(parseTuSacCommand('hint')).toEqual({ args: ['hint'] });
    expect(parseTuSacCommand('log')).toEqual({ args: ['log'] });
    expect(parseTuSacCommand('reset')).toEqual({ args: ['reset'] });
    expect(parseTuSacCommand('r')).toEqual({ args: ['reset'] });
  });

  it('綴り違いを提案し、遠い語は素で拒む', () => {
    const result = parseTuSacCommand('mled');
    expect(result).toHaveProperty('error');
    if ('error' in result) expect(result.error).toContain('meld');
    expect(parseTuSacCommand('zzzzz')).toHaveProperty('error');
  });

  it('ヘルプが全コマンドに触れている', () => {
    const help = TUSAC_CLI_HELP.join('\n');
    for (const cmd of ['draw', 'take', 'meld', 'discard', 'next', 'hint', 'log', 'reset']) {
      expect(help).toContain(cmd);
    }
  });
});
