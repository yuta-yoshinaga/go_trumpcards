import { useTranslation } from 'react-i18next';
import { btnDanger, btnPrimary } from '../../styles/buttonStyles';
import { KbdBadge } from '../KbdBadge';
import { BJ_SUGGEST_STAND, BJ_SUGGEST_SURRENDER, highlightClass } from './bjConstants';

/** Props for BlackJack early surrender phase controls. */
export interface BjEarlySurrenderPhaseControlsProps {
  loading: boolean;
  hintEnabled: boolean;
  suggestedAction: number;
  onSurrender: () => void;
  onContinue: () => void;
}

/** Renders early surrender and continue buttons for BlackJack. */
export function BjEarlySurrenderPhaseControls(props: BjEarlySurrenderPhaseControlsProps) {
  const { t } = useTranslation('blackjack');
  const surHighlight = props.hintEnabled && props.suggestedAction === BJ_SUGGEST_SURRENDER;
  const contHighlight = props.hintEnabled && props.suggestedAction === BJ_SUGGEST_STAND;
  return (
    <div className="flex justify-center gap-2">
      <button
        type="button"
        className={highlightClass(btnDanger, surHighlight)}
        disabled={props.loading}
        onClick={props.onSurrender}
        aria-keyshortcuts="u"
      >
        {t('button.earlySurrender')}
        <KbdBadge label={t('kbd.earlySurrender')} />
      </button>
      <button
        type="button"
        className={highlightClass(btnPrimary, contHighlight)}
        disabled={props.loading}
        onClick={props.onContinue}
        aria-keyshortcuts="n"
      >
        {t('button.continue')}
        <KbdBadge label={t('kbd.continue')} />
      </button>
    </div>
  );
}
