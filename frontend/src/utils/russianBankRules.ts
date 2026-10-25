import type { Card, RussianBankResponse } from '../types/card';

/** Move-source zones, shared by the page and the CLI. */
export const ZONE_RESERVE = 0;
export const ZONE_WASTE = 1;
export const ZONE_TABLEAU = 2;

/** A selected Russian Bank move source. */
export interface RussianBankSource {
  zone: number;
  fromOpp: boolean;
  col: number;
}

/** Returns the card selected from a player's reserve, waste, or tableau. */
export function selectedRussianBankCard(state: RussianBankResponse, selected: RussianBankSource): Card | undefined {
  const owner = state.players.find((p) => p.isHuman === !selected.fromOpp);
  if (selected.zone === ZONE_RESERVE) return owner?.reserveTop;
  if (selected.zone === ZONE_WASTE) return owner?.wasteTop;
  const pile = state.tableau[selected.col] ?? [];
  return pile[pile.length - 1];
}

/**
 * Returns the foundation the selected card actually goes to, or -1.
 *
 * The destination is decided by suit, so every foundation slot sends the same
 * command; the page lights only this one instead of offering eight that behave
 * alike (#6473). It reads the server's `foundationNext` rather than copying the
 * rule, so the page cannot disagree with the domain's `rbFoundationFor`, and it
 * takes the first match, as `rbFoundationFor` does.
 */
export function russianBankFoundationTarget(state: RussianBankResponse, card: Card | undefined): number {
  if (!card) return -1;
  return state.foundationNext.findIndex((n) => (n.design === '' || n.design === card.design) && n.value === card.value);
}

/** Lists legal tableau destinations for a selected card, excluding its own source column. */
export function russianBankTableauTargets(
  state: RussianBankResponse,
  card: Card | undefined,
  selected?: RussianBankSource,
): Set<number> {
  const targets = new Set<number>();
  if (!card) return targets;
  state.tableau.forEach((_, i) => {
    if (selected?.zone === ZONE_TABLEAU && selected.col === i) return;
    const next = state.tableauNext[i];
    if (
      next &&
      (next.any || (next.black === (card.design === 'SPADE' || card.design === 'CLOVER') && next.value === card.value))
    )
      targets.add(i);
  });
  return targets;
}
