/** Props for the LandscapeBanner component. */
interface LandscapeBannerProps {
  /** The message to display in the banner. */
  message: string;
}

/** Banner suggesting landscape orientation on small portrait screens. */
export function LandscapeBanner({ message }: LandscapeBannerProps) {
  return (
    <div className="hidden portrait:flex sm:hidden items-center gap-2 px-4 py-2 bg-yellow-500/90 text-black text-sm font-medium">
      <span aria-hidden="true">&#8635;</span>
      <span>{message}</span>
    </div>
  );
}
