import { describe, expect, it } from 'vitest';
import { comboAnnouncement } from './comboAnnounce';

describe('comboAnnouncement', () => {
  it('reports the running chain once the badge appears', () => {
    // バッジが出るのは 2 以上。読み上げも同じ境界で始める。
    expect(comboAnnouncement(2, 1)).toEqual({ key: 'comboAnnounce', count: 2 });
    expect(comboAnnouncement(7, 6)).toEqual({ key: 'comboAnnounce', count: 7 });
  });

  it('reports the break when a chain drops out of range', () => {
    expect(comboAnnouncement(0, 4)).toEqual({ key: 'comboEnded', count: 0 });
    expect(comboAnnouncement(1, 2)).toEqual({ key: 'comboEnded', count: 0 });
  });

  // **負のコントロール。** 連鎖が無かったところで「途切れた」と言うと、
  // 一手ごとに雑音が出る。
  it('says nothing when there was no chain to break', () => {
    expect(comboAnnouncement(0, 0)).toBeNull();
    expect(comboAnnouncement(1, 1)).toBeNull();
    expect(comboAnnouncement(0, 1)).toBeNull();
  });
});
