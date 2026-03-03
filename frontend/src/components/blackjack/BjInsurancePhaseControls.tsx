import { btnDanger, btnWarning } from '../../styles/buttonStyles';
import { BJ_SUGGEST_DECLINE_INSURANCE, highlightClass } from './bjConstants';

export interface BjInsurancePhaseControlsProps {
  loading: boolean;
  hintEnabled: boolean;
  suggestedAction: number;
  onInsurance: () => void;
  onDecline: () => void;
}

export function BjInsurancePhaseControls(props: BjInsurancePhaseControlsProps) {
  return (
    <>
      <button type="button" className={btnWarning} disabled={props.loading} onClick={props.onInsurance}>
        インシュランス
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
        辞退
      </button>
    </>
  );
}
