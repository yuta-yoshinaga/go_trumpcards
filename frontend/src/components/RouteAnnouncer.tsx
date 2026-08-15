import { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';

/** The id of the landmark that receives focus after a navigation. */
const MAIN_ID = 'main-content';

/**
 * Announces route changes and moves focus to the main landmark.
 *
 * A HashRouter navigation fires no page-load event, so assistive tech gets no
 * signal that the view changed: switching between 318 games was silent, and
 * focus stayed on the link that was clicked, leaving keyboard users to tab past
 * the whole navigation again to reach the board. `document.title` changes are
 * not a reliable substitute — without a load event, whether a screen reader
 * re-reads the title is implementation-dependent. See issue #5360.
 *
 * Deliberately one announcement per navigation, matching the granularity of the
 * existing search-result-count live regions. Per-move announcements during play
 * would drown the useful ones.
 *
 * `<main>` already carries `tabIndex={-1}` for the skip link, so it can receive
 * programmatic focus here without new markup. Focus is set with
 * `preventScroll` so the announcement does not also jump the viewport.
 */
export function RouteAnnouncer() {
  const { pathname } = useLocation();
  const [message, setMessage] = useState('');
  // Seeded with the initial path so the first render is a no-op: an ordinary
  // page load is already announced by assistive tech, and stealing focus would
  // fight a deep link that scrolled somewhere specific. Comparing paths rather
  // than tracking a boolean also keeps a re-render on the same route silent.
  const announcedPath = useRef(pathname);

  useEffect(() => {
    if (announcedPath.current === pathname) return;
    announcedPath.current = pathname;
    document.getElementById(MAIN_ID)?.focus({ preventScroll: true });
    // The title is set by useDocumentTitle on the page that just mounted, so it
    // is the page's own name by the time this effect runs.
    setMessage(document.title);
  }, [pathname]);

  return (
    <div role="status" aria-live="polite" className="sr-only">
      {message}
    </div>
  );
}
