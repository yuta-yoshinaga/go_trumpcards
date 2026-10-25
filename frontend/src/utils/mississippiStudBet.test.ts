import { describe, expect, it } from 'vitest';
import { MississippiStudPhase } from '../types/phases';
import {
  anteMax,
  canPlay,
  isBettingStreet,
  MISSISSIPPI_STUD_ANTE_MIN,
  MISSISSIPPI_STUD_ANTE_STEP,
  MISSISSIPPI_STUD_PLAY_MULTIPLIERS,
  playAmount,
} from './mississippiStudBet';

describe('Mississippi Stud betting rules', () => {
  it('defines the ante bounds and calculates the chip maximum', () => {
    expect(MISSISSIPPI_STUD_ANTE_MIN).toBe(10);
    expect(MISSISSIPPI_STUD_ANTE_STEP).toBe(10);
    expect(MISSISSIPPI_STUD_PLAY_MULTIPLIERS).toEqual([1, 2, 3]);
    expect(anteMax(39)).toBe(9);
    expect(anteMax(40)).toBe(10);
    expect(anteMax(800)).toBe(200);
  });
  it('checks play affordability', () => {
    expect(canPlay(1, 100, 100)).toBe(true);
    expect(canPlay(2, 199, 100)).toBe(false);
    expect(playAmount(3, 100)).toBe(300);
  });
  it('recognizes each betting street only', () => {
    expect(isBettingStreet(MississippiStudPhase.THIRD_STREET)).toBe(true);
    expect(isBettingStreet(MississippiStudPhase.FOURTH_STREET)).toBe(true);
    expect(isBettingStreet(MississippiStudPhase.FIFTH_STREET)).toBe(true);
    expect(isBettingStreet(MississippiStudPhase.ANTE)).toBe(false);
    expect(isBettingStreet(MississippiStudPhase.END)).toBe(false);
  });
});
