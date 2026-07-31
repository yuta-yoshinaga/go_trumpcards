// API client for sixbidsolo. Split-file layout introduced by issue #4434;
// gameApi.ts re-exports this file, so existing imports keep working.

import type { SixBidSoloResponse } from '../../types/card';
import { gameExec } from '../gameExec';

/** Configuration options for Six-Bid Solo game settings. */
export interface SixBidSoloConfigInput {
  cpuDifficulty?: number;
  /** Hands in a game. */
  targetHands?: number;
}

/** Commands the /sixbidsolo/exec endpoint accepts. */
export type SixBidSoloCommand = 'reset' | 'bid' | 'pass' | 'declare' | 'play' | 'next' | 'log';

/** Options carried alongside a Six-Bid Solo command. */
export interface SixBidSoloExecOptions {
  /**
   * 1 = Solo, 2 = Heart Solo, 3 = Misère, 4 = Guarantee Solo,
   * 5 = Spread Misère, 6 = Call Solo. Required for `bid`.
   */
  bid?: number;
  /** 1 = Spade, 2 = Clover, 3 = Heart, 4 = Diamond. Required for `declare`. */
  suit?: number;
  /** The card a call solo names — **both fields or neither**. */
  calledSuit?: number;
  calledValue?: number;
  /** Hand index. Required for `play`. */
  cardIndex?: number;
  config?: SixBidSoloConfigInput;
}

/**
 * API client for the Six-Bid Solo /sixbidsolo/exec endpoint.
 *
 * Six-Bid Solo deals **eleven cards each and a three-card widow** from a
 * 36-card pack; the widow is credited to the declarer at the end, except at
 * either misère. Six bids in ascending order swap the target and the payment
 * together: a plain bid must **exceed** 60 card points and pays the difference
 * from 60, while misère asks for **zero card points** — not zero tricks. Only
 * `reset` takes a `config`.
 */
export const sixBidSoloApi = {
  exec: (command: SixBidSoloCommand, opts?: SixBidSoloExecOptions) =>
    gameExec<SixBidSoloResponse>('sixbidsolo', {
      command,
      bid: opts?.bid,
      suit: opts?.suit,
      calledSuit: opts?.calledSuit,
      calledValue: opts?.calledValue,
      cardIndex: opts?.cardIndex,
      config: opts?.config,
    }),
};
