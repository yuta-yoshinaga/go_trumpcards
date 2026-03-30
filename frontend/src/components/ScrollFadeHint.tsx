/** Right-edge gradient overlay indicating horizontally scrollable content. Hidden on desktop (sm+). */
export function ScrollFadeHint() {
  return (
    <div
      className="pointer-events-none absolute right-0 top-0 bottom-0 w-8 bg-gradient-to-l from-black/50 to-transparent sm:hidden"
      aria-hidden="true"
    />
  );
}
