import { useTranslation } from 'react-i18next';
import { btnDanger, btnWarning } from '../../styles/buttonStyles';
import { KbdBadge } from '../KbdBadge';
import { BJ_SUGGEST_DECLINE_INSURANCE, highlightClass } from './bjConstants';

/** Props for BlackJack insurance phase controls. */
export interface BjInsurancePhaseControlsProps {
  loading: boolean;
  hintEnabled: boolean;
  suggestedAction: number;
  onInsurance: () => void;
  onDecline: () => void;
}

/** Renders insurance and decline buttons for BlackJack insurance phase. */
export function BjInsurancePhaseControls(props: BjInsurancePhaseControlsProps) {
  const { t } = useTranslation('blackjack');
  return (
    <>
      <button
        type="button"
        className={btnWarning}
        disabled={props.loading}
        onClick={props.onInsurance}
        aria-keyshortcuts="i"
      >
        {t('button.insurance')}
        <KbdBadge label={t('kbd.insurance')} />
      </button>
      <button
        type="button"
        className={highlightClass(
          btnDanger,
          props.suggestedAction === BJ_SUGGEST_DECLINE_INSURANCE && props.hintEnabled,
        )}
        disabled={props.loading}
        onClick={props.onDecline}
        aria-keyshortcuts="n"
      >
        {t('button.decline')}
        <KbdBadge label={t('kbd.decline')} />
      </button>
    </>
  );
}
