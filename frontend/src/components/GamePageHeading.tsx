/** Renders a visually-hidden h1 heading for the current game page (WCAG 2.4.6). */
export function GamePageHeading({ title }: { title: string }) {
  return <h1 className="sr-only">{title}</h1>;
}
