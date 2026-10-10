/** Props for {@link GameFooter}. */
export interface GameFooterProps {
  className?: string;
  /** Supporting footer content, such as the player hand and game messages. */
  children?: React.ReactNode;
  /**
   * The always-visible primary action row. When provided, the footer itself does not
   * scroll; only its children scroll internally, with a 50vh cap from the sm breakpoint up.
   * Migrated pages should always provide this prop, passing `null` when a turn has no
   * actions, so the footer layout does not change between turns: `null` and `false`
   * keep this layout and omit the (otherwise empty, padded) action row.
   */
  actions?: React.ReactNode;
  /** Stable tutorial anchor rendered on the footer even when its actions are absent. */
  dataTutorial?: string;
}

/**
 * Renders a game footer with safe-area padding. During migration, pages may omit
 * `actions` to retain the legacy scrolling footer; pages that provide it keep
 * their action row visible while only the supporting content scrolls.
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
 * In the legacy form without `actions`, the mobile cap is lifted from `sm` up,
 * where the viewport is tall enough that it would only add a pointless inner scrollbar.
 */
export function GameFooter({ className, children, actions, dataTutorial }: GameFooterProps) {
  if (actions !== undefined) {
    return (
      <footer
        className={['shrink-0', 'border-t', 'flex', 'flex-col', 'max-h-[45vh]', 'sm:max-h-[50vh]', className]
          .filter(Boolean)
          .join(' ')}
        data-tutorial={dataTutorial}
        style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
      >
        <div className="min-h-0 overflow-y-auto" data-testid="game-footer-content">
          {children}
        </div>
        {actions != null && actions !== false && (
          <div className="shrink-0 pt-2" data-testid="game-footer-actions">
            {actions}
          </div>
        )}
      </footer>
    );
  }
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
