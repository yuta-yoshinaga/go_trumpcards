/** Props for {@link GameFooter}. */
export interface GameFooterProps {
  className?: string;
  children: React.ReactNode;
}

/**
 * Renders a sticky footer area for game action buttons with safe-area padding.
 *
 * The height cap is load-bearing on mobile. This footer is `shrink-0`, so it takes
 * whatever height its content wants and the sibling `flex-1 overflow-y-auto` play
 * area is what gives way. Measured at 375x667 across all 219 game pages, the
 * tallest footer was 558px — 84% of the viewport — and 26 pages were left with
 * under 80px for their actual content. Capping at 45vh leaves at least 102px of
 * play area on every page that has both a footer and a scroll region, at the cost
 * of an inner scroll on the 47 pages whose controls exceed the cap. That trade is
 * deliberate: a footer that scrolls keeps the cards visible while the player
 * reaches a button, whereas a document that scrolls does not. See issue #4373.
 *
 * The cap is lifted from `sm` up, where the viewport is tall enough that it would
 * only add a pointless inner scrollbar.
 */
export function GameFooter({ className, children }: GameFooterProps) {
  return (
    <footer
      className={['shrink-0', 'border-t', 'max-h-[45vh] overflow-y-auto sm:max-h-none sm:overflow-y-visible', className]
        .filter(Boolean)
        .join(' ')}
      style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
    >
      {children}
    </footer>
  );
}
