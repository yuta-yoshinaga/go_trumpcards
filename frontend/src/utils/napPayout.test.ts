import { describe, expect, it } from 'vitest';
import { napPayout } from './napPayout';

describe('napPayout', () => {
  it('stakes the trick count either way for the numbered contracts', () => {
    expect(napPayout(2)).toEqual({ make: 2, fail: 2 });
    expect(napPayout(3)).toEqual({ make: 3, fail: 3 });
    expect(napPayout(4)).toEqual({ make: 4, fail: 4 });
  });

  // **ここだけ非対称。**達成 10 に対し、失敗しても相手が得るのは各 5。
  it('pays Nap ten to make and five apiece to beat', () => {
    expect(napPayout(5)).toEqual({ make: 10, fail: 5 });
  });

  it('stakes nothing on a pass', () => {
    expect(napPayout(0)).toBeNull();
  });
});
