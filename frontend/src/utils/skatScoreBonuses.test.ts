import { describe, expect, it } from 'vitest';
import { skatScoreBonusKeys, skatScoreFormulaKey } from './skatScoreBonuses';

const none = {
  base: 11,
  matadors: 2,
  multiplier: 4,
  hand: false,
  schneider: false,
  schwarz: false,
  doubled: false,
  overbid: false,
  bid: 18,
  value: 44,
  null: false,
};

describe('skatScoreBonusKeys', () => {
  it('lists nothing when no bonus applies', () => {
    expect(skatScoreBonusKeys(none)).toEqual([]);
  });

  // **1 件ずつ独立に見る。**まとめて 1 本にすると、フラグを 1 つ落とす変異が
  // 他のフラグの出力に隠れる。
  it.each([
    ['hand', 'bonus.hand'],
    ['schneider', 'bonus.schneider'],
    ['schwarz', 'bonus.schwarz'],
  ])('lists %s on its own', (flag, key) => {
    expect(skatScoreBonusKeys({ ...none, [flag]: true })).toEqual([key]);
  });

  // **敗北の2倍とオーバービッドは乗数のボーナスではない。**式そのものが変わるので
  // skatScoreFormulaKey が言う。ここに混ぜると同じことを二度言う。
  it('does not list the doubling or the overbid', () => {
    expect(skatScoreBonusKeys({ ...none, doubled: true, overbid: true })).toEqual([]);
  });

  it('keeps the announcement order when several apply', () => {
    expect(skatScoreBonusKeys({ ...none, schneider: true, hand: true, schwarz: true })).toEqual([
      'bonus.hand',
      'bonus.schneider',
      'bonus.schwarz',
    ]);
  });
});

// **式は 3 通りある。**1 つに固定すると、負けたラウンド (およそ半分) で
// 「11 × 3 = 66」という嘘の式が出る (#5561 のレビュー指摘)。
describe('skatScoreFormulaKey', () => {
  it('states base x multiplier for a win', () => {
    expect(skatScoreFormulaKey(none)).toBe('scoreBreakdown');
  });

  it('states the doubling for a loss', () => {
    expect(skatScoreFormulaKey({ ...none, doubled: true })).toBe('scoreBreakdownDoubled');
  });

  it('states the bid for an overbid', () => {
    expect(skatScoreFormulaKey({ ...none, overbid: true })).toBe('scoreBreakdownOverbid');
  });

  // オーバービッドは基礎点と無関係な値に置き換わるので、2倍より優先する。
  // (現状ドメインでは同時に立たないが、優先順位を決めておかないと将来沈黙する。)
  it('prefers the overbid sentence when both are set', () => {
    expect(skatScoreFormulaKey({ ...none, doubled: true, overbid: true })).toBe('scoreBreakdownOverbid');
  });
});
