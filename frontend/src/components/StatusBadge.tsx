export type StatusBadgeVariant = 'success' | 'warning' | 'danger';

const variantClasses: Record<StatusBadgeVariant, string> = {
  success: 'bg-game-status-active text-white rounded-[6px] px-2 py-[1px] ml-1.5 text-xs',
  warning: 'bg-game-status-waiting text-game-text-strong rounded-[6px] px-2 py-[1px] ml-1.5 text-xs font-bold',
  danger: 'bg-game-status-out text-white rounded-[6px] px-2 py-[1px] ml-1.5 text-xs font-bold',
};

export function StatusBadge({ variant, children }: { variant: StatusBadgeVariant; children: React.ReactNode }) {
  return <span className={variantClasses[variant]}>{children}</span>;
}
