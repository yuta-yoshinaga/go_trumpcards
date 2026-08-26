import { describe, expect, it } from 'vitest';
import type { Card, SpeculationResponse } from '../../../types/card';
import { SpeculationPhase } from '../../../types/phases';
import { formatSpeculationState } from './speculationFormatter';

const card = (design: string, value: number): Card => ({ design, value }) as Card;

const seat = (name: string, chips: number, hiddenCount: number, best?: Card) => ({ name, chips, hiddenCount, best });

const base: SpeculationResponse = {
  phase: SpeculationPhase.FLIP,
  seats: [seat('You', 190, 3), seat('CPU1', 190, 3), seat('CPU2', 190, 2)],
  trumpSuit: 3,
  trumpCard: card('HEART', 9),
  pot: 30,
  turnSeat: 1,
  bestSeat: -1,
  offerFrom: -1,
  offerTo: -1,
  offerAmount: 0,
  roundNo: 0,
  winnerSeat: -1,
  gameEndFlag: false,
  config: { players: 3, initialChips: 200, stake: 10, rounds: 5 },
  message: '',
};

const withState = (over: Partial<SpeculationResponse>): SpeculationResponse => ({ ...base, ...over });

describe('formatSpeculationState', () => {
  it('見出しとフェーズ・ラウンド・ポットを出す', () => {
    const out = formatSpeculationState(base);
    expect(out).toContain('Speculation');
    expect(out).toContain('Phase: TURN UP');
    // roundNo は消化済み数なので、表示は +1。
    expect(out).toContain('Round: 1 / 5');
    expect(out).toContain('Pot: 30');
  });

  it('切り札のスート記号とその札を出す', () => {
    expect(formatSpeculationState(base)).toContain('Trump: ♥ (♥9)');
  });

  // **切り札が未定のときに ♠ と書いてはいけない。** -1 を配列の先頭に
  // 丸めると、切り札の無いラウンドがスペードの卓に見える。
  it('切り札が未定なら記号を出さない', () => {
    const out = formatSpeculationState(withState({ trumpSuit: -1, trumpCard: undefined }));
    expect(out).toContain('Trump: -');
    expect(out).not.toContain('Trump: ♠');
  });

  it('各席のチップと伏せ札の枚数を出す', () => {
    const out = formatSpeculationState(base);
    expect(out).toContain('You: 190 chips / 3 face down');
    expect(out).toContain('CPU2: 190 chips / 2 face down');
  });

  it('手番の席に印を付ける', () => {
    const lines = formatSpeculationState(withState({ turnSeat: 1 })).split('\n');
    const you = lines.find((l) => l.includes('You:')) ?? '';
    const cpu1 = lines.find((l) => l.includes('CPU1:')) ?? '';
    expect(cpu1).toContain('>');
    expect(you).not.toContain('>');
  });

  // **最高札の持ち主が「いない」のは -1。** 0 を「いない」と読むと、
  // 誰も切り札を出していない盤面で人間の行に印が付く。
  it('bestSeat が -1 のときはどの席にも最高札の印を付けない', () => {
    const lines = formatSpeculationState(withState({ bestSeat: -1 })).split('\n');
    const seatLines = lines.filter((l) => l.includes('face down'));
    expect(seatLines).toHaveLength(3);
    for (const line of seatLines) expect(line.startsWith('*')).toBe(false);
  });

  it('bestSeat の席にだけ最高札の印と札を出す', () => {
    const out = formatSpeculationState(
      withState({
        bestSeat: 2,
        seats: [seat('You', 190, 3), seat('CPU1', 190, 3), seat('CPU2', 190, 2, card('HEART', 13))],
      }),
    );
    const lines = out.split('\n');
    expect(lines.find((l) => l.includes('CPU2:'))).toMatch(/^\*/);
    expect(lines.find((l) => l.includes('You:'))?.startsWith('*')).toBe(false);
    expect(out).toContain('[best ♥K]');
  });

  it('人間が売り手のときは売る側の案内を出す', () => {
    const out = formatSpeculationState(
      withState({ phase: SpeculationPhase.AUCTION, offerFrom: 2, offerTo: 0, offerAmount: 25, bestSeat: 0 }),
    );
    expect(out).toContain('CPU2 offers 25 for your card.');
    expect(out).toContain('`a` to sell');
  });

  it('人間が買い手のときは買う側の案内を出す', () => {
    const out = formatSpeculationState(
      withState({ phase: SpeculationPhase.AUCTION, offerFrom: 0, offerTo: 2, offerAmount: 25, bestSeat: 2 }),
    );
    expect(out).toContain('CPU2 will part with the best trump for 25.');
    expect(out).toContain('`bid <amount>` to raise');
  });

  it('競りが開いていなければ申し出は出さない', () => {
    expect(formatSpeculationState(base)).not.toContain('to sell');
  });

  it('決着でポットを取った席を出す', () => {
    expect(formatSpeculationState(withState({ phase: SpeculationPhase.RESULT, winnerSeat: 0 }))).toContain(
      'You take the pot!',
    );
    expect(formatSpeculationState(withState({ phase: SpeculationPhase.RESULT, winnerSeat: 2 }))).toContain(
      'CPU2 takes the pot.',
    );
  });

  // **流局は「誰かが負けた」ではない。** 勝者 -1 を席 0 と読むと、
  // 切り札が 1 枚も出なかったラウンドを人間の勝ちとして表示する。
  it('切り札が出なかったラウンドは流局として出す', () => {
    const out = formatSpeculationState(withState({ phase: SpeculationPhase.RESULT, winnerSeat: -1 }));
    expect(out).toContain('No trump appeared. The stakes are returned.');
    expect(out).not.toContain('You take the pot!');
  });

  it('ゲーム終了なら最終チップを出す', () => {
    const out = formatSpeculationState(
      withState({ gameEndFlag: true, winnerSeat: 0, phase: SpeculationPhase.GAME_END }),
    );
    expect(out).toContain('Phase: FINISHED');
    expect(out).toContain('Final chips: 190');
  });
});
