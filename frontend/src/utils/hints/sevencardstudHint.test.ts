import { describe, expect, it } from 'vitest';
import type { Card, SevenCardStudResponse } from '../../types/card';
import { SevenCardStudPhase } from '../../types/phases';
import { getSevenCardStudHint } from './sevencardstudHint';

const card = (design: Card['design'], value: number): Card => ({ design, value });

type Extra = { hole?: Card[]; door?: Card[]; currentBet?: number };

function base({
  hole = [card('SPADE', 4), card('HEART', 6)],
  door = [card('DIAMOND', 9)],
  currentBet = 0,
  ...overrides
}: Partial<SevenCardStudResponse> & Extra = {}) {
  return {
    players: [
      {
        id: 0,
        isHuman: true,
        holeCards: hole,
        doorCards: door,
        chips: 500,
        currentBet,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
        pfr: 0,
        threeBet: 0,
        af: '',
      },
      {
        id: 1,
        isHuman: false,
        holeCards: [],
        doorCards: [card('CLOVER', 12)],
        chips: 500,
        currentBet: 0,
        folded: false,
        allIn: false,
        handRank: 0,
        handName: '',
        bestHand: [],
        playStyleName: '',
        totalHands: 0,
        vpip: 0,
        pfr: 0,
        threeBet: 0,
        af: '',
      },
    ],
    communityCard: null,
    pot: 30,
    sidePots: [],
    dealerIdx: 0,
    currentTurn: 0,
    phase: SevenCardStudPhase.FOURTH_STREET,
    gameEndFlag: false,
    lastBet: 0,
    minRaise: 10,
    bettingLimit: 0,
    raiseCount: 0,
    maxBetAmount: 0,
    roundResults: [],
    cpuActions: [],
    handCount: 1,
    ante: 1,
    bringIn: 5,
    smallBet: 10,
    bigBet: 20,
    tournamentMode: false,
    anteLevelHands: 0,
    anteMultiplier: 100,
    tableSize: 2,
    bringInPlayerIdx: 1,
    rebuyAvailable: false,
    addonAvailable: false,
    rebuyCounts: [],
    addonUsed: [],
    rebuyEnabled: false,
    addonEnabled: false,
    rebuyMaxCount: 0,
    rebuyChips: 0,
    ...overrides,
  } as unknown as SevenCardStudResponse;
}

