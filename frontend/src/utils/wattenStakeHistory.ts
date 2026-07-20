import type { ActionLogEntry } from '../types/card';

/** Base stake every Watten deal starts at (mirrors the domain `WattenBaseStake`). */
export const WATTEN_BASE_STAKE = 2;

/** A single stake-escalation event reconstructed from the action log. */
export interface WattenStakeEvent {
  /** Stable key for React lists (the source entry's turn number). */
  key: number;
  /** Which escalation action occurred. */
  type: 'raise' | 'hold' | 'fold';
  /** Seat index of the player who acted. */
  playerIdx: number;
  /** Stake value after this event (the proposed stake for a raise, the settled stake for hold/fold). */
  stake: number;
}

/**
 * Reconstructs the current deal's raise/respond escalation from the action log.
 *
 * Only the current deal is considered (entries after the last `declare`). The
 * stake value at each event is derived deterministically from the base stake and
 * the raise/hold/fold sequence — a raise proposes `running + 1`, a hold settles at
 * that proposal, and a fold concedes the last settled stake — so no prose parsing
 * is needed. Returns an empty array when there was no escalation.
 *
 * @param entries - Action-log entries (or null/undefined when not yet loaded).
 * @param baseStake - Stake each deal opens at (defaults to {@link WATTEN_BASE_STAKE}).
 * @returns Ordered stake-escalation events for the current deal.
 */
export function buildWattenStakeHistory(
  entries: ActionLogEntry[] | null | undefined,
  baseStake = WATTEN_BASE_STAKE,
): WattenStakeEvent[] {
  if (!entries || entries.length === 0) return [];

  // Restrict to the current deal: everything after the most recent declare.
  let start = 0;
  for (let i = entries.length - 1; i >= 0; i--) {
    if (entries[i].actionType === 'declare') {
      start = i + 1;
      break;
    }
  }

  const events: WattenStakeEvent[] = [];
  let running = baseStake;
  let pending = 0;
  for (let i = start; i < entries.length; i++) {
    const e = entries[i];
    if (e.actionType === 'raise') {
      pending = running + 1;
      events.push({ key: e.turnNumber, type: 'raise', playerIdx: e.playerIdx, stake: pending });
    } else if (e.actionType === 'hold') {
      if (pending > 0) running = pending;
      pending = 0;
      events.push({ key: e.turnNumber, type: 'hold', playerIdx: e.playerIdx, stake: running });
    } else if (e.actionType === 'fold') {
      pending = 0;
      events.push({ key: e.turnNumber, type: 'fold', playerIdx: e.playerIdx, stake: running });
    }
  }
  return events;
}
