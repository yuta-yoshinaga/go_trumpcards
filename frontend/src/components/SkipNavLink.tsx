/** Props for the SkipNavLink component. */
export interface SkipNavLinkProps {
  /** The id of the target element to skip to. */
  targetId: string;
  /** The visible label text for the skip link. */
  label: string;
}

/** Renders a visually hidden skip navigation link that becomes visible on focus (WCAG 2.4.1). */
export function SkipNavLink({ targetId, label }: SkipNavLinkProps) {
  return (
    <a
      href={`#${targetId}`}
      className="sr-only focus:not-sr-only focus:fixed focus:top-4 focus:left-4 focus:z-50 focus:bg-ds-accent focus:text-ds-text-on-accent focus:px-4 focus:py-2 focus:rounded"
    >
      {label}
    </a>
  );
}
