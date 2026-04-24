import { PineapplePage } from './PineapplePage';

/** Renders the Crazy Pineapple Poker variant: same UI as Pineapple Poker
 * but wired to the `/crazypineapple/exec` endpoint and i18n namespace.
 * Crazy Pineapple discards AFTER the flop betting round (Pineapple discards
 * immediately after the flop is dealt). */
export function CrazyPineapplePage() {
  return <PineapplePage variant="crazypineapple" />;
}
