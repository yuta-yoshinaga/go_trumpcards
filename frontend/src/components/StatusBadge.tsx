export type StatusBadgeVariant = 'success' | 'warning' | 'danger';

const variantClasses: Record<StatusBadgeVariant, string> = {
  success: 'bg-game-status-active text-white rounded-[6px] px-2 py-[1px] ml-1.5 text-[0.8em]',
  warning: 'bg-game-status-waiting text-[#222] rounded-[6px] px-2 py-[1px] ml-1.5 text-[0.8em] font-bold',
  danger: 'bg-game-status-out text-white rounded-[6px] px-2 py-[1px] ml-1.5 text-[0.8em] font-bold',
};

export function StatusBadge({ variant, children }: { variant: StatusBadgeVariant; children: React.ReactNode }) {
  return <span className={variantClasses[variant]}>{children}</span>;
}
