interface GameFooterProps {
  className?: string;
  children: React.ReactNode;
}

export function GameFooter({ className, children }: GameFooterProps) {
  return (
    <div
      className={`shrink-0 border-t ${className ?? ''}`}
      style={{ paddingBottom: 'calc(env(safe-area-inset-bottom) + 12px)' }}
    >
      {children}
    </div>
  );
}
