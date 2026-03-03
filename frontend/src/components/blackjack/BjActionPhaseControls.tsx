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

export function BjActionPhaseControls(props: BjActionPhaseControlsProps) {
  return (
    <>
      <button
        type="button"
        className={highlightClass(btnPrimary, props.suggestedAction === BJ_SUGGEST_HIT && props.hintEnabled)}
        disabled={props.loading}
        onClick={props.onHit}
      >
        ヒット
      </button>
      <button
        type="button"
        className={highlightClass(btnPrimary, props.suggestedAction === BJ_SUGGEST_STAND && props.hintEnabled)}
        disabled={props.loading}
        onClick={props.onStand}
      >
        スタンド
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
          ダブルダウン
        </button>
      )}
      {props.showSplit && (
        <button
          type="button"
          className={highlightClass(btnSuccess, props.suggestedAction === BJ_SUGGEST_SPLIT && props.hintEnabled)}
          disabled={props.loading}
          onClick={props.onSplit}
        >
          スプリット
        </button>
      )}
      {props.showSurrender && (
        <button
          type="button"
          className={highlightClass(btnDanger, props.suggestedAction === BJ_SUGGEST_SURRENDER && props.hintEnabled)}
          disabled={props.loading}
          onClick={props.onSurrender}
        >
          サレンダー
        </button>
      )}
    </>
  );
}
