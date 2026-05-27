import { PineapplePage } from './PineapplePage';

/** Renders the Irish Poker variant: 4 hole cards dealt, discard 2 after flop
 * betting, then play continues with 2 hole cards like standard Hold'em. */
export function IrishPokerPage() {
  return <PineapplePage variant="irishpoker" />;
}
