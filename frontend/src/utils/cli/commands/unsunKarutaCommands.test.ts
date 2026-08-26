import { describe, expect, it } from 'vitest';
import { parseUnsunKarutaCommand, UNSUN_KARUTA_HELP } from './unsunKarutaCommands';

describe('parseUnsunKarutaCommand', () => {
  it.each([
    ['p 0', 0],
    ['play 3', 3],
  ])('%s plays a card without declaring', (input, cardIndex) => {
    expect(parseUnsunKarutaCommand(input)).toEqual({ args: ['play', { cardIndex, declare: false }] });
  });

  // **宣言は札と一緒に飛ぶ。** meri だけを送るコマンドは存在しない ── その盤面が
  // ドメインに無いため。
  it.each(['m 2', 'meri 2', 'monchi 2'])('%s plays the card and declares', (input) => {
    expect(parseUnsunKarutaCommand(input)).toEqual({ args: ['play', { cardIndex: 2, declare: true }] });
  });

  it('rejects a play without an index', () => {
    expect(parseUnsunKarutaCommand('p')).toEqual({ error: 'Usage: p <idx>' });
    expect(parseUnsunKarutaCommand('meri x')).toEqual({ error: 'Usage: meri <idx>' });
  });

  it.each([
    ['n', 'next'],
    ['next', 'next'],
    ['nr', 'nextround'],
    ['nextround', 'nextround'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('%s maps to %s', (input, command) => {
    expect(parseUnsunKarutaCommand(input)).toEqual({ args: [command] });
  });

  it('suggests a near miss', () => {
    expect(parseUnsunKarutaCommand('mri 1')).toEqual({
      error: 'Unknown command: mri. Did you mean: meri?',
    });
  });

  it('reports an unknown command with nothing close to it', () => {
    expect(parseUnsunKarutaCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが知らないコマンドを宣伝していないこと。
  it('advertises only commands the parser accepts', () => {
    for (const line of UNSUN_KARUTA_HELP) {
      const first = line.split(/[\s/]/)[0];
      expect(parseUnsunKarutaCommand(`${first} 0`)).not.toHaveProperty('error');
    }
  });
});
