import { useTranslation } from 'react-i18next';
import { btnPrimary, btnSuccess, btnWarning } from '../../styles/buttonStyles';
import {
  BJ_COUNTING_HILO,
  BJ_COUNTING_KO,
  BJ_COUNTING_OMEGA2,
  BJ_COUNTING_ZEN,
  BJ_VALID_PENETRATIONS,
} from './bjConstants';

const VALID_DECK_COUNTS = [1, 2, 4, 6, 8] as const;
const VALID_CPU_COUNTS = [0, 1, 2, 3] as const;
const VALID_HAND_COUNTS = [1, 2, 3] as const;
const COUNTING_SYSTEMS = [BJ_COUNTING_HILO, BJ_COUNTING_KO, BJ_COUNTING_ZEN, BJ_COUNTING_OMEGA2] as const;
const VALID_SURRENDER_RULES = [0, 1, 2] as const;

/** Props for BlackJack bet phase controls. */
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
  doubleAfterSplit: boolean;
  onToggleDAS: () => void;
  countingSystem: number;
  onCountingSystemChange: (v: number) => void;
  deckPenetration: number;
  onDeckPenetrationChange: (v: number) => void;
  handCount: number;
  onHandCountChange: (v: number) => void;
  surrenderRule: number;
  onSurrenderRuleChange: (v: number) => void;
  loading: boolean;
  onBet: () => void;
  perfectPairsBet: number;
  onPerfectPairsBetChange: (v: number) => void;
  twentyOnePlus3Bet: number;
  onTwentyOnePlus3BetChange: (v: number) => void;
  /** When true, the advanced settings section is expanded by default (e.g. on desktop). */
  autoExpandAdvanced?: boolean;
}

