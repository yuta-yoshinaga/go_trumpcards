import { useTranslation } from 'react-i18next';
import { btnDanger, btnWarning } from '../../styles/buttonStyles';
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
      <button type="button" className={btnWarning} disabled={props.loading} onClick={props.onInsurance}>
        {t('button.insurance')}
      </button>
      <button
        type="button"
        className={highlightClass(
          btnDanger,
          props.suggestedAction === BJ_SUGGEST_DECLINE_INSURANCE && props.hintEnabled,
        )}
        disabled={props.loading}
        onClick={props.onDecline}
      >
        {t('button.decline')}
      </button>
    </>
  );
}
