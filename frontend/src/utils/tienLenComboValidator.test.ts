import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { type TienLenCombo, type TienLenPlayability, tienLenPlayability } from './tienLenComboValidator';

const c = (value: number, design = 'SPADE'): Card => ({ value, design }) as unknown as Card;
const run = (values: number[]) => values.map((v) => c(v));

describe('tienLenPlayability (same rules as domain tienLenIsPlayable)', () => {
  it.each([
    ['anything playable on empty table', [c(3)], [], 'invalid', 'ok'],
    ['stronger single beats weaker', [c(6)], [c(5)], 'single', 'ok'],
    ['weaker single cannot beat stronger', [c(5)], [c(6)], 'single', 'tooWeak'],
    ['same value higher suit wins', [c(5, 'HEART')], [c(5, 'SPADE')], 'single', 'ok'],
    ['pair must beat pair', [c(6), c(6, 'HEART')], [c(5), c(5, 'HEART')], 'pair', 'ok'],
    ['single cannot beat pair', [c(2)], [c(5), c(5, 'HEART')], 'pair', 'typeMismatch'],
    ['longer straight cannot beat shorter', run([8, 9, 10, 11]), run([5, 6, 7]), 'straight', 'countMismatch'],
    ['stronger straight beats weaker', run([6, 7, 8]), run([5, 6, 7]), 'straight', 'ok'],
    ['three pair run cuts single 2', [4, 4, 5, 5, 6, 6].map((v) => c(v)), [c(2, 'HEART')], 'single', 'ok'],
    ['four of a kind cuts single 2', [9, 9, 9, 9].map((v) => c(v)), [c(2, 'HEART')], 'single', 'ok'],
    [
      'stronger three pair run beats weaker',
      [5, 5, 6, 6, 7, 7].map((v) => c(v)),
      [4, 4, 5, 5, 6, 6].map((v) => c(v)),
      'threePairRun',
      'ok',
    ],
    [
      'weaker three pair run cannot beat stronger',
      [4, 4, 5, 5, 6, 6].map((v) => c(v)),
      [5, 5, 6, 6, 7, 7].map((v) => c(v)),
      'threePairRun',
      'tooWeak',
    ],
    [
      'four of a kind beats three pair run',
      [9, 9, 9, 9].map((v) => c(v)),
      [4, 4, 5, 5, 6, 6].map((v) => c(v)),
      'threePairRun',
      'ok',
    ],
    [
      'three pair run cannot beat four of a kind',
      [4, 4, 5, 5, 6, 6].map((v) => c(v)),
      [9, 9, 9, 9].map((v) => c(v)),
      'fourOfAKind',
      'typeMismatch',
    ],
    [
      'stronger four of a kind beats weaker',
      [9, 9, 9, 9].map((v) => c(v)),
      [7, 7, 7, 7].map((v) => c(v)),
      'fourOfAKind',
      'ok',
    ],
    ['bomb cannot cut an unrelated single', [9, 9, 9, 9].map((v) => c(v)), [c(13)], 'single', 'typeMismatch'],
    ['bomb cannot cut a straight', [4, 4, 5, 5, 6, 6].map((v) => c(v)), run([3, 4, 5]), 'straight', 'typeMismatch'],
  ] as [string, Card[], Card[], TienLenCombo, TienLenPlayability][])('%s', (_name, play, table, type, expected) => {
    expect(tienLenPlayability(play, table, type)).toBe(expected);
  });
});
