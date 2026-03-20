import { useTranslation } from 'react-i18next';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../../styles/buttonStyles';
import {
  BJ_SUGGEST_DOUBLE,
  BJ_SUGGEST_DOUBLE_STAND,
  BJ_SUGGEST_HIT,
  BJ_SUGGEST_SPLIT,
  BJ_SUGGEST_STAND,
  BJ_SUGGEST_SURRENDER,
  highlightClass,
} from './bjConstants';

/** Props for BlackJack action phase control buttons. */
export interface BjActionPhaseControlsProps {
  loading: boolean;
  hintEnabled: boolean;
  suggestedAction: number;
  showDoubleDown: boolean;
  showSplit: boolean;
  showSurrender: boolean;
  onHit: () => void;
  onStand: () => void;
  onDoubleDown: () => void;
  onSplit: () => void;
  onSurrender: () => void;
}

/** Renders BlackJack action phase buttons (hit, stand, double, split, surrender). */
export function BjActionPhaseControls(props: BjActionPhaseControlsProps) {
  const { t } = useTranslation('blackjack');
  return (
    <>
      <button
        type="button"
        className={highlightClass(btnPrimary, props.suggestedAction === BJ_SUGGEST_HIT && props.hintEnabled)}
        disabled={props.loading}
        onClick={props.onHit}
      >
        {t('button.hit')}
      </button>
      <button
        type="button"
        className={highlightClass(btnPrimary, props.suggestedAction === BJ_SUGGEST_STAND && props.hintEnabled)}
        disabled={props.loading}
        onClick={props.onStand}
      >
        {t('button.stand')}
      </button>
      {props.showDoubleDown && (
        <button
          type="button"
          className={highlightClass(
            btnWarning,
            (props.suggestedAction === BJ_SUGGEST_DOUBLE || props.suggestedAction === BJ_SUGGEST_DOUBLE_STAND) &&
              props.hintEnabled,
          )}
          disabled={props.loading}
          onClick={props.onDoubleDown}
        >
          {t('button.doubleDown')}
        </button>
      )}
      {props.showSplit && (
        <button
          type="button"
          className={highlightClass(btnSuccess, props.suggestedAction === BJ_SUGGEST_SPLIT && props.hintEnabled)}
          disabled={props.loading}
          onClick={props.onSplit}
        >
          {t('button.split')}
        </button>
      )}
      {props.showSurrender && (
        <button
          type="button"
          className={highlightClass(btnDanger, props.suggestedAction === BJ_SUGGEST_SURRENDER && props.hintEnabled)}
          disabled={props.loading}
          onClick={props.onSurrender}
        >
          {t('button.surrender')}
        </button>
      )}
    </>
  );
}
