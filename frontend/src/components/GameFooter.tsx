/** Props for {@link GameFooter}. */
export interface GameFooterProps {
  className?: string;
  children: React.ReactNode;
  /** When true, applies glassmorphism + rounded top corners + scroll on mobile. */
  floating?: boolean;
}

/** Renders a sticky footer area for game action buttons with safe-area padding. */
export function GameFooter({ className, children, floating = false }: GameFooterProps) {
  const floatingClass = floating
    ? 'glass-panel rounded-t-xl max-h-[60vh] overflow-y-auto sm:max-h-none sm:overflow-y-visible'
    : '';
  return (
    <footer
      className={['shrink-0', 'border-t', floatingClass, className].filter(Boolean).join(' ')}
      style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
    >
      {children}
    </footer>
  );
}
