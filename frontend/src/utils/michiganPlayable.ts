import type { MichiganResponse } from '../types/card';

/** The next-playable summary for the human's Michigan hand, derived from the current sequence state. */
export interface MichiganNextPlayable {
  /** Whether the human may start a fresh sequence (each suit's lowest card leads). */
  isNewSequence: boolean;
  /** Display name of the active sequence's suit (empty when a new sequence is needed). */
  suitName: string;
  /** The rank that continues the active sequence (`seqHighValue + 1`); only meaningful when `isNewSequence` is false. */
  nextValue: number;
}

/**
 * Computes what the human needs to play next onto the current Michigan sequence.
 *
 * A new sequence is signalled by `needNewSequence` or `seqSuit === 0`, in which
 * case any suit's lowest card may lead. Otherwise the sole legal continuation is
 * the same-suit card one rank above `seqHighValue`.
 *
 * @param state - The sequence-relevant subset of the Michigan game state.
 * @returns The derived next-playable summary.
 */
export function michiganNextPlayable(
  state: Pick<MichiganResponse, 'seqSuit' | 'seqSuitName' | 'seqHighValue' | 'needNewSequence'>,
): MichiganNextPlayable {
  return {
    isNewSequence: state.needNewSequence || state.seqSuit === 0,
    suitName: state.seqSuitName,
    nextValue: state.seqHighValue + 1,
  };
}
