import { useTranslation } from 'react-i18next';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { useActionLog } from '../hooks/useActionLog';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useHeartsGame } from '../hooks/useHeartsGame';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { HEARTS_PHASE } from '../types/card';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

const passDirectionKeys = ['left', 'right', 'across', 'none'] as const;

export function HeartsPage() {
  const { t } = useTranslation('hearts');
  const { t: tc } = useTranslation('common');
  const {
    state,
    loading,
    error,
    exec,
    heartsConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handlePass,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useHeartsGame();
  const { actionLog, showActionLog, hideActionLog } = useActionLog('hearts');

  if (!state) return null;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPassPhase = state.phase === HEARTS_PHASE.PASS;
  const isPlayPhase = state.phase === HEARTS_PHASE.PLAY;
  const isTrickEnd = state.phase === HEARTS_PHASE.TRICK_END;
  const isRoundEnd = state.phase === HEARTS_PHASE.ROUND_END;
  const isGameEnd = state.phase === HEARTS_PHASE.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a2c5c]" aria-busy={loading}>
      {loading && <span className="sr-only">{tc('status.loading')}</span>}

      {/* Settings */}
      <details className="px-4 pt-2">
        <summary className="text-white/70 text-sm cursor-pointer">{t('settings.title')}</summary>
        <div className="mt-2 flex flex-wrap gap-4 text-sm text-white/70">
          <label>
            {t('settings.cpuDifficulty')}
            <select
              value={heartsConfig.cpuDifficulty}
              onChange={(e) => handleConfigChange('cpuDifficulty', e.target.value)}
              className="ml-1 bg-gray-700 text-white rounded px-1"
            >
              {CPU_DIFFICULTY_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {t(`settings.${o.label.toLowerCase()}`)}
                </option>
              ))}
            </select>
          </label>
          <label>
            {t('settings.pointLimit')}
            <select
              value={heartsConfig.pointLimit}
              onChange={(e) => handleConfigChange('pointLimit', e.target.value)}
              className="ml-1 bg-gray-700 text-white rounded px-1"
            >
              {POINT_LIMIT_OPTIONS.map((v) => (
                <option key={v} value={v}>
                  {v}
                </option>
              ))}
            </select>
          </label>
        </div>
      </details>

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Round/Trick info */}
        <div className="text-white text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span>{state.heartsBroken ? t('heartsBroken') : t('heartsNotBroken')}</span>
        </div>

        {/* Pass direction (pass phase) */}
        {isPassPhase && (
          <div className="text-yellow-300 text-center mb-2">
            {t(`passDirection.${passDirectionKeys[state.passDirection]}`)}
          </div>
        )}

        {/* CPU players */}
        {state.players
          .filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-white/70 text-sm">
                {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}
              </div>
            </div>
          ))}

        {/* Current trick */}
        {state.currentTrick.length > 0 && (
          <div className="my-3 p-3 rounded bg-black/40">
            <div className="text-white/70 text-sm mb-1">{t('currentTrick')}</div>
            <div className="flex gap-2">
              {state.currentTrick.map((trickCard) => (
                <div key={`trick-${trickCard.playerIdx}`} className="text-center">
                  <CardImage card={trickCard.card} />
                  <div className="text-white/50 text-xs mt-1">
                    {playerName(
                      state.players[trickCard.playerIdx]?.id ?? trickCard.playerIdx,
                      state.players[trickCard.playerIdx]?.isHuman ?? false,
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Score table */}
        <div className="my-3 p-2 rounded bg-black/30">
          <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
          <table className="w-full text-sm text-white/70">
            <thead>
              <tr>
                <th className="text-left">{t('scoresPlayer')}</th>
                <th>{t('scoresRound')}</th>
                <th>{t('scoresTotal')}</th>
                <th>{t('scoresTricks')}</th>
              </tr>
            </thead>
            <tbody>
              {state.players.map((p) => (
                <tr key={p.id} className={p.isHuman ? 'text-yellow-300' : ''}>
                  <td>{playerName(p.id, p.isHuman)}</td>
                  <td className="text-center">{p.roundScore}</td>
                  <td className="text-center">{p.cumulativeScore}</td>
                  <td className="text-center">{p.trickCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Error */}
        <ErrorAlert message={error} />

        {/* Action log */}
        {isGameEnd && actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
        {isGameEnd && !actionLog && (
          <div className="text-center my-2">
            <button type="button" className={btnSecondary} onClick={showActionLog}>
              {tc('actionLog.view')}
            </button>
          </div>
        )}
      </div>

      {/* Footer */}
      <GameFooter className="bg-[#101c3a] border-white/20 px-4 py-2.5">
        {/* Human cards */}
        {humanPlayer && (
          <div className="flex flex-wrap gap-1 mb-2">
            {humanPlayer.cards.map((card, idx) => (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => toggleCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={selectedCardIndices.includes(idx)}
                className={`${selectedCardIndices.includes(idx) ? 'ring-2 ring-yellow-400 -translate-y-1' : ''} transition-transform`}
              >
                <CardImage card={card} />
              </button>
            ))}
          </div>
        )}

        <div className="flex gap-2 items-center">
          {isPassPhase && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handlePass}
              disabled={loading || selectedCardIndices.length !== 3}
            >
              {t('passButton')}
            </button>
          )}
          {isHumanTurn && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handlePlay}
              disabled={loading || selectedCardIndices.length !== 1}
            >
              {t('playButton')}
            </button>
          )}
          {isTrickEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
              {t('nextTrick')}
            </button>
          )}
          {isRoundEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
              {t('nextRound')}
            </button>
          )}
          <button
            type="button"
            className={btnWarning}
            onClick={() =>
              exec('reset', undefined, undefined, {
                cpuDifficulty: heartsConfig.cpuDifficulty,
                pointLimit: heartsConfig.pointLimit,
              })
            }
            disabled={loading}
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
    </div>
  );
}
