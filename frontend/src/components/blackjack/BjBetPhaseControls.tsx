import { btnPrimary, btnSuccess, btnWarning } from '../../styles/buttonStyles';

const VALID_DECK_COUNTS = [1, 2, 4, 6, 8] as const;
const VALID_CPU_COUNTS = [0, 1, 2, 3] as const;

export interface BjBetPhaseControlsProps {
  betAmount: number;
  onBetAmountChange: (v: number) => void;
  deckCount: number;
  onDeckCountChange: (v: number) => void;
  cpuPlayerCount: number;
  onCpuPlayerCountChange: (v: number) => void;
  hintEnabled: boolean;
  onToggleHint: () => void;
  dealerHitsSoft17: boolean;
  onToggleSoft17: () => void;
  countingEnabled: boolean;
  onToggleCounting: () => void;
  loading: boolean;
  onBet: () => void;
}

export function BjBetPhaseControls(props: BjBetPhaseControlsProps) {
  return (
    <>
      <div className="flex items-center justify-center gap-2 mb-2">
        <label htmlFor="bj-bet-amount" className="text-white text-sm">
          ベット額:
        </label>
        <input
          id="bj-bet-amount"
          type="number"
          min={10}
          step={10}
          value={props.betAmount}
          onChange={(e) => props.onBetAmountChange(Number(e.target.value))}
          className="w-20 px-2 py-1 rounded text-sm"
          disabled={props.loading}
        />
      </div>
      <div className="flex items-center justify-center gap-2 mb-2">
        <label htmlFor="bj-deck-count" className="text-white text-sm">
          デッキ数:
        </label>
        <select
          id="bj-deck-count"
          value={props.deckCount}
          onChange={(e) => props.onDeckCountChange(Number(e.target.value))}
          className="px-2 py-1 rounded text-sm"
          disabled={props.loading}
        >
          {VALID_DECK_COUNTS.map((d) => (
            <option key={d} value={d}>
              {d}デッキ
            </option>
          ))}
        </select>
      </div>
      <div className="flex items-center justify-center gap-2 mb-2">
        <label htmlFor="bj-cpu-count" className="text-white text-sm">
          CPU人数:
        </label>
        <select
          id="bj-cpu-count"
          value={props.cpuPlayerCount}
          onChange={(e) => props.onCpuPlayerCountChange(Number(e.target.value))}
          className="px-2 py-1 rounded text-sm"
          disabled={props.loading}
        >
          {VALID_CPU_COUNTS.map((c) => (
            <option key={c} value={c}>
              {c}人
            </option>
          ))}
        </select>
      </div>
      <div className="flex items-center justify-center gap-2 mb-2 flex-wrap">
        <button
          type="button"
          className={props.hintEnabled ? btnSuccess : btnWarning}
          disabled={props.loading}
          onClick={props.onToggleHint}
        >
          ヒント {props.hintEnabled ? 'ON' : 'OFF'}
        </button>
        <button
          type="button"
          className={props.dealerHitsSoft17 ? btnSuccess : btnWarning}
          disabled={props.loading}
          onClick={props.onToggleSoft17}
        >
          {props.dealerHitsSoft17 ? 'H17' : 'S17'}
        </button>
        <button
          type="button"
          className={props.countingEnabled ? btnSuccess : btnWarning}
          disabled={props.loading}
          onClick={props.onToggleCounting}
        >
          カウント {props.countingEnabled ? 'ON' : 'OFF'}
        </button>
      </div>
      <button type="button" className={btnPrimary} disabled={props.loading} onClick={props.onBet}>
        ベット
      </button>
    </>
  );
}
