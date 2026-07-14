import { expect, test } from '@playwright/test';
import { navigateTo, waitForLoaded } from './helpers';

/** A dealt card with its rank (A=14 high), suit symbol, and 0-based button index. */
type PCard = { r: number; s: string; i: number };

/** Parse the rank value from a card alt label like "♥ A" or "♦ 10" (Ace high). */
function parseRank(label: string): number {
  const token = label.trim().split(/\s+/).pop() ?? '';
  if (token === 'A') return 14;
  if (token === 'K') return 13;
  if (token === 'Q') return 12;
  if (token === 'J') return 11;
  return Number.parseInt(token, 10);
}

/** Parse the suit symbol from a card alt label like "♥ A". */
function parseSuit(label: string): string {
  return label.trim().split(/\s+/)[0] ?? '';
}

/** Ranks ordered by frequency (desc), then by rank (desc) — poker kicker order. */
function freqOrder(ranks: number[]): number[] {
  const count = new Map<number, number>();
  for (const r of ranks) count.set(r, (count.get(r) ?? 0) + 1);
  return [...ranks].sort((a, b) => {
    const ca = count.get(a) ?? 0;
    const cb = count.get(b) ?? 0;
    return ca !== cb ? cb - ca : b - a;
  });
}

/** Distinct count values (desc) for a set of ranks, e.g. a full house → [3, 2]. */
function countShape(ranks: number[]): number[] {
  const count = new Map<number, number>();
  for (const r of ranks) count.set(r, (count.get(r) ?? 0) + 1);
  return [...count.values()].sort((a, b) => b - a);
}

/** True when 5 ranks form a straight (Ace high; A-2-3-4-5 wheel allowed). */
function isStraight(ranks: number[]): boolean {
  const sorted = [...ranks].sort((a, b) => a - b);
  if (new Set(sorted).size !== 5) return false;
  if (sorted[4] - sorted[0] === 4) return true;
  return sorted.join(',') === '2,3,4,5,14';
}

/** Score a 5-card hand as [category, ...frequency-ordered ranks]. Higher is stronger. */
function rank5(cards: PCard[]): number[] {
  const ranks = cards.map((c) => c.r);
  const counts = countShape(ranks);
  const flush = new Set(cards.map((c) => c.s)).size === 1;
  const straight = isStraight(ranks);
  let cat: number;
  if (flush && straight) cat = 8;
  else if (counts[0] === 4) cat = 7;
  else if (counts[0] === 3 && counts[1] === 2) cat = 6;
  else if (flush) cat = 5;
  else if (straight) cat = 4;
  else if (counts[0] === 3) cat = 3;
  else if (counts[0] === 2 && counts[1] === 2) cat = 2;
  else if (counts[0] === 2) cat = 1;
  else cat = 0;
  return [cat, ...freqOrder(ranks)];
}

/** Score a 3-card front hand (no straights/flushes count): trips=3, pair=1, high=0. */
function rank3(cards: PCard[]): number[] {
  const ranks = cards.map((c) => c.r);
  const counts = countShape(ranks);
  const cat = counts[0] === 3 ? 3 : counts[0] === 2 ? 1 : 0;
  return [cat, ...freqOrder(ranks)];
}

/** Lexicographic compare of two hand scores. >0 if a is stronger than b. */
function cmp(a: number[], b: number[]): number {
  const n = Math.min(a.length, b.length);
  for (let i = 0; i < n; i++) {
    if (a[i] !== b[i]) return a[i] - b[i];
  }
  return 0;
}

/** Generate all k-combinations of arr. */
function combinations<T>(arr: T[], k: number): T[][] {
  const res: T[][] = [];
  const rec = (start: number, combo: T[]) => {
    if (combo.length === k) {
      res.push([...combo]);
      return;
    }
    for (let i = start; i < arr.length; i++) {
      combo.push(arr[i]);
      rec(i + 1, combo);
      combo.pop();
    }
  };
  rec(0, []);
  return res;
}

/**
 * Find a legal (non-foul) arrangement: front(3) ≤ middle(5) ≤ back(5) by hand
 * strength. Brute-forces all 3/5/5 partitions and returns the button indices to
 * assign to front and middle (the remaining 5 fall to back automatically).
 * Returns null only if no legal arrangement exists (effectively never for a
 * random 13-card deal).
 */
function legalArrangement(cards: PCard[]): { front: number[]; middle: number[] } | null {
  const pos = cards.map((_, i) => i);
  for (const f of combinations(pos, 3)) {
    const fset = new Set(f);
    const rem = pos.filter((i) => !fset.has(i));
    const fScore = rank3(f.map((i) => cards[i]));
    for (const m of combinations(rem, 5)) {
      const mScore = rank5(m.map((i) => cards[i]));
      if (cmp(fScore, mScore) > 0) continue;
      const mset = new Set(m);
      const back = rem.filter((i) => !mset.has(i));
      if (cmp(mScore, rank5(back.map((i) => cards[i]))) <= 0) {
        return { front: f.map((i) => cards[i].i), middle: m.map((i) => cards[i].i) };
      }
    }
  }
  return null;
}

test.describe('Chinese Poker E2E', () => {
  test('plays a round: bet → set hands → result → reset', async ({ page }) => {
    await navigateTo(page, '/chinesepoker');

    // BET phase: click ベット
    const betButton = page.getByRole('button', { name: 'ベット' });
    await expect(betButton).toBeVisible();
    await betButton.click();
    await waitForLoaded(page);

    // SET HANDS phase: read the 13 dealt cards and compute a legal arrangement.
    // A naive "first 3 → front, next 5 → middle" split is almost always a foul
    // (front ≤ middle ≤ back is violated), which the server rejects — so the
    // arrangement must be rank-aware to reliably reach the END phase.
    const cards = page.locator('[data-testid^="cp-hand-card-"]');
    await expect(cards.first()).toBeVisible({ timeout: 10_000 });
    const count = await cards.count();

    const dealt: PCard[] = [];
    for (let i = 0; i < count; i++) {
      const label = (await cards.nth(i).locator('img').first().getAttribute('alt')) ?? '';
      dealt.push({ r: parseRank(label), s: parseSuit(label), i });
    }

    const arrangement = legalArrangement(dealt);
    // A legal arrangement effectively always exists for a random 13-card deal;
    // throwing (rather than asserting) narrows the type for the clicks below.
    if (!arrangement) throw new Error('no legal (non-foul) arrangement for the dealt hand');

    // Click the 3 front cards first (they fill the front slots), then the 5
    // middle cards (they fill the middle slots); the remaining 5 fall to back.
    for (const i of arrangement.front) await cards.nth(i).click();
    for (const i of arrangement.middle) await cards.nth(i).click();

    const setButton = page.getByRole('button', { name: 'セット' });
    await expect(setButton).toBeVisible();
    await setButton.click();
    await waitForLoaded(page);

    // END phase: 次のゲーム button should be visible
    const resetButton = page.getByRole('button', { name: '次のゲーム' });
    await expect(resetButton).toBeVisible({ timeout: 10_000 });

    // Reset back to bet phase
    await resetButton.click();
    await waitForLoaded(page);
    await expect(page.getByRole('button', { name: 'ベット' })).toBeVisible();
  });
});
