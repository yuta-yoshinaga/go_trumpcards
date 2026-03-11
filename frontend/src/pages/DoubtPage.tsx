import { useTranslation } from 'react-i18next';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { DoubtCpuArea } from '../components/doubt/DoubtCpuArea';
import { DoubtHandCard } from '../components/doubt/DoubtHandCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { useActionLog } from '../hooks/useActionLog';
import {
  actionDesc,
  CPU_MEMORY_OPTIONS,
  DOUBT_WINDOW_OPTIONS,
  PENALTY_DRAW_LIMIT_OPTIONS,
  useDoubtGame,
} from '../hooks/useDoubtGame';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import type { DoubtCpuAction } from '../types/card';
import { valueName } from '../utils/cardUtils';
import { playerName } from '../utils/playerUtils';

export function DoubtPage() {
  const { t } = useTranslation('doubt');
  const { t: tc } = useTranslation('common');
  const {
    state,
    loading,
    error,
    exec,
    countdown,
    doubtConfig,
    selectedCardIndices,
    toggleCard,
    claimedValue,
    setClaimedValue,
    handleConfigChange,
    handleConfigToggle,
    handlePlay,
    handleDoubt,
    handleSkip,
    handleCpuDoubtConfirm,
  } = useDoubtGame();

  const { actionLog, showActionLog, hideActionLog } = useActionLog('doubt');

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

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a2c5c]" aria-busy={loading} aria-live="polite">
      <LoadingSpinner loading={loading} />
      {/* Settings panel */}
      <details className="px-4 pt-2">
        <summary className="text-white/70 text-xs cursor-pointer select-none">{t('settings.title')}</summary>
        <div className="bg-black/30 rounded-lg p-3 mt-1 flex flex-wrap gap-4 text-sm text-white">
          <label htmlFor="doubtWindowSec" className="flex items-center gap-2">
            {t('settings.doubtTime')}
            <select
              id="doubtWindowSec"
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
          <label htmlFor="cpuMemoryLevel" className="flex items-center gap-2">
            {t('settings.cpuMemory')}
            <select
              id="cpuMemoryLevel"
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
          <label htmlFor="penaltyDrawLimit" className="flex items-center gap-2">
            {t('settings.penaltyDrawLimit')}
            <select
              id="penaltyDrawLimit"
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
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={doubtConfig.cpuHesitationEnabled}
              onChange={(e) => handleConfigToggle('cpuHesitationEnabled', e.target.checked)}
            />
            {t('settings.cpuHesitation')}
          </label>
          <label className="flex items-center gap-2">
            <input
              type="checkbox"
              checked={doubtConfig.cpuMetaAI}
              onChange={(e) => handleConfigToggle('cpuMetaAI', e.target.checked)}
            />
            {t('settings.cpuMetaAI')}
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

        {/* Action log */}
        {state.gameEndFlag && !actionLog && (
          <div className="text-center my-2">
            <button type="button" className={btnSecondary} onClick={showActionLog}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
        {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
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
                <DoubtHandCard
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
            onClick={() => {
              hideActionLog();
              exec('reset', undefined, undefined, undefined, doubtConfig);
            }}
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
