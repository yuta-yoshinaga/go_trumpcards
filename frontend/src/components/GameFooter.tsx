interface GameFooterProps {
  className?: string;
  children: React.ReactNode;
}

/** Renders a sticky footer area for game action buttons with safe-area padding. */
export function GameFooter({ className, children }: GameFooterProps) {
  return (
    <footer
      className={`shrink-0 border-t ${className ?? ''}`}
      style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
    >
      {children}
    </footer>
  );
}