/** Renders BlackJack bet phase controls with basic settings and collapsible advanced options. */
export function BjBetPhaseControls(props: BjBetPhaseControlsProps) {
  const { t } = useTranslation('blackjack');
  return (
    <>
      {/* Basic settings: bet amount, hand count */}
      <div className="flex items-center justify-center gap-2 mb-2">
        <label htmlFor="bj-bet-amount" className="text-white text-sm">
          {t('betAmount')}
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
        <label htmlFor="bj-hand-count" className="text-white text-sm">
          {t('handCount')}
        </label>
        <select
          id="bj-hand-count"
          value={props.handCount}
          onChange={(e) => props.onHandCountChange(Number(e.target.value))}
          className="px-2 py-1 rounded text-sm"
          disabled={props.loading}
        >
          {VALID_HAND_COUNTS.map((h) => (
            <option key={h} value={h}>
              {h}
              {t('handCountUnit')}
            </option>
          ))}
        </select>
      </div>

      {/* Advanced settings: collapsible */}
      <details className="mb-2 text-white text-sm" open={props.autoExpandAdvanced || undefined}>
        <summary className="cursor-pointer select-none text-center text-yellow-300 hover:text-yellow-200 py-1">
          {t('advancedSettings')}
          {(props.perfectPairsBet > 0 || props.twentyOnePlus3Bet > 0) && (
            <span className="ml-2 inline-block bg-yellow-500 text-black text-xs font-bold px-1.5 py-0.5 rounded-full">
              {t('sideBetActive')}
            </span>
          )}
        </summary>
        <div className="mt-2 space-y-2 glass-panel rounded-lg p-3">
          {/* Side bets */}
          <div className="flex items-center justify-center gap-2 flex-wrap">
            <label htmlFor="bj-pp-bet" className="text-white text-sm" title={t('sideBetTooltip.pp')}>
              {t('sideBetLabel.pp')}
            </label>
            <input
              id="bj-pp-bet"
              type="number"
              min={0}
              max={10000}
              step={10}
              value={props.perfectPairsBet}
              onChange={(e) => props.onPerfectPairsBetChange(Number(e.target.value))}
              className="w-20 px-2 py-1 rounded text-sm"
              disabled={props.loading}
            />
            <label htmlFor="bj-t3-bet" className="text-white text-sm" title={t('sideBetTooltip.t3')}>
              {t('sideBetLabel.t3')}
            </label>
            <input
              id="bj-t3-bet"
              type="number"
              min={0}
              max={10000}
              step={10}
              value={props.twentyOnePlus3Bet}
              onChange={(e) => props.onTwentyOnePlus3BetChange(Number(e.target.value))}
              className="w-20 px-2 py-1 rounded text-sm"
              disabled={props.loading}
            />
          </div>
          {/* Deck & CPU count */}
          <div className="flex items-center justify-center gap-2 flex-wrap">
            <label htmlFor="bj-deck-count" className="text-white text-sm">
              {t('deckCount')}
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
                  {d}
                  {t('deckUnit')}
                </option>
              ))}
            </select>
            <label htmlFor="bj-cpu-count" className="text-white text-sm">
              {t('cpuCount')}
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
                  {c}
                  {t('cpuUnit')}
                </option>
              ))}
            </select>
          </div>
          {/* Toggle buttons & selects */}
          <div className="flex items-center justify-center gap-2 flex-wrap">
            <button
              type="button"
              className={props.hintEnabled ? btnSuccess : btnWarning}
              disabled={props.loading}
              onClick={props.onToggleHint}
            >
              {t('hint')} {props.hintEnabled ? 'ON' : 'OFF'}
            </button>
            <button
              type="button"
              className={props.dealerHitsSoft17 ? btnSuccess : btnWarning}
              disabled={props.loading}
              onClick={props.onToggleSoft17}
              title={t('soft17Tooltip')}
            >
              {props.dealerHitsSoft17 ? 'H17' : 'S17'}
            </button>
            <button
              type="button"
              className={props.countingEnabled ? btnSuccess : btnWarning}
              disabled={props.loading}
              onClick={props.onToggleCounting}
            >
              {t('counting')} {props.countingEnabled ? 'ON' : 'OFF'}
            </button>
            <select
              aria-label={t('countingSystem')}
              value={props.countingSystem}
              onChange={(e) => props.onCountingSystemChange(Number(e.target.value))}
              className="px-2 py-1 rounded text-sm"
              disabled={props.loading || !props.countingEnabled}
            >
              {COUNTING_SYSTEMS.map((cs) => (
                <option key={cs} value={cs}>
                  {t(`countingSystemNames.${cs}`)}
                </option>
              ))}
            </select>
          </div>
          <div className="flex items-center justify-center gap-2 flex-wrap">
            <button
              type="button"
              className={props.doubleAfterSplit ? btnSuccess : btnWarning}
              disabled={props.loading}
              onClick={props.onToggleDAS}
              title={t('dasTooltip')}
            >
              {t('das')} {props.doubleAfterSplit ? 'ON' : 'OFF'}
            </button>
            <label htmlFor="bj-penetration" className="text-white text-sm">
              {t('penetration')}
            </label>
            <select
              id="bj-penetration"
              value={props.deckPenetration}
              onChange={(e) => props.onDeckPenetrationChange(Number(e.target.value))}
              className="px-2 py-1 rounded text-sm"
              disabled={props.loading}
            >
              {BJ_VALID_PENETRATIONS.map((p) => (
                <option key={p} value={p}>
                  {p}%
                </option>
              ))}
            </select>
            <label htmlFor="bj-surrender-rule" className="text-white text-sm">
              {t('surrenderRule')}
            </label>
            <select
              id="bj-surrender-rule"
              value={props.surrenderRule}
              onChange={(e) => props.onSurrenderRuleChange(Number(e.target.value))}
              className="px-2 py-1 rounded text-sm"
              disabled={props.loading}
            >
              {VALID_SURRENDER_RULES.map((r) => (
                <option key={r} value={r}>
                  {t(`surrenderRuleNames.${r}`)}
                </option>
              ))}
            </select>
          </div>
        </div>
      </details>

      <button
        type="button"
        className={btnPrimary}
        disabled={props.loading}
        onClick={props.onBet}
        data-tutorial="bj-bet-button"
      >
        {t('button.bet')}
      </button>
    </>
  );
}
