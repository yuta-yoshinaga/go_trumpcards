import { useEffect } from 'react';
import { SITE_NAME } from '../constants/site';

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
    document.title = title ? `${title} - ${SITE_NAME}` : SITE_NAME;
    return () => {
      document.title = SITE_NAME;
    };
  }, [title]);
}
