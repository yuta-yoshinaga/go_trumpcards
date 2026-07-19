import { describe, expect, it } from 'vitest';
import { moveFocus } from './gridNav';

// 4x3 grid (cols=4, total=12):
//  0  1  2  3
//  4  5  6  7
//  8  9 10 11
describe('moveFocus', () => {
  const cols = 4;
  const total = 12;

  it('moves right/left within a row', () => {
    expect(moveFocus(5, 'right', cols, total)).toBe(6);
    expect(moveFocus(5, 'left', cols, total)).toBe(4);
  });

  it('moves up/down by one column width', () => {
    expect(moveFocus(5, 'down', cols, total)).toBe(9);
    expect(moveFocus(5, 'up', cols, total)).toBe(1);
  });

  it('clamps at the left edge', () => {
    expect(moveFocus(4, 'left', cols, total)).toBe(4);
    expect(moveFocus(0, 'left', cols, total)).toBe(0);
  });

  it('clamps at the right edge', () => {
    expect(moveFocus(3, 'right', cols, total)).toBe(3);
    expect(moveFocus(7, 'right', cols, total)).toBe(7);
  });

  it('clamps at the top edge', () => {
    expect(moveFocus(2, 'up', cols, total)).toBe(2);
  });

  it('clamps at the bottom edge', () => {
    expect(moveFocus(9, 'down', cols, total)).toBe(9);
  });

  it('does not run past the last cell on a partial final row', () => {
    // total=10, cols=4 -> last row is 8,9 only
    expect(moveFocus(9, 'right', cols, 10)).toBe(9);
    expect(moveFocus(6, 'down', cols, 10)).toBe(6);
    expect(moveFocus(5, 'down', cols, 10)).toBe(9);
  });

  it('skips cells flagged by isSkipped', () => {
    // 5 is skipped, so moving right from 4 lands on 6
    const skip = (i: number) => i === 5;
    expect(moveFocus(4, 'right', cols, total, skip)).toBe(6);
  });

  it('skips multiple consecutive cells', () => {
    const skip = (i: number) => i === 5 || i === 6;
    expect(moveFocus(4, 'right', cols, total, skip)).toBe(7);
  });

  it('stays put when only skipped cells lie ahead before an edge', () => {
    const skip = (i: number) => i === 6 || i === 7;
    expect(moveFocus(5, 'right', cols, total, skip)).toBe(5);
  });

  it('returns index for a degenerate grid', () => {
    expect(moveFocus(0, 'right', 0, total)).toBe(0);
    expect(moveFocus(0, 'right', cols, 0)).toBe(0);
  });
});
