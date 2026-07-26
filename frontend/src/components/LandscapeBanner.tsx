/** Props for the LandscapeBanner component. */
export interface LandscapeBannerProps {
  /** The message to display in the banner. */
  message: string;
}

/** Banner suggesting landscape orientation on small portrait screens. */
export function LandscapeBanner({ message }: LandscapeBannerProps) {
  return (
    <div className="hidden portrait:flex sm:hidden items-center gap-2 px-4 py-2 bg-ds-warning/90 text-ds-text-on-accent text-sm font-medium">
      <svg
        xmlns="http://www.w3.org/2000/svg"
        width="20"
        height="20"
        viewBox="0 0 24 24"
        fill="none"
        stroke="currentColor"
        strokeWidth="2"
        strokeLinecap="round"
        strokeLinejoin="round"
        aria-hidden="true"
        className="flex-shrink-0 animate-pulse-once"
      >
        <rect x="4" y="2" width="16" height="20" rx="2" ry="2" />
        <line x1="12" y1="18" x2="12.01" y2="18" />
      </svg>
      <span>{message}</span>
    </div>
  );
}
