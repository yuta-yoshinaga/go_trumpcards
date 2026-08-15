import { useEffect } from 'react';
import { SITE_NAME } from '../constants/site';

/** Notified with the full document title whenever a page sets one. */
type TitleListener = (title: string) => void;

const listeners = new Set<TitleListener>();

/**
 * Subscribes to page-title changes. Returns an unsubscribe function.
 *
 * `RouteAnnouncer` uses this instead of reading `document.title` after a
 * navigation. Reading it directly looks equivalent and is not: the announcer is
 * declared before `<Routes>`, React flushes sibling effects in declaration
 * order, so it runs *before* the destination page's `useDocumentTitle` and sees
 * the bare site name that the unmounting page's cleanup left behind. Lazy route
 * chunks widen the same gap. Publishing from the setter removes the ordering
 * question entirely — the announcement happens when the title actually lands,
 * however late that is. See issue #5360.
 *
 * The unmount cleanup deliberately does *not* publish: resetting to the bare
 * site name is not a page identity worth announcing.
 */
export function subscribeDocumentTitle(listener: TitleListener): () => void {
  listeners.add(listener);
  return () => {
    listeners.delete(listener);
  };
}

/**
 * Sets `document.title` for the lifetime of the calling component.
 *
 * Extracted from `useGamePageSetup` so pages that are not games can identify
 * themselves too. Only game pages managed the title before, so navigating to
 * `/discover`, `/legal` or a 404 left the tab showing the bare site name with
 * no page identity — and in a HashRouter SPA there is no page-load event to
 * fall back on. See issue #5360.
 *
 * An empty `title` yields the bare site name rather than a dangling
 * `" - Trump Cards"`, which is what a page whose translation has not resolved
 * yet would otherwise render.
 */
export function useDocumentTitle(title: string): void {
  useEffect(() => {
    const full = title ? `${title} - ${SITE_NAME}` : SITE_NAME;
    document.title = full;
    for (const listener of listeners) listener(full);
    return () => {
      document.title = SITE_NAME;
    };
  }, [title]);
}
