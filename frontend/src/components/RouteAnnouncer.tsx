import { useEffect, useRef, useState } from 'react';
import { useLocation } from 'react-router-dom';
import { subscribeDocumentTitle } from '../hooks/useDocumentTitle';

/** The id of the landmark that receives focus after a navigation. */
const MAIN_ID = 'main-content';

/**
 * Announces route changes and moves focus to the main landmark.
 *
 * A HashRouter navigation fires no page-load event, so assistive tech gets no
 * signal that the view changed: switching between 318 games was silent, and
 * focus stayed on the link that was clicked, leaving keyboard users to tab past
 * the whole navigation again to reach the board. See issue #5360.
 *
 * The two halves key off different signals on purpose:
 *
 * - **Focus** follows the pathname. It must move as soon as the route changes,
 *   and it does not care what the destination is called.
 * - **The announcement** waits for `useDocumentTitle` to publish. Reading
 *   `document.title` inside the pathname effect looks equivalent and is not:
 *   this component is declared before `<Routes>`, and React flushes sibling
 *   effects in declaration order, so it runs *before* the destination page's
 *   `useDocumentTitle` and captures the bare site name that the unmounting
 *   page's cleanup left behind. Lazy route chunks widen the same gap — the new
 *   page may not have mounted at all yet. Subscribing instead makes the
 *   announcement independent of when the page gets around to setting it.
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
  const focusedPath = useRef(pathname);
  // Set when a navigation happens, cleared once that navigation is announced,
  // so a title change that is not a navigation (a page renaming itself mid-play)
  // stays silent.
  const pendingAnnounce = useRef(false);

  useEffect(() => {
    if (focusedPath.current === pathname) return;
    focusedPath.current = pathname;
    pendingAnnounce.current = true;
    document.getElementById(MAIN_ID)?.focus({ preventScroll: true });
  }, [pathname]);

  useEffect(() => {
    // Subscribed once for the component's lifetime: the announcement is driven
    // by the title landing, not by this effect re-running.
    return subscribeDocumentTitle((title) => {
      if (!pendingAnnounce.current) return;
      pendingAnnounce.current = false;
      setMessage(title);
    });
  }, []);

  return (
    <div role="status" aria-live="polite" className="sr-only">
      {message}
    </div>
  );
}
