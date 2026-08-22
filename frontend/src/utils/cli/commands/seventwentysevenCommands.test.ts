import { describe, expect, it } from 'vitest';
import { parseSevenTwentySevenCommand, SEVENTWENTYSEVEN_HELP } from './seventwentysevenCommands';

describe('parseSevenTwentySevenCommand', () => {
  // **`card` は引く、`stand` は止まる。** どちらも追加パラメータを取らない。
  it.each([
    ['c', 'card'],
    ['card', 'card'],
    ['s', 'stand'],
    ['stand', 'stand'],
  ])('%s maps to %s with no parameter', (input, command) => {
    const r = parseSevenTwentySevenCommand(input);
    expect('args' in r && r.args).toEqual([command]);
  });

  // **Guts の in/out は受け付けない。** 残すと、打てるのに何も起きないコマンドになる。
  it.each(['i', 'in', 'o', 'out', 'declare'])('rejects the Guts command %s', (input) => {
    const r = parseSevenTwentySevenCommand(input);
    expect('error' in r).toBe(true);
  });

  it.each([
    ['n', 'nextround'],
    ['nr', 'nextround'],
    ['next', 'nextround'],
    ['nextround', 'nextround'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['l', 'log'],
    ['log', 'log'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('%s maps to %s', (input, command) => {
    const r = parseSevenTwentySevenCommand(input);
    expect('args' in r && r.args).toEqual([command]);
  });

  // **設定はリセット経由。** config は reset でしか受け付けられないので、
  // sp/sa/sc/st は reset にたたむ。
  it.each([
    ['sp 5', { playerCount: 5 }],
    ['setplayers 2', { playerCount: 2 }],
    ['sa 25', { ante: 25 }],
    ['setante 1', { ante: 1 }],
    ['sc 500', { startingChips: 500 }],
    ['setchips 10', { startingChips: 10 }],
    ['st 20', { targetRounds: 20 }],
    ['setrounds 1', { targetRounds: 1 }],
  ])('%s resets with the new config', (input, config) => {
    const r = parseSevenTwentySevenCommand(input);
    expect('args' in r && r.args).toEqual(['reset', config]);
  });

  // 範囲外は弾く。通してしまうとサーバ側で黙って丸められる。
  it.each(['sp 1', 'sp 8', 'sp x', 'sa 0', 'sa', 'sc 9', 'st 0'])('rejects out-of-range %s', (input) => {
    const r = parseSevenTwentySevenCommand(input);
    expect('error' in r).toBe(true);
  });

  it('suggests a near miss', () => {
    const r = parseSevenTwentySevenCommand('carrd');
    expect('error' in r && r.error).toContain('card');
  });

  it('reports an unknown command without a suggestion', () => {
    const r = parseSevenTwentySevenCommand('zzzz');
    expect('error' in r && r.error).toContain('zzzz');
  });

  // **ヘルプが載せるのは、パーサが知っているコマンドだけ。**
  // 引数の範囲エラー (sc は 10 以上) は「知らないコマンド」とは別物なので、
  // 見るのは "Unknown command" が返らないこと。
  it('documents only commands the parser knows', () => {
    for (const line of SEVENTWENTYSEVEN_HELP) {
      const token = line.trim().split(/[\s/]/)[0];
      const r = parseSevenTwentySevenCommand(`${token} 3`);
      const err = 'error' in r ? r.error : '';
      expect(err).not.toContain('Unknown command');
    }
    // **負のコントロール**: 知らないコマンドなら本当に Unknown が返ること。
    const unknown = parseSevenTwentySevenCommand('zzzz 3');
    expect('error' in unknown && unknown.error).toContain('Unknown command');
  });

  // Guts の in/out がヘルプに残っていないこと。
  it('does not advertise the Guts commands', () => {
    expect(SEVENTWENTYSEVEN_HELP.join('\n')).not.toMatch(/\bin\b|\bout\b/);
  });
});
