import { useEffect, useState } from 'react';

/** Props for {@link LiveAnnouncement}. */
export interface LiveAnnouncementProps {
  /** What to announce. Empty means nothing is pending. */
  message: string;
  /**
   * Interrupt whatever the screen reader is saying. Reserve it for a prompt the
   * player cannot act correctly without — an irreversible one-time choice, not
   * a status update.
   */
  assertive?: boolean;
}

/**
 * Screen-reader-only live region that announces `message` when it appears.
 *
 * **The region is mounted before it has anything to say.** A live region that
 * enters the DOM already holding its text is usually not announced at all:
 * assistive tech watches an existing region for changes, and a region and its
 * content arriving in the same commit is not a change. Holding the text back
 * by one commit is what makes the announcement happen, and it is the whole
 * reason this is a component rather than an `aria-live` attribute on the
 * visible element.
 *
 * The visible element stays as it is; this is an additional channel, so the
 * announcement never depends on how the visible copy is styled or positioned.
 */
export function LiveAnnouncement({ message, assertive = false }: LiveAnnouncementProps) {
  const [announced, setAnnounced] = useState('');

  useEffect(() => {
    setAnnounced(message);
  }, [message]);

  return (
    <div role="status" aria-live={assertive ? 'assertive' : 'polite'} aria-atomic="true" className="sr-only">
      {announced}
    </div>
  );
}
