/** A phase checkpoint in one Napoleon round. */
export type NapoleonRoundCheckpoint = 'bid' | 'declare' | 'exchange' | 'play' | 'trick-end' | 'round-boundary';

/** Return checkpoints that have not yet been observed. */
export function missingNapoleonRoundCheckpoints(
  required: ReadonlySet<NapoleonRoundCheckpoint>,
  observed: ReadonlySet<NapoleonRoundCheckpoint>,
): NapoleonRoundCheckpoint[] {
  return [...required].filter((checkpoint) => !observed.has(checkpoint));
}
