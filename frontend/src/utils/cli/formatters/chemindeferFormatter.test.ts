import { describe, expect, it } from 'vitest';
import type { Card, ChemindeFerResponse } from '../../../types/card';
import { ChemindeFerPhase } from '../../../types/phases';
import { formatChemindeFerState } from './chemindeferFormatter';

const king: Card = { design: 'SPADE', value: 13 };
const five: Card = { design: 'HEART', value: 5 };

const seat = (id: number, over: Record<string, unknown> = {}) => ({
  id,
  name: `Player${id + 1}`,
  isHuman: id === 0,
  chips: 1000,
  bet: 0,
  isBanker: id === 0,
  isRepresentative: false,
  lastNet: 0,
  ...over,
});

const base = {
  players: [seat(0), seat(1)],
  phase: ChemindeFerPhase.STAKE,
  bankerIdx: 0,
  betTurn: -1,
  stake: 0,
  remainingStake: 0,
  totalBet: 0,
  stakeMin: 10,
  stakeMax: 1000,
  betMin: 0,
  betMax: 0,
  representativeIdx: -1,
  punterMayChoose: false,
  bankerHand: [],
  punterHand: [],
  bankerTotal: 0,
  punterTotal: 0,
  punterDrew: false,
  result: 0,
  roundNumber: 1,
  remainingCards: 312,
  isHumanTurn: true,
  gameEndFlag: false,
  message: '',
} as unknown as ChemindeFerResponse;

const at = (over: Partial<ChemindeFerResponse>) => ({ ...base, ...over }) as ChemindeFerResponse;

describe('formatChemindeFerState', () => {
  it('フェーズ・ラウンド・親を出す', () => {
    const out = formatChemindeFerState(at({ config: { rounds: 12, initialChips: 1000 } }));
    expect(out).toContain('STAKE');
    expect(out).toContain('Coup: 1 / 12');
    expect(out).toContain('Banker: seat 0');
  });

  it('張る前は賭けの行を出さない', () => {
    expect(formatChemindeFerState(base)).not.toContain('Bet total');
  });

  it('張った後は賭けの進み具合を出す', () => {
    const out = formatChemindeFerState(
      at({ phase: ChemindeFerPhase.BET, stake: 300, remainingStake: 250, totalBet: 50, betTurn: 1, betMax: 250 }),
    );
    expect(out).toContain('Bet total: 50 (uncovered 250)');
    expect(out).toContain('Seat 1 to bet (up to 250)');
  });

  it('配る前の手札はダッシュで出す', () => {
    expect(formatChemindeFerState(at({ stake: 100 }))).not.toContain('Punter:');
  });

  it('手札と合計を出す', () => {
    const out = formatChemindeFerState(
      at({
        phase: ChemindeFerPhase.BANKER_DRAW,
        stake: 100,
        punterHand: [king, five],
        punterTotal: 5,
        bankerHand: [king, king],
        bankerTotal: 0,
      }),
    );
    expect(out).toContain('= 5');
    expect(out).toContain('= 0');
  });

  // **選べなかったことを書く。** 書かないと、規則で決まった手が
  // 「サーバが勝手に決めた」ように読める。
  it('選べない合計ではその旨を出す', () => {
    const out = formatChemindeFerState(
      at({
        phase: ChemindeFerPhase.PUNTER_DRAW,
        stake: 100,
        punterMayChoose: false,
        punterHand: [king, { design: 'CLOVER', value: 3 }],
        punterTotal: 3,
        bankerHand: [king, king],
      }),
    );
    expect(out).toContain('no choice at this total');
  });

  it('合計 5 では選べない旨を出さない', () => {
    const out = formatChemindeFerState(
      at({
        phase: ChemindeFerPhase.PUNTER_DRAW,
        stake: 100,
        punterMayChoose: true,
        punterHand: [king, five],
        punterTotal: 5,
        bankerHand: [king, king],
      }),
    );
    expect(out).not.toContain('no choice at this total');
  });

  it('チップ行に親の印と賭け金を出す', () => {
    const out = formatChemindeFerState(
      at({ players: [seat(0, { isBanker: true }), seat(1, { isBanker: false, bet: 50, chips: 950 })] }),
    );
    expect(out).toContain('#0*:1000');
    expect(out).toContain('#1:950(50)');
  });

  it('決着と終局を出す', () => {
    expect(formatChemindeFerState(at({ result: 1 }))).toContain('banker wins');
    expect(formatChemindeFerState(at({ result: 2 }))).toContain('punters win');
    expect(formatChemindeFerState(at({ result: 3 }))).toContain('a tie');
    expect(formatChemindeFerState(at({ gameEndFlag: true }))).toContain('Game over.');
  });

  // **卓の結果と自分の損益は別の情報** (#5774)。
  it('自分の純増減も出す', () => {
    const won = formatChemindeFerState(
      at({ result: 2, players: [seat(0, { isHuman: true, lastNet: 200 }), seat(1, { isBanker: true })] }),
    );
    expect(won).toContain('your result: +200');

    const lost = formatChemindeFerState(
      at({ result: 1, players: [seat(0, { isHuman: true, lastNet: -50 }), seat(1, { isBanker: true })] }),
    );
    expect(lost).toContain('your result: -50');

    // 賭けていない回は行を落とさず「増減なし」と言う。
    expect(formatChemindeFerState(at({ result: 3 }))).toContain('your result: no change');
    // ラウンド中は決着行ごと出ない。
    expect(formatChemindeFerState(at({ result: 0 }))).not.toContain('your result:');
  });
});
