import { useTranslation } from 'react-i18next';

/** One player's round summary entry. */
export interface RoundScoreEntry {
  name: string;
  roundScore: number;
  cumulativeScore: number;
}

/** Props for {@link RoundScoreAnnouncement}. */
export interface RoundScoreAnnouncementProps {
  active: boolean;
  entries: RoundScoreEntry[];
  /** Separator used between entries; defaults to the shared English separator. */
  separator?: string;
}

/**
 * Screen-reader-only live region that announces round score deltas when a
 * round ends. The visible score table remains unchanged; this component is
 * the announcement channel for assistive tech.
 */
export function RoundScoreAnnouncement({ active, entries, separator = ', ' }: RoundScoreAnnouncementProps) {
  const { t } = useTranslation('common');
  const details = entries
    .map((e) =>
      t('roundScoreAnnouncement.entry', {
        name: e.name,
        round: e.roundScore,
        total: e.cumulativeScore,
      }),
    )
    .join(separator);
  return (
    <div role="status" aria-live="polite" aria-atomic="true" className="sr-only">
      {active ? t('roundScoreAnnouncement.message', { details }) : ''}
    </div>
  );
}
