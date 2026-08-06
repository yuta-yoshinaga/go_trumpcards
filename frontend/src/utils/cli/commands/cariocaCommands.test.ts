import { describe, expect, it } from 'vitest';
import { CARIOCA_HELP, parseCariocaCommand } from './cariocaCommands';

describe('parseCariocaCommand', () => {
  it('parses the two draw commands', () => {
    expect(parseCariocaCommand('ds')).toEqual({ args: ['drawstock'] });
    expect(parseCariocaCommand('drawstock')).toEqual({ args: ['drawstock'] });
    expect(parseCariocaCommand('dd')).toEqual({ args: ['drawdiscard'] });
    expect(parseCariocaCommand('drawdiscard')).toEqual({ args: ['drawdiscard'] });
  });

  // 契約は「スロットごとに 1 グループ」。CUI の parseSlotIndices と同じ形。
  it('parses meldcontract as one group per slot', () => {
    expect(parseCariocaCommand('mc 0,1,2 3,4,5')).toEqual({
      args: [
        'meldcontract',
        {
          indicesPerSlot: [
            [0, 1, 2],
            [3, 4, 5],
          ],
        },
      ],
    });
    expect(parseCariocaCommand('meldcontract 0,1,2')).toEqual({
      args: ['meldcontract', { indicesPerSlot: [[0, 1, 2]] }],
    });
  });

  it('rejects a malformed contract group', () => {
    expect(parseCariocaCommand('mc')).toEqual({ error: 'Usage: mc <a,b,c> <d,e,f>' });
    expect(parseCariocaCommand('mc 0,x,2')).toEqual({ error: 'Usage: mc <a,b,c> <d,e,f>' });
    expect(parseCariocaCommand('mc 0,,2')).toEqual({ error: 'Usage: mc <a,b,c> <d,e,f>' });
  });

  it('parses meldextra', () => {
    expect(parseCariocaCommand('me 0 1 2')).toEqual({ args: ['meldextra', { cardIndices: [0, 1, 2] }] });
    expect(parseCariocaCommand('me')).toEqual({ error: 'Usage: me <idx...>' });
  });

  it('parses layoff as player/meld/card', () => {
    expect(parseCariocaCommand('lo 1 0 3')).toEqual({
      args: ['layoff', { targetPlayerIdx: 1, meldIdx: 0, cardIndex: 3 }],
    });
    // 3 つ揃っていなければエラー (CUI と同じ)。
    expect(parseCariocaCommand('lo 1 0')).toEqual({ error: 'Usage: lo <playerIdx> <meldIdx> <cardIdx>' });
  });

  it('parses discard', () => {
    expect(parseCariocaCommand('d 2')).toEqual({ args: ['discard', { cardIndex: 2 }] });
    expect(parseCariocaCommand('discard')).toEqual({ error: 'Usage: d <idx>' });
  });

  it('parses the round and log commands', () => {
    expect(parseCariocaCommand('nr')).toEqual({ args: ['nextround'] });
    expect(parseCariocaCommand('log')).toEqual({ args: ['log'] });
    expect(parseCariocaCommand('r')).toEqual({ args: ['reset'] });
  });

  it('suggests the nearest command for a typo', () => {
    expect(parseCariocaCommand('drawstok')).toEqual({
      error: 'Unknown command: drawstok. Did you mean: drawstock?',
    });
  });

  // ヘルプに載せただけで動かない行を防ぐ。「不明なコマンド」にならないことだけ見る
  // (引数の形はコマンドごとに違うので、ここでは形は問わない)。
  it('documents only commands it knows', () => {
    for (const line of CARIOCA_HELP) {
      const first = line.split(/[\s/]/)[0];
      const result = parseCariocaCommand(first);
      if ('error' in result) {
        expect(result.error).not.toMatch(/Unknown command/);
      }
    }
  });
});
