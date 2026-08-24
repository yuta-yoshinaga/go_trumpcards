import { describe, expect, it } from 'vitest';
import { parseQuodlibetCommand, QUODLIBET_HELP } from './quodlibetCommands';

describe('parseQuodlibetCommand', () => {
  it.each([
    ['c 0', 0],
    ['contract 11', 11],
  ])('%s chooses a contract', (input, contract) => {
    expect(parseQuodlibetCommand(input)).toEqual({ args: ['contract', { contract }] });
  });

  it.each([
    ['p 0', 0],
    ['play 5', 5],
  ])('%s plays a card', (input, cardIndex) => {
    expect(parseQuodlibetCommand(input)).toEqual({ args: ['play', { cardIndex }] });
  });

  it('rejects a contract or play with no number', () => {
    expect(parseQuodlibetCommand('c')).toEqual({ error: 'Usage: c <0-11>' });
    expect(parseQuodlibetCommand('play x')).toEqual({ error: 'Usage: p <idx>' });
  });

  it.each([
    ['pass', 'pass'],
    ['nd', 'nextdeal'],
    ['nextdeal', 'nextdeal'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('%s maps to %s', (input, command) => {
    expect(parseQuodlibetCommand(input)).toEqual({ args: [command] });
  });

  it('suggests a near miss', () => {
    expect(parseQuodlibetCommand('conract 1')).toEqual({
      error: 'Unknown command: conract. Did you mean: contract?',
    });
  });

  it('reports an unknown command with nothing close to it', () => {
    expect(parseQuodlibetCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが知らないコマンドを宣伝していないこと。
  it('advertises only commands the parser accepts', () => {
    for (const line of QUODLIBET_HELP) {
      const first = line.split(/[\s/]/)[0];
      expect(parseQuodlibetCommand(`${first} 0`)).not.toHaveProperty('error');
    }
  });
});
