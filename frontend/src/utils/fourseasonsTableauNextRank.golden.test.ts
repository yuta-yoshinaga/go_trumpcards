import { describe, expect, it } from 'vitest';
import { fourseasonsTableauNextRank } from '../pages/FourSeasonsPage';
import golden from './__fixtures__/fourseasonsTableauNextRank.golden.json';

/**
 * The "what does this cross column accept next" rule lives twice: here (used by
 * the page's `title`/`aria-label`) and in `FourSeasonsTableauNextRank` in
 * `internal/adapter/presenter/FourSeasonsCuiPresenter.go`, which the CUI prints
 * (#5738). These vectors are also asserted by
 * `TestFourSeasonsTableauNextRank_GoldenVectors`, so changing the rule on one
 * side alone fails that side, and regenerating the vectors fails the other.
 */
describe('fourseasonsTableauNextRank golden vectors (shared with the Go presenter)', () => {
  it('has vectors to check', () => {
    expect(golden.cases.length).toBeGreaterThan(0);
  });

  it.each(golden.cases)('$name', (c) => {
    expect(fourseasonsTableauNextRank(c.top)).toBe(c.next);
  });

  it('has no next rank for an empty column', () => {
    expect(fourseasonsTableauNextRank(undefined)).toBeNull();
  });
});
