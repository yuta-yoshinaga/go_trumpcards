import { describe, expect, it } from 'vitest';
import { DEHLA_PAKAD_HELP, parseDehlaPakadCommand } from './dehlaPakadCommands';

describe('parseDehlaPakadCommand', () => {
  it.each([
    ['t 1', 1],
    ['trump 4', 4],
  ])('%s calls the trump', (input, trumpSuit) => {
    expect(parseDehlaPakadCommand(input)).toEqual({ args: ['trump', { trumpSuit }] });
  });

  it.each([
    ['p 0', 0],
    ['play 12', 12],
  ])('%s plays a card', (input, cardIndex) => {
    expect(parseDehlaPakadCommand(input)).toEqual({ args: ['play', { cardIndex }] });
  });

  it('rejects a trump or play with no number', () => {
    expect(parseDehlaPakadCommand('t')).toEqual({ error: 'Usage: t <1-4>' });
    expect(parseDehlaPakadCommand('play x')).toEqual({ error: 'Usage: p <idx>' });
  });

  it.each([
    ['nh', 'nexthand'],
    ['nexthand', 'nexthand'],
    ['h', 'hint'],
    ['hint', 'hint'],
    ['r', 'reset'],
    ['reset', 'reset'],
  ])('%s maps to %s', (input, command) => {
    expect(parseDehlaPakadCommand(input)).toEqual({ args: [command] });
  });

  it('suggests a near miss', () => {
    expect(parseDehlaPakadCommand('trmp 1')).toEqual({
      error: 'Unknown command: trmp. Did you mean: trump?',
    });
  });

  it('reports an unknown command with nothing close to it', () => {
    expect(parseDehlaPakadCommand('zzzz')).toEqual({ error: 'Unknown command: zzzz' });
  });

  // ヘルプが知らないコマンドを宣伝していないこと。
  it('advertises only commands the parser accepts', () => {
    for (const line of DEHLA_PAKAD_HELP) {
      const first = line.split(/[\s/]/)[0];
      expect(parseDehlaPakadCommand(`${first} 1`)).not.toHaveProperty('error');
    }
  });
});
