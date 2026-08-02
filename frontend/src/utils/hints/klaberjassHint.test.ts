import { describe, expect, it } from 'vitest';
import type { Card, KlaberjassResponse } from '../../types/card';
import { KlaberjassPhase } from '../../types/phases';
import { getKlaberjassHint } from './klaberjassHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hand?: Card[] };

function base({ hand = [card('SPADE', 7), card('HEART', 9)], ...overrides }: Partial<KlaberjassResponse> & Extra = {}) {
  return {
    players: [
      { id: 0, isHuman: true, cardCount: hand.length, cards: hand, sequences: [] },
      { id: 1, isHuman: false, cardCount: 6, cards: [], sequences: [] },
    ],
    phase: KlaberjassPhase.PLAY,
    dealNumber: 1,
    currentPlayerIdx: 0,
    bidPlayerIdx: 0,
    dealerIdx: 1,
    trumpSuit: 1,
    turnUpCard: card('SPADE', 10),
    makerIdx: 0,
    trick: [],
    trickLeaderIdx: 0,
    trickNumber: 1,
    validPlays: [0, 1],
    sequenceWinner: -1,
    belaHolder: -1,
    belaScored: false,
    dixUsed: false,
    bete: false,
    schmeissBy: -1,
    targetScore: 501,
    gameEndFlag: false,
    winnerIdx: -1,
    message: '',
    ...overrides,
  } as KlaberjassResponse;
}

describe('getKlaberjassHint', () => {
  it('stays quiet once the game is over', () => {
    expect(getKlaberjassHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet when the opponent is on turn', () => {
    expect(getKlaberjassHint(base({ currentPlayerIdx: 1 }))).toBeNull();
  });

  it('stays quiet between hands', () => {
    expect(getKlaberjassHint(base({ phase: KlaberjassPhase.HAND_END }))).toBeNull();
  });

  // **追随・切り札・オーバートランプはすべて強制。**選べる札が 1 枚しかない
  // 局面は珍しくなく、しかも気づきにくい。
  it('names the only legal card', () => {
    expect(getKlaberjassHint(base({ validPlays: [1] }))).toEqual({
      targetAction: 'card-1',
      reason: 'frontendHint.klaberjassForced',
      confidence: 'strong',
    });
  });

  // **札 0 も強制手になりうる。**真偽値で見ると先頭の札だけ落ちる。
  it('keeps a forced play on card index 0', () => {
    expect(getKlaberjassHint(base({ validPlays: [0] }))?.targetAction).toBe('card-0');
  });

  // **ベラは K+Q を「2 枚目を出すときに」宣言する。**持っているのに気づかず
  // 崩すと 20 点が消える。
  it('warns about an unscored bela', () => {
    const hand = [card('SPADE', 13), card('SPADE', 12), card('HEART', 9)];
    const s = base({ hand, belaHolder: 0, belaScored: false, validPlays: [0, 1, 2] });
    expect(getKlaberjassHint(s)).toEqual({
      targetAction: 'bela',
      reason: 'frontendHint.klaberjassBela',
      confidence: 'strong',
    });
  });

  it('says nothing more about a bela once it has scored', () => {
    const hand = [card('SPADE', 13), card('SPADE', 12), card('HEART', 9)];
    const s = base({ hand, belaHolder: 0, belaScored: true, validPlays: [0, 1, 2] });
    expect(getKlaberjassHint(s)?.targetAction).not.toBe('bela');
  });

  it('ignores a bela held by the opponent', () => {
    const s = base({ belaHolder: 1, belaScored: false, validPlays: [0, 1] });
    expect(getKlaberjassHint(s)?.targetAction).not.toBe('bela');
  });

  // 選択の余地があるときは、サーバが出した合法手の中から一番弱い札を残さず出す
  // 判断まではせず、選べること自体を伝える。
  it('points at the legal plays when there is a choice', () => {
    expect(getKlaberjassHint(base({ validPlays: [0, 1] }))).toEqual({
      targetAction: 'card-0',
      reason: 'frontendHint.klaberjassChoose',
      confidence: 'moderate',
    });
  });

  it('stays quiet when the server offers no legal play', () => {
    expect(getKlaberjassHint(base({ validPlays: [] }))).toBeNull();
  });

  // ビッドは表向きの札を取るかどうか。手札に切り札候補が多いほど取る価値がある。
  it('recommends accepting a turn-up that matches the hand', () => {
    const hand = [card('SPADE', 13), card('SPADE', 12), card('SPADE', 9), card('HEART', 7)];
    const s = base({ phase: KlaberjassPhase.BID_TURN_UP, hand, bidPlayerIdx: 0, turnUpCard: card('SPADE', 10) });
    expect(getKlaberjassHint(s)).toEqual({
      targetAction: 'accept',
      reason: 'frontendHint.klaberjassTakeTurnUp',
      confidence: 'moderate',
    });
  });

  it('recommends passing a turn-up the hand cannot support', () => {
    const hand = [card('HEART', 7), card('HEART', 8), card('CLOVER', 9), card('DIAMOND', 7)];
    const s = base({ phase: KlaberjassPhase.BID_TURN_UP, hand, bidPlayerIdx: 0, turnUpCard: card('SPADE', 10) });
    expect(getKlaberjassHint(s)?.targetAction).toBe('pass');
  });

  it('stays quiet during the bid when it is not the human decision', () => {
    const s = base({ phase: KlaberjassPhase.BID_TURN_UP, bidPlayerIdx: 1 });
    expect(getKlaberjassHint(s)).toBeNull();
  });

  it('stays quiet when no seat is the human', () => {
    const s = base();
    s.players = s.players.map((p) => ({ ...p, isHuman: false }));
    expect(getKlaberjassHint(s)).toBeNull();
  });

  // 切り札が未決のあいだ (trumpSuit 0) はベラの判定ができない。
  it('does not claim a bela before trump is fixed', () => {
    const hand = [card('SPADE', 13), card('SPADE', 12), card('HEART', 9)];
    const s = base({ hand, trumpSuit: 0, belaHolder: 0, belaScored: false, validPlays: [0, 1, 2] });
    expect(getKlaberjassHint(s)?.targetAction).not.toBe('bela');
  });

  it('stays quiet without a turn-up card to judge', () => {
    const s = base({ phase: KlaberjassPhase.BID_TURN_UP, bidPlayerIdx: 0, turnUpCard: null });
    expect(getKlaberjassHint(s)).toBeNull();
  });
});
