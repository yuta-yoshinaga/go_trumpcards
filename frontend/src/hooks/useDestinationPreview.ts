import { useCallback, useMemo, useState } from 'react';

/** What {@link useDestinationPreview} hands back to a page. */
export interface DestinationPreview<S> {
  /**
   * The card whose destinations should be highlighted right now: the committed
   * selection when there is one, otherwise the card under the pointer/focus.
   */
  source: S | null;
  /** True when `source` came from hover or focus rather than a click. */
  isPreview: boolean;
  /** Spread onto a card control to arm the preview for the source it represents. */
  previewProps: (source: S) => {
    onMouseEnter: () => void;
    onMouseLeave: () => void;
    onFocus: () => void;
    onBlur: () => void;
  };
  /** Drop the preview (used when a page resets or commits a move). */
  clear: () => void;
}

/**
 * Turns hover/focus over a card into the same "where could this go?" question
 * the page already answers for the selected card.
 *
 * **The legality computation is not duplicated.** A page keeps calling whatever
 * util or server field it already uses; this only decides *which* card that
 * computation is asked about, so the preview can never disagree with the
 * highlight shown after the click.
 *
 * **A selection wins over hover.** Once a card is picked the targets must stay
 * put while the pointer travels to one of them — otherwise the highlight would
 * chase the cursor and vanish exactly when it is being aimed at.
 *
 * **Focus counts as hover.** Hover is a pointer affordance and a keyboard user
 * would otherwise never see it; the cards are already focusable buttons, so
 * `onFocus`/`onBlur` costs nothing and covers them. Touch devices simply never
 * fire these, which leaves the pre-selection preview a desktop affordance by
 * construction rather than by a media query.
 *
 * @param selected - The page's committed selection, or null.
 * @returns The source to highlight, whether it is a preview, and the handlers.
 */
export function useDestinationPreview<S>(selected: S | null): DestinationPreview<S> {
  const [hovered, setHovered] = useState<S | null>(null);

  const previewProps = useCallback(
    (source: S) => ({
      onMouseEnter: () => setHovered(source),
      // Leaving any card drops the preview outright. Comparing against the
      // current value would need object identity, and pages build a fresh
      // source object per render.
      onMouseLeave: () => setHovered(null),
      onFocus: () => setHovered(source),
      onBlur: () => setHovered(null),
    }),
    [],
  );

  const clear = useCallback(() => setHovered(null), []);

  return useMemo(
    () => ({
      source: selected ?? hovered,
      isPreview: selected === null && hovered !== null,
      previewProps,
      clear,
    }),
    [selected, hovered, previewProps, clear],
  );
}
