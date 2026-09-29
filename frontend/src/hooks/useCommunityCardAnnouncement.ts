import type { TFunction } from 'i18next';
import { useEffect, useRef, useState } from 'react';
import type { Card } from '../types/common';
import { cardAlt } from '../utils/cardAlt';

/** Announces newly revealed community cards and clears the announcement when a hand resets. */
export function useCommunityCardAnnouncement(communityCards: Card[], t: TFunction): string {
  const [announcement, setAnnouncement] = useState('');
  const announcedCount = useRef(0);

  useEffect(() => {
    if (communityCards.length < announcedCount.current) {
      announcedCount.current = communityCards.length;
      setAnnouncement('');
      return;
    }
    if (communityCards.length > announcedCount.current) {
      setAnnouncement(
        t('communityCardsRevealed', {
          cards: communityCards.slice(announcedCount.current).map(cardAlt).join(t('listSeparator')),
        }),
      );
      announcedCount.current = communityCards.length;
    }
  }, [communityCards, t]);

  return announcement;
}
