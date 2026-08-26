import { describe, expect, it } from 'vitest';
import { EIGHT_GAME_HELP, HORSE_HELP, parseHorseCommand } from './horseCommands';

describe('parseHorseCommand', () => {
  it.each([
    ['f', 'fold'],
    ['fold', 'fold'],
    ['x', 'check'],
    ['check', 'check'],
    ['c', 'call'],
    ['call', 'call'],
    ['allin', 'allin'],
  ])('parses %s', (input, action) => {
    expect(parseHorseCommand(input)).toEqual({ args: ['action', { action }] });
  });

  it('parses bet and raise with an amount', () => {
    expect(parseHorseCommand('b 50')).toEqual({ args: ['action', { action: 'bet', amount: 50 }] });
    expect(parseHorseCommand('raise 120')).toEqual({ args: ['action', { action: 'raise', amount: 120 }] });
  });

  // **額の無いベットは送らない。** 送ってもサーバに断られるだけで理由が残らない。
  it.each(['b', 'bet', 'raise', 'b x'])('rejects %s without a usable amount', (input) => {
    const result = parseHorseCommand(input);
    expect(result).toHaveProperty('error');
  });

  it('parses next, hint and reset', () => {
    expect(parseHorseCommand('n')).toEqual({ args: ['next'] });
    expect(parseHorseCommand('h')).toEqual({ args: ['hint'] });
    expect(parseHorseCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests a close command', () => {
    const result = parseHorseCommand('foldd');
    expect(result).toHaveProperty('error');
    expect('error' in result && result.error).toContain('fold');
  });

  it('rejects a command with nothing close to it', () => {
    expect(parseHorseCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  it('documents every action in the help text', () => {
    for (const token of ['fold', 'check', 'call', 'b <amount>', 'raise <amount>', 'allin', 'next']) {
      expect(HORSE_HELP.join('\n')).toContain(token);
    }
  });

  // **番号は 0 始まり。** 引き直しのある他のゲームと数え方を揃えないと、
  // 同じ `d 0 2` が 1 枚ずれた札を捨てる。
  it('parses a draw as zero-based card indices', () => {
    expect(parseHorseCommand('d 0 2')).toEqual({ args: ['draw', { cardIndices: [0, 2] }] });
    expect(parseHorseCommand('draw 4')).toEqual({ args: ['draw', { cardIndices: [4] }] });
  });

  // **引数無しの d はスタンドパットではない。** 読み替えると、番号を書き忘れた
  // 手が「引かない」として黙って通る。
  it.each(['d', 'draw', 'd x', 'd -1', 'd 1.5'])('rejects %s', (input) => {
    expect(parseHorseCommand(input)).toHaveProperty('error');
  });

  it('stands pat with an empty index list', () => {
    expect(parseHorseCommand('sp')).toEqual({ args: ['draw', { cardIndices: [] }] });
    expect(parseHorseCommand('stand')).toEqual({ args: ['draw', { cardIndices: [] }] });
  });

  // **引き直しの案内は 8 種目のほうにだけ。** H.O.R.S.E. にドロー系の種目は
  // 無いので、載せると打てない手を勧めることになる。
  it('documents the draw only in the Eight-Game help', () => {
    expect(EIGHT_GAME_HELP.join('\n')).toContain('d <idx>...');
    expect(EIGHT_GAME_HELP.join('\n')).toContain('sp / stand');
    expect(HORSE_HELP.join('\n')).not.toContain('d <idx>...');
    expect(HORSE_HELP.join('\n')).not.toContain('sp / stand');
  });
});
