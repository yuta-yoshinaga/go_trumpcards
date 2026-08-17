import { describe, expect, it } from 'vitest';
import { skatScoreBonusKeys } from './skatScoreBonuses';

const none = {
  base: 11,
  matadors: 2,
  multiplier: 4,
  hand: false,
  schneider: false,
  schwarz: false,
  doubled: false,
  overbid: false,
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
    ['doubled', 'bonus.doubled'],
    ['overbid', 'bonus.overbid'],
  ])('lists %s on its own', (flag, key) => {
    expect(skatScoreBonusKeys({ ...none, [flag]: true })).toEqual([key]);
  });

  it('keeps the announcement order when several apply', () => {
    expect(skatScoreBonusKeys({ ...none, overbid: true, hand: true, schwarz: true })).toEqual([
      'bonus.hand',
      'bonus.schwarz',
      'bonus.overbid',
    ]);
  });
});