describe('getSevenCardStudHint', () => {
  it('stays quiet once the hand is over', () => {
    expect(getSevenCardStudHint(base({ gameEndFlag: true }))).toBeNull();
  });

  it('stays quiet at showdown, where there is nothing to decide', () => {
    expect(getSevenCardStudHint(base({ phase: SevenCardStudPhase.SHOWDOWN }))).toBeNull();
  });

  it('stays quiet when another seat is to act', () => {
    expect(getSevenCardStudHint(base({ currentTurn: 1 }))).toBeNull();
  });

  it('stays quiet for a folded seat', () => {
    const s = base();
    s.players[0].folded = true;
    expect(getSevenCardStudHint(s)).toBeNull();
  });

  it('stays quiet for an all-in seat, which has nothing left to choose', () => {
    const s = base();
    s.players[0].allIn = true;
    expect(getSevenCardStudHint(s)).toBeNull();
  });

  it('does not read the forced bring-in as strength', () => {
    const s = base({ phase: SevenCardStudPhase.THIRD_STREET, bringInPlayerIdx: 0 });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudBringIn');
  });

  it('says nothing special about the bring-in seat on later streets', () => {
    // ブリングインは 3rd street だけの強制。以降は普通の判断に戻る。
    const s = base({ phase: SevenCardStudPhase.FOURTH_STREET, bringInPlayerIdx: 0 });
    expect(getSevenCardStudHint(s)?.reason).not.toBe('frontendHint.sevencardstudBringIn');
  });

  it('raises on a pair', () => {
    const s = base({ hole: [card('SPADE', 9), card('HEART', 4)], door: [card('DIAMOND', 9)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
  });

  it('finds a pair across the hole and door cards, not only within one', () => {
    const s = base({ hole: [card('SPADE', 3)], door: [card('DIAMOND', 3), card('CLOVER', 12)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
  });

  it('checks for free when nothing is owed', () => {
    expect(getSevenCardStudHint(base({ lastBet: 0 }))?.reason).toBe('frontendHint.sevencardstudCheckFree');
  });

  it('treats a matched bet as nothing owed', () => {
    // 既に同額を払っていれば負債は 0。差し引かないとフォールドを勧めてしまう。
    const s = base({ lastBet: 20, currentBet: 20 });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudCheckFree');
  });

  it('calls a bet with a high card', () => {
    const s = base({ lastBet: 20, hole: [card('SPADE', 1), card('HEART', 4)], door: [card('DIAMOND', 6)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudCallHigh');
  });

  it('folds a bet with nothing', () => {
    const s = base({ lastBet: 20, hole: [card('SPADE', 4), card('HEART', 6)], door: [card('DIAMOND', 9)] });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudFoldWeak');
  });

  it('plays a low draw under Hi-Lo that it would fold at high only', () => {
    const low = [card('SPADE', 2), card('HEART', 3)];
    const door = [card('DIAMOND', 5), card('CLOVER', 6), card('SPADE', 7)];
    const hiLo = base({ lastBet: 20, isHiLo: true, hole: low, door });
    expect(hiLo?.isHiLo).toBe(true);
    expect(getSevenCardStudHint(hiLo)?.reason).toBe('frontendHint.sevencardstudPlayLow');

    // **同じ手が高目だけのテーブルではただの弱い手。**分岐が isHiLo を見ている確認。
    const high = base({ lastBet: 20, hole: low, door });
    expect(getSevenCardStudHint(high)?.reason).toBe('frontendHint.sevencardstudFoldWeak');
  });

  it('answers a paired low hand as a pair, before looking at the low', () => {
    // 2-2-3-4-5。ロー判定より先にペア分岐で返る。この順序が `lowCards` に
    // 重複ランクを渡さない根拠なので、順序自体を固定しておく。
    const s = base({
      lastBet: 20,
      isHiLo: true,
      hole: [card('SPADE', 2), card('HEART', 2)],
      door: [card('DIAMOND', 3), card('CLOVER', 4), card('SPADE', 5)],
    });
    expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
  });

  it('does not call nine a low card', () => {
    const s = base({
      lastBet: 20,
      isHiLo: true,
      hole: [card('SPADE', 2), card('HEART', 3)],
      door: [card('DIAMOND', 5), card('CLOVER', 6), card('SPADE', 9)],
    });
    expect(getSevenCardStudHint(s)?.reason).not.toBe('frontendHint.sevencardstudPlayLow');
  });

  it('stays quiet without any cards', () => {
    expect(getSevenCardStudHint(base({ hole: [], door: [] }))).toBeNull();
  });

  it('checks rather than calling a low draw when nothing is owed', () => {
    // `lastBet` はストリートごとに 0 に戻る。先に動く番では Call ボタン自体が
    // 描画されないので、コールを勧めると押せない操作を指す (#4643 のレビュー指摘)。
    const s = base({
      lastBet: 0,
      isHiLo: true,
      hole: [card('SPADE', 2), card('HEART', 3)],
      door: [card('DIAMOND', 5), card('CLOVER', 6), card('SPADE', 7)],
    });
    const hint = getSevenCardStudHint(s);
    expect(hint?.targetAction).toBe('check');
    expect(hint?.reason).toBe('frontendHint.sevencardstudCheckLow');
  });

  // **シカゴのスペード半分は伏せ札の最高スペードで決まる (#6621)。**
  // A♠ は誰にも抜かれないので、役が弱くても半分は確定する。それを通常の
  // スタッドと同じ基準で降ろすと、確実な半分を捨てさせることになる。
  describe('Chicago spade half', () => {
    const weakDoor = [card('DIAMOND', 4), card('CLOVER', 6), card('HEART', 9)];

    it('names the locked spade half instead of a generic high card', () => {
      const hole = [card('SPADE', 1), card('HEART', 3)];
      const chicago = base({ lastBet: 20, isChicago: true, hole, door: weakDoor });
      const hint = getSevenCardStudHint(chicago);
      expect(hint?.reason).toBe('frontendHint.sevencardstudPlaySpadeLock');
      expect(hint?.targetAction).toBe('call');
      // 半分が確定している局面なので、様子見の moderate ではない。
      expect(hint?.confidence).toBe('strong');

      // **同じ手を通常のスタッドに置くと、A はただの高い札になる** ── 打ち手は
      // 同じ「コール」でも、確定した半分なのか様子見なのかが区別できなかった。
      // 分岐が isChicago を見ていることの確認でもある。
      const plain = base({ lastBet: 20, hole, door: weakDoor });
      const plainHint = getSevenCardStudHint(plain);
      expect(plainHint?.reason).toBe('frontendHint.sevencardstudCallHigh');
      expect(plainHint?.confidence).toBe('moderate');
    });

    it('takes the free card instead of calling when nothing is owed', () => {
      const s = base({
        lastBet: 0,
        isChicago: true,
        hole: [card('SPADE', 1), card('HEART', 3)],
        door: weakDoor,
      });
      const hint = getSevenCardStudHint(s);
      expect(hint?.reason).toBe('frontendHint.sevencardstudCheckSpadeLock');
      expect(hint?.targetAction).toBe('check');
    });

    // **スートを見ていること。** 半分が付くのはスペードだけで、A♥ では
    // 何も確定しない。ランクだけ見る実装だとここが通ってしまう。
    it('does not treat an off-suit ace as the spade half', () => {
      const s = base({
        lastBet: 20,
        isChicago: true,
        hole: [card('HEART', 1), card('DIAMOND', 3)],
        door: weakDoor,
      });
      expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudCallHigh');
    });

    // **抜かれうるスペードでは半分を約束しない。** K♠ は伏せ札の A♠ に負ける
    // ので、「半分は確定」と言えるのは A♠ だけ。
    it('does not promise the spade half for a king of spades', () => {
      const s = base({
        lastBet: 20,
        isChicago: true,
        hole: [card('SPADE', 13), card('HEART', 3)],
        door: weakDoor,
      });
      expect(getSevenCardStudHint(s)?.reason).not.toBe('frontendHint.sevencardstudPlaySpadeLock');
    });

    // **門札の A♠ は勘定に入らない。** ドメインの EvalChicagoSpade は
    // holeCards しか走らない。ここを holeCards+doorCards で見ると、
    // 取れない半分を約束することになる。
    it('ignores an ace of spades that is face up', () => {
      const s = base({
        lastBet: 20,
        isChicago: true,
        hole: [card('HEART', 3), card('DIAMOND', 8)],
        door: [card('SPADE', 1), card('CLOVER', 6), card('HEART', 9)],
      });
      expect(getSevenCardStudHint(s)?.reason).not.toBe('frontendHint.sevencardstudPlaySpadeLock');
    });

    // ペアはスペードより先に返る (レイズの方が強い助言なので順序を固定する)。
    it('still answers a pair as a pair', () => {
      const s = base({
        lastBet: 20,
        isChicago: true,
        hole: [card('SPADE', 1), card('HEART', 1)],
        door: weakDoor,
      });
      expect(getSevenCardStudHint(s)?.reason).toBe('frontendHint.sevencardstudRaisePair');
    });
  });
});
