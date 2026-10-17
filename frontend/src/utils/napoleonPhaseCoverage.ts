/** A phase checkpoint in one Napoleon round. */
export type NapoleonRoundCheckpoint = 'bid' | 'declare' | 'exchange' | 'play' | 'trick-end' | 'round-boundary';

const REQUIRED: readonly NapoleonRoundCheckpoint[] = ['bid', 'play', 'round-boundary'];

/** Return checkpoints that have not yet been observed. */
export function missingNapoleonRoundCheckpoints(
  observed: ReadonlySet<NapoleonRoundCheckpoint>,
): NapoleonRoundCheckpoint[] {
  return REQUIRED.filter((checkpoint) => !observed.has(checkpoint));
}
