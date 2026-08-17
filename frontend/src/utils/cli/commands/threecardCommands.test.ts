import { describe, expect, it } from 'vitest';
import { parseThreecardCommand, THREECARD_HELP } from './threecardCommands';

describe('parseThreecardCommand', () => {
  it('parses bet with amount', () => {
    expect(parseThreecardCommand('b 100')).toEqual({ args: ['bet', 100] });
    expect(parseThreecardCommand('bet 50')).toEqual({ args: ['bet', 50] });
  });

  it('parses bet with amount and pair plus bet', () => {
    expect(parseThreecardCommand('b 100 50')).toEqual({ args: ['bet', 100, 50] });
  });

  it('returns error for bet without amount', () => {
    const result = parseThreecardCommand('b');
    expect('error' in result).toBe(true);
  });

  it('parses play', () => {
    expect(parseThreecardCommand('p')).toEqual({ args: ['play'] });
    expect(parseThreecardCommand('play')).toEqual({ args: ['play'] });
  });

  it('parses fold', () => {
    expect(parseThreecardCommand('f')).toEqual({ args: ['fold'] });
    expect(parseThreecardCommand('fold')).toEqual({ args: ['fold'] });
  });

  it('parses log', () => {
    expect(parseThreecardCommand('log')).toEqual({ args: ['log'] });
  });

  it('parses reset', () => {
    expect(parseThreecardCommand('r')).toEqual({ args: ['reset'] });
    expect(parseThreecardCommand('reset')).toEqual({ args: ['reset'] });
  });

  it('returns error for unknown command', () => {
    const result = parseThreecardCommand('xyz');
    expect('error' in result).toBe(true);
  });
});

// #5513: ボタン (tc-rebet-button) はワンクリックで直前と同額を賭け直せるのに、
// CLI には同等のコマンドが無く毎ラウンド bet <ante> <pairPlus> を手打ちしていた。
describe('parseThreecardCommand rebet', () => {
  it.each(['rb', 'rebet', 'REBET'])('accepts %j', (input) => {
    expect(parseThreecardCommand(input)).toEqual({ args: ['rebet'] });
  });

  // **金額は送らない。** サーバが直前の額を覚えているので、クライアントが
  // 別に持つと2つの記憶がずれる。
  it('sends no amount, leaving the server as the source of truth', () => {
    const result = parseThreecardCommand('rebet');
    expect('args' in result && result.args).toHaveLength(1);
  });

  it('documents rebet in the help text', () => {
    expect(THREECARD_HELP.some((l) => /rebet/.test(l))).toBe(true);
  });
});
