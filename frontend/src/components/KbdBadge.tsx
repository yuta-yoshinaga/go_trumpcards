/** Props for the {@link KbdBadge} component. */
export interface KbdBadgeProps {
  /** Key label rendered inside the badge (e.g. "Space", "Enter", "S"). */
  label: string;
}

/**
 * Renders a small `<kbd>` chip used to advertise keyboard shortcuts on
 * action buttons. The badge is non-interactive — purely an affordance.
 */
export function KbdBadge({ label }: KbdBadgeProps) {
  return (
    <kbd className="ml-2 inline-flex items-center px-1.5 py-0.5 rounded border border-white/40 bg-white/15 text-[10px] font-mono leading-none">
      {label}
    </kbd>
  );
}
