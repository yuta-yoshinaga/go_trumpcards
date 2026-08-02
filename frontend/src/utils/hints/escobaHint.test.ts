import { describe, expect, it } from 'vitest';
import type { Card, EscobaResponse } from '../../types/card';
import { getEscobaHint } from './escobaHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[]; table?: Card[] };

function base({
  hand = [card('SPADE', 7), card('HEART', 3)],
  table = [card('CLOVER', 8), card('DIAMOND', 4)],
  ...overrides
}: Partial<EscobaResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        handCount: hand.length,
        cards: hand,
        capturedCount: 0,
        capturedCards: [],
        escobaCount: 0,
      },
      { id: 1, isHuman: false, handCount: 3, cards: [], capturedCount: 0, capturedCards: [], escobaCount: 0 },
    ],
    tableCards: table,
    phase: 'playerTurn',
    roundNumber: 1,
    currentTurn: 0,
    dealerIdx: 1,
    stockRemaining: 20,
    lastCaptureIdx: -1,
    winnerIdx: -1,
    gameEndFlag: false,
    isHumanTurn: true,
    handCaptures: [[], []],
    message: '',
    config: { cpuDifficulty: 1, targetScore: 21 },
    ...overrides,
  } as EscobaResponse;
}

describe('getEscobaHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getEscobaHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet between rounds', () => {
    expect(getEscobaHint(base({ phase: 'roundEnd' }))).toBeNull();
  });

  it('stays quiet when it is not the human turn', () => {
    expect(getEscobaHint(base({ isHumanTurn: false }))).toBeNull();
  });

  // **場を全部さらうとエスコバで 1 点。**他のどの取り方より優先する。
  it('takes the capture that clears the table', () => {
    const s = base({
      table: [card('CLOVER', 8), card('DIAMOND', 4)],
      handCaptures: [[[0]], [[0, 1]]],
    });
    expect(getEscobaHint(s)).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.escobaEscoba',
      confidence: 'strong',
    });
  });

  // 場が空にならないなら、枚数を多く取れる手を選ぶ。
  it('prefers the capture that takes the most cards', () => {
    const s = base({
      table: [card('CLOVER', 8), card('DIAMOND', 4), card('SPADE', 2)],
      handCaptures: [[[0]], [[1, 2]]],
    });
    expect(getEscobaHint(s)).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.escobaCaptureMost',
      confidence: 'moderate',
    });
  });

  // **札 0 も取り手になりうる。**真偽値で見ると先頭だけ落ちる。
  it('keeps a capture on card index 0', () => {
    const s = base({ handCaptures: [[[0, 1]], []] });
    expect(getEscobaHint(s)?.targetAction).toBe('card-0');
  });

  // 取れないときは一番小さい札を置く。相手に 15 を作らせにくい。
  it('lays off the lowest card when nothing captures', () => {
    const hand = [card('SPADE', 7), card('HEART', 3)];
    expect(getEscobaHint(base({ hand, handCaptures: [[], []] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.escobaLayLow',
      confidence: 'moderate',
    });
  });

  // エスコバが 2 つあるときは、多く取れるほうを選ぶ。
  it('prefers the larger of two table-clearing captures', () => {
    const s = base({
      table: [card('CLOVER', 8), card('DIAMOND', 4)],
      handCaptures: [[[0, 1]], [[0, 1]]],
    });
    expect(getEscobaHint(s)?.targetAction).toBe('card-0');
  });

  // **場が空なら「全部さらう」は成立しない。**tableSize 0 をエスコバにしない。
  it('does not call an empty table an escoba', () => {
    const s = base({ table: [], handCaptures: [[], []] });
    expect(getEscobaHint(s)?.reason).toBe('frontendHint.escobaLayLow');
  });

  // 一番小さい札が先頭でない場合も選べる（走査の後ろ側の分岐）。
  it('finds the lowest card later in the hand', () => {
    const hand = [card('SPADE', 7), card('HEART', 5), card('CLOVER', 2)];
    expect(getEscobaHint(base({ hand, handCaptures: [[], [], []] }))?.targetAction).toBe('card-2');
  });

  // 先頭が既に一番小さい場合（走査で一度も更新されない側）。
  it('keeps the first card when it is already the lowest', () => {
    const hand = [card('CLOVER', 2), card('SPADE', 7), card('HEART', 5)];
    expect(getEscobaHint(base({ hand, handCaptures: [[], [], []] }))?.targetAction).toBe('card-0');
  });

  it('stays quiet without a visible hand', () => {
    expect(getEscobaHint(base({ hand: [], handCaptures: [] }))).toBeNull();
  });

  // handCaptures は手札と同じ長さで届く。短ければ読める範囲だけ見る。
  it('survives a capture list shorter than the hand', () => {
    const hand = [card('SPADE', 7), card('HEART', 3), card('CLOVER', 2)];
    expect(getEscobaHint(base({ hand, handCaptures: [[[0]]] }))?.targetAction).toBe('card-0');
  });
});
