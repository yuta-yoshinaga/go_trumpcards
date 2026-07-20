import type { Card } from '../types/card';

/** A single recorded "say word" attempt in Mao: the spoken word, the board
 * (the discard-pile top card at the moment of speaking), and whether it drew a
 * hidden-rule penalty. */
export interface MaoSayWordAttempt {
  /** The word the player tried saying. */
  word: string;
  /** The discard-pile top card when the word was spoken (null if the pile was empty). */
  board: Card | null;
  /** True when the attempt broke the hidden rule (penalty); false when it was accepted. */
  penalty: boolean;
}

/** Maximum number of say-word attempts retained in the local history. */
export const MAX_SAY_WORD_HISTORY = 50;

/** Append a say-word attempt to the history (newest last), capping the length
 * to {@link MAX_SAY_WORD_HISTORY} by dropping the oldest entries. Returns a new
 * array; the input is never mutated. */
export function appendSayWordAttempt(
  history: readonly MaoSayWordAttempt[],
  attempt: MaoSayWordAttempt,
  max: number = MAX_SAY_WORD_HISTORY,
): MaoSayWordAttempt[] {
  const next = [...history, attempt];
  return next.length > max ? next.slice(next.length - max) : next;
}
