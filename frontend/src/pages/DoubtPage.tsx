import { useCallback, useEffect, useRef, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { doubtApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { CpuTurnArea } from '../components/CpuTurnArea';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { useCardSelection } from '../hooks/useCardSelection';
import { useGameApi } from '../hooks/useGameApi';
import { btnDanger, btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { playerAreaBase } from '../styles/gameStyles';
import type { Card, DoubtConfig, DoubtCpuAction, DoubtPlayerData } from '../types/card';
import { valueName } from '../utils/cardUtils';
import { playerName } from '../utils/playerUtils';

const playerAreaClass = `${playerAreaBase} p-[10px] flex-[1_1_150px] min-w-[120px]`;

// ── CPU player area ──────────────────────────────────────────────────────────

function DoubtCpuArea({
  player,
  isCurrentTurn,
  hasTell,
}: {
  player: DoubtPlayerData;
  isCurrentTurn: boolean;
  hasTell: boolean;
}) {
  const { t } = useTranslation('doubt');
  const { t: tc } = useTranslation('common');
  return (
    <CpuTurnArea
      playerId={player.id}
      isHuman={player.isHuman}
      isCurrentTurn={isCurrentTurn}
      isFinished={player.isFinished}
      dimFinished={false}
      finishedLabel={player.isFinished ? tc('status.finished') : undefined}
      className={playerAreaClass}
      nameClassName="text-sm"
    >
      <div className="text-[#ccc] text-[0.85em]">{t('cardCount', { count: player.cardCount })}</div>
      {hasTell && (
        <span className="animate-sweat-drop text-lg" role="img" aria-label={t('tell')}>
          💧
        </span>
      )}
    </CpuTurnArea>
  );
}

// ── Human hand card ───────────────────────────────────────────────────────────

interface HandCardProps {
  card: Card;
  index: number;
  selected: boolean;
  selectable: boolean;
  onToggle: (idx: number) => void;
}

function HandCard({ card, index, selected, selectable, onToggle }: HandCardProps) {
  return (
    <button
      type="button"
      data-testid="hand-card"
      aria-pressed={selected}
      disabled={!selectable}
      onClick={() => onToggle(index)}
      style={{
        background: 'none',
        padding: 0,
        cursor: selectable ? 'pointer' : 'default',
        borderRadius: 8,
        border: selected ? '3px solid #5cb85c' : '3px solid transparent',
        transform: selected ? 'translateY(-8px)' : 'none',
        transition: 'transform 0.15s, border 0.15s',
        opacity: !selectable ? 0.5 : 1,
        boxSizing: 'border-box',
      }}
    >
      <CardImage card={card} width={52} />
    </button>
  );
}

// ── Action info ───────────────────────────────────────────────────────────────

function actionDesc(
  action: DoubtCpuAction,
  players: DoubtPlayerData[],
  t: (key: string, opts?: Record<string, unknown>) => string,
): string {
  const p = players[action.playerIdx];
  const name = p ? playerName(p.id, p.isHuman) : `Player ${action.playerIdx}`;
  return t('actionDesc', { name, count: action.cardCount, value: valueName(action.claimedValue) });
}

// ── Main page ────────────────────────────────────────────────────────────────

const DEFAULT_DOUBT_CONFIG: DoubtConfig = { doubtWindowSec: 10, cpuMemoryLevel: 1, penaltyDrawLimit: 0 };

const DOUBT_WINDOW_OPTIONS = [3, 5, 10] as const;

const CPU_MEMORY_OPTIONS = [
  { value: 0, label: 'Easy' },
  { value: 1, label: 'Normal' },
  { value: 2, label: 'Hard' },
] as const;

const PENALTY_DRAW_LIMIT_OPTIONS = [0, 3, 5, 10] as const;

export function DoubtPage() {
  const { t } = useTranslation('doubt');
  const { t: tc } = useTranslation('common');
  const { selected: selectedCardIndices, toggle: toggleCard, clear: clearSelection } = useCardSelection();
  const [claimedValue, setClaimedValue] = useState(1);
  const [countdown, setCountdown] = useState<number | null>(null);
  const [doubtConfig, setDoubtConfig] = useState<DoubtConfig>(DEFAULT_DOUBT_CONFIG);
  const countdownRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const autoSkipRef = useRef(false);
  const cpuDoubtersRef = useRef<number[]>([]);

  const onSuccess = useCallback(() => {
    clearSelection();
    setClaimedValue(1);
  }, [clearSelection]);
  const { state, loading, error, exec: rawExec } = useGameApi(doubtApi.exec, { onSuccess });

  // Keep cpuDoubters ref current so the auto-skip effect avoids stale state
  useEffect(() => {
    if (state) cpuDoubtersRef.current = state.cpuDoubters;
  }, [state]);

  const stopCountdown = useCallback(() => {
    autoSkipRef.current = false;
    if (countdownRef.current !== null) {
      clearInterval(countdownRef.current);
      countdownRef.current = null;
    }
    setCountdown(null);
  }, []);

  const startCountdown = useCallback(
    (sec: number) => {
      stopCountdown();
      setCountdown(sec);
      countdownRef.current = setInterval(() => {
        setCountdown((prev) => {
          const cur = prev as number;
          if (cur <= 1) {
            clearInterval(countdownRef.current as ReturnType<typeof setInterval>);
            countdownRef.current = null;
            autoSkipRef.current = true;
            return null;
          }
          return cur - 1;
        });
      }, 1000);
    },
    [stopCountdown],
  );

  const exec = useCallback(
    (...args: Parameters<typeof rawExec>) => {
      stopCountdown();
      return rawExec(...args);
    },
    [rawExec, stopCountdown],
  );

  const handleConfigChange = useCallback((key: keyof DoubtConfig, value: string) => {
    const parsed = Number(value);
    if (!Number.isNaN(parsed)) {
      setDoubtConfig((prev) => ({ ...prev, [key]: parsed }));
    }
  }, []);

  useEffect(() => {
    exec('reset', undefined, undefined, undefined, DEFAULT_DOUBT_CONFIG);
  }, [exec]);

  // Auto-skip when countdown timer expires without user interaction (mirrors CLI timeout behaviour)
  useEffect(() => {
    if (countdown !== null) return;
    if (!autoSkipRef.current) return;
    autoSkipRef.current = false;
    exec('skip', undefined, undefined, cpuDoubtersRef.current);
  }, [countdown, exec]);

  // Start countdown when it's the doubt phase and a CPU played (human needs to decide)
  useEffect(() => {
    if (!state) return;
    if (state.phase === 1 && state.lastAction !== null) {
      const lastActionPlayer = state.players[state.lastAction.playerIdx];
      if (lastActionPlayer && !lastActionPlayer.isHuman) {
        startCountdown(state.doubtWindowSec);
      }
    }
  }, [state, startCountdown]);

  if (!state) return null;

  const isHumanTurn = !state.gameEndFlag && state.players[state.currentTurn]?.isHuman === true;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const isDoubtPhase = state.phase === 1;
  const cpuPlayed = isDoubtPhase && state.lastAction !== null && !state.players[state.lastAction.playerIdx]?.isHuman;

  const cpuTells = new Set(
    [...state.cpuActions, state.lastAction]
      .filter((a): a is DoubtCpuAction => a !== null && a.hasTell === true)
      .map((a) => a.playerIdx),
  );

  const handlePlay = () => {
    exec('play', selectedCardIndices, claimedValue);
  };

  const handleDoubt = () => {
    stopCountdown();
    exec('doubt', undefined, undefined, [0, ...state.cpuDoubters]);
  };

  const handleSkip = () => {
    stopCountdown();
    exec('skip', undefined, undefined, state.cpuDoubters);
  };

  const handleCpuDoubtConfirm = () => {
    exec('doubt', undefined, undefined, state.cpuDoubters);
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a2c5c]" aria-busy={loading} aria-live="polite">
      {loading && <span className="sr-only">{tc('status.loading')}</span>}
      {/* Settings panel */}
      <details className="px-4 pt-2">
        <summary className="text-white/70 text-xs cursor-pointer select-none">{t('settings.title')}</summary>
        <div className="bg-black/30 rounded-lg p-3 mt-1 flex flex-wrap gap-4 text-sm text-white">
          <label className="flex items-center gap-2">
            {t('settings.doubtTime')}
            <select
              className="bg-black/50 text-white rounded px-2 py-1 border border-white/30"
              value={doubtConfig.doubtWindowSec}
              onChange={(e) => handleConfigChange('doubtWindowSec', e.target.value)}
            >
              {DOUBT_WINDOW_OPTIONS.map((sec) => (
                <option key={sec} value={sec}>
                  {t('settings.sec', { sec })}
                </option>
              ))}
            </select>
          </label>
          <label className="flex items-center gap-2">
            {t('settings.cpuMemory')}
            <select
              className="bg-black/50 text-white rounded px-2 py-1 border border-white/30"
              value={doubtConfig.cpuMemoryLevel}
              onChange={(e) => handleConfigChange('cpuMemoryLevel', e.target.value)}
            >
              {CPU_MEMORY_OPTIONS.map((opt) => (
                <option key={opt.value} value={opt.value}>
                  {opt.label}
                </option>
              ))}
            </select>
          </label>
          <label className="flex items-center gap-2">
            {t('settings.penaltyDrawLimit')}
            <select
              className="bg-black/50 text-white rounded px-2 py-1 border border-white/30"
              value={doubtConfig.penaltyDrawLimit}
              onChange={(e) => handleConfigChange('penaltyDrawLimit', e.target.value)}
            >
              {PENALTY_DRAW_LIMIT_OPTIONS.map((v) => (
                <option key={v} value={v}>
                  {v === 0 ? t('settings.unlimited') : t('settings.cards', { count: v })}
                </option>
              ))}
            </select>
          </label>
        </div>
      </details>

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* CPU player areas */}
        <div className="flex gap-2 flex-wrap mb-3">
          {cpuPlayers.map((player) => (
            <DoubtCpuArea
              key={player.id}
              player={player}
              isCurrentTurn={state.currentTurn === player.id}
              hasTell={cpuTells.has(player.id)}
            />
          ))}
        </div>

        {/* Table area */}
        <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2">
          <div className="text-white font-bold mb-1">{t('table')}</div>
          <div className="text-[#ccc] text-[0.9em]">{t('tableCards', { count: state.tableCardCount })}</div>
          {state.lastAction && (
            <div className="text-yellow-300 text-[0.85em] mt-1">{actionDesc(state.lastAction, state.players, t)}</div>
          )}
        </div>

        {/* Doubt/Skip UI */}
        {isDoubtPhase && !state.gameEndFlag && (
          <div className="bg-black/40 rounded-[10px] py-3 px-4 my-2">
            {cpuPlayed ? (
              <>
                <div className="text-white font-bold mb-2">{t('doubtQuestion')}</div>
                {countdown !== null && (
                  <div className="text-yellow-300 text-lg font-bold mb-2">{t('countdown', { sec: countdown })}</div>
                )}
                {state.cpuDoubters.length > 0 && (
                  <div className="text-[#ccc] text-[0.85em] mb-2">
                    {t('cpuDoubters', { names: state.cpuDoubters.map((idx) => playerName(idx, false)).join(', ') })}
                  </div>
                )}
                <div className="flex gap-2">
                  <button type="button" className={btnDanger} disabled={loading} onClick={handleDoubt}>
                    {t('doubtButton')}
                  </button>
                  <button type="button" className={btnWarning} disabled={loading} onClick={handleSkip}>
                    {t('skipButton')}
                  </button>
                </div>
              </>
            ) : (
              <>
                <div className="text-white font-bold mb-2">{t('cpuJudging')}</div>
                {state.cpuDoubters.length > 0 && (
                  <div className="text-red-300 text-[0.9em] mb-2">
                    {t('cpuDoubtExclaim', { names: state.cpuDoubters.map((idx) => playerName(idx, false)).join(', ') })}
                  </div>
                )}
                <button type="button" className={btnPrimary} disabled={loading} onClick={handleCpuDoubtConfirm}>
                  {t('confirmButton')}
                </button>
              </>
            )}
          </div>
        )}

        {/* Last doubt result */}
        {state.lastDoubtResult && (
          <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-[0.85em]">
            <div className="text-white font-bold mb-1">{t('doubtResult.title')}</div>
            <div className={state.lastDoubtResult.wasLying ? 'text-red-300' : 'text-green-300'}>
              {state.lastDoubtResult.wasLying ? t('doubtResult.wasLying') : t('doubtResult.wasTruth')}
            </div>
            <div className="text-[#ccc]">
              {t('doubtResult.loserTook', {
                name: playerName(
                  state.players[state.lastDoubtResult.loserIdx]?.id ?? state.lastDoubtResult.loserIdx,
                  state.players[state.lastDoubtResult.loserIdx]?.isHuman ?? false,
                ),
                count: state.lastDoubtResult.cardCount,
              })}
            </div>
            {state.lastDoubtResult.discardedCount > 0 && (
              <div className="text-yellow-300">
                {t('doubtResult.discarded', { count: state.lastDoubtResult.discardedCount })}
              </div>
            )}
            {state.lastDoubtResult.revealedCards.length > 0 && (
              <div className="flex flex-wrap gap-1 mt-1">
                {state.lastDoubtResult.revealedCards.map((card, i) => (
                  <CardImage key={`${card.design}-${card.value}-${i}`} card={card} width={36} />
                ))}
              </div>
            )}
          </div>
        )}

        {/* Human/CPU action logs */}
        {state.humanAction && !isDoubtPhase && (
          <div className="bg-black/40 rounded-lg text-[#cfc] py-2 px-3.5 my-2 text-[0.85em]">
            {actionDesc(state.humanAction, state.players, t)}
          </div>
        )}
        {state.cpuActions && state.cpuActions.length > 0 && (
          <div className="bg-black/40 rounded-lg text-[#ccc] py-2 px-3.5 my-2 whitespace-pre-line text-[0.85em]">
            {[tc('label.cpuActions'), ...state.cpuActions.map((a) => actionDesc(a, state.players, t))].join('\n')}
          </div>
        )}

        {/* Result message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
      </div>

      {/* Sticky footer: human player hand + action buttons */}
      <GameFooter className="bg-[#101c3a] border-white/20 px-4 py-2.5">
        {/* Human player info */}
        {humanPlayer && (
          <div className="mb-2">
            <div className="text-white font-bold text-sm mb-1">
              {t('yourCards', { count: humanPlayer.cardCount })}
              {isHumanTurn && state.phase === 0 && (
                <span className="text-green-400 text-xs ml-2">{t('selectPrompt')}</span>
              )}
            </div>
            {/* Human cards */}
            <div className="flex flex-wrap gap-1">
              {humanPlayer.cards?.map((card, i) => (
                <HandCard
                  key={`${card.design}-${card.value}-${i}`}
                  card={card}
                  index={i}
                  selected={selectedCardIndices.includes(i)}
                  selectable={isHumanTurn && state.phase === 0 && !loading}
                  onToggle={toggleCard}
                />
              ))}
            </div>

            {/* Claimed value input (shown when cards are selected) */}
            {selectedCardIndices.length > 0 && isHumanTurn && state.phase === 0 && (
              <div className="mt-2 flex items-center gap-2">
                <span className="text-white text-sm">{t('claimedValue')}</span>
                <input
                  type="number"
                  min={1}
                  max={13}
                  value={claimedValue}
                  onChange={(e) => {
                    const num = Number(e.target.value);
                    setClaimedValue(Math.max(1, Math.min(13, num)));
                  }}
                  className="bg-black/50 text-white rounded px-2 py-1 w-16 text-sm border border-white/30"
                />
                <span className="text-[#ccc] text-xs">({valueName(claimedValue)})</span>
              </div>
            )}
          </div>
        )}

        <ErrorAlert message={error} />

        {/* Action buttons */}
        <div className="text-center">
          <button
            type="button"
            className={`${btnPrimary} min-w-[90px]`}
            disabled={loading}
            onClick={() => exec('reset', undefined, undefined, undefined, doubtConfig)}
          >
            {tc('button.reset')}
          </button>
          {isHumanTurn && state.phase === 0 && (
            <button
              type="button"
              className={`${btnSuccess} min-w-[90px]`}
              disabled={loading || selectedCardIndices.length === 0}
              onClick={handlePlay}
            >
              {t('playButton')}
            </button>
          )}
        </div>
      </GameFooter>
    </div>
  );
}
