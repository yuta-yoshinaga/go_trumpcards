import { describe, expect, it } from 'vitest';
import type { Card } from '../types/card';
import { appendSayWordAttempt, MAX_SAY_WORD_HISTORY, type MaoSayWordAttempt } from './maoSayWordHistory';

const board: Card = { design: 'HEART', value: 7 };

describe('appendSayWordAttempt', () => {
  it('appends a new attempt to the end (newest last)', () => {
    const first: MaoSayWordAttempt = { word: 'mao', board, penalty: false };
    const second: MaoSayWordAttempt = { word: 'no mao', board: null, penalty: true };
    const afterFirst = appendSayWordAttempt([], first);
    const afterSecond = appendSayWordAttempt(afterFirst, second);
    expect(afterFirst).toEqual([first]);
    expect(afterSecond).toEqual([first, second]);
  });

  it('does not mutate the input history', () => {
    const history: MaoSayWordAttempt[] = [{ word: 'mao', board, penalty: false }];
    const result = appendSayWordAttempt(history, { word: 'again', board, penalty: true });
    expect(history).toHaveLength(1);
    expect(result).toHaveLength(2);
  });

  it('caps the history at the max, dropping the oldest entries', () => {
    let history: MaoSayWordAttempt[] = [];
    for (let i = 0; i < MAX_SAY_WORD_HISTORY + 5; i++) {
      history = appendSayWordAttempt(history, { word: `w${i}`, board: null, penalty: false });
    }
    expect(history).toHaveLength(MAX_SAY_WORD_HISTORY);
    // Oldest five (w0..w4) dropped; first retained is w5, last is the newest.
    expect(history[0].word).toBe('w5');
    expect(history[history.length - 1].word).toBe(`w${MAX_SAY_WORD_HISTORY + 4}`);
  });

  it('respects a custom max', () => {
    const history = appendSayWordAttempt(
      [
        { word: 'a', board: null, penalty: false },
        { word: 'b', board: null, penalty: false },
      ],
      { word: 'c', board: null, penalty: true },
      2,
    );
    expect(history.map((h) => h.word)).toEqual(['b', 'c']);
  });
});
