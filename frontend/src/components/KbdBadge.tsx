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
  // The kbd is purely a visual affordance — the parent button's accessible
  // name (e.g. "Slap!") already conveys the action, so we hide the chip's
  // text from assistive tech to avoid screen readers announcing
  // "Slap! Space". Wrapping the label in an aria-hidden span (rather than
  // putting aria-hidden on the <kbd> itself) keeps biome's a11y rule happy
  // and still excludes the badge text from the accessible-name computation.
  return (
    <kbd className="ml-2 inline-flex items-center px-1.5 py-0.5 rounded border border-white/40 bg-white/15 text-[10px] font-mono leading-none">
      <span aria-hidden="true">{label}</span>
    </kbd>
  );
}
