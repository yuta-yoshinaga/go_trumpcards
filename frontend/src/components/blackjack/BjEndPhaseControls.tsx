import { btnPrimary } from '../../styles/buttonStyles';

export interface BjEndPhaseControlsProps {
  loading: boolean;
  onReset: () => void;
}

export function BjEndPhaseControls(props: BjEndPhaseControlsProps) {
  return (
    <button type="button" className={btnPrimary} disabled={props.loading} onClick={props.onReset}>
      リセット
    </button>
  );
}
