import { useTranslation } from 'react-i18next';

/** Props for BlackJack hand status badges. */
export interface HandStatusBadgesProps {
  busted: boolean;
  doubled: boolean;
  isBlackJack: boolean;
  surrendered: boolean;
}

/** Renders status badges (bust, double-down, blackjack, surrender) for a hand. */
export function HandStatusBadges({ busted, doubled, isBlackJack, surrendered }: HandStatusBadgesProps) {
  const { t } = useTranslation('blackjack');
  return (
    <>
      {busted && (
        <abbr title={t('status.bustTooltip')} className="ml-1">
          [{t('status.bust')}]
        </abbr>
      )}
      {doubled && (
        <abbr title={t('status.ddTooltip')} className="ml-1">
          [{t('status.dd')}]
        </abbr>
      )}
      {isBlackJack && (
        <abbr title={t('status.bjTooltip')} className="ml-1">
          [{t('status.bj')}]
        </abbr>
      )}
      {surrendered && (
        <abbr
          title={t('status.surTooltip')}
          className="ml-1 text-xs bg-ds-surface-elevated text-ds-text-primary px-1 rounded"
        >
          [{t('status.sur')}]
        </abbr>
      )}
    </>
  );
}
