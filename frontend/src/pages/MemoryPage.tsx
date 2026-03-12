import { useTranslation } from 'react-i18next';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { useActionLog } from '../hooks/useActionLog';
import { CPU_DIFFICULTY_OPTIONS, useMemoryGame } from '../hooks/useMemoryGame';
import { btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { MEMORY_PHASE } from '../types/card';
import { playerName } from '../utils/playerUtils';

export function MemoryPage() {
  const { t } = useTranslation('memory');
  const { t: tc } = useTranslation('common');
  const { state, loading, error, exec, memoryConfig, handleConfigChange, handleFlip, handleNext } = useMemoryGame();
  const { actionLog, showActionLog, hideActionLog } = useActionLog('memory');

  if (!state) return null;

  const isFlip1 = state.phase === MEMORY_PHASE.FLIP1;
  const isFlip2 = state.phase === MEMORY_PHASE.FLIP2;
  const isResult = state.phase === MEMORY_PHASE.RESULT;
  const isGameEnd = state.phase === MEMORY_PHASE.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isFlip1 || isFlip2) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a2c5c]" aria-busy={loading}>
      <LoadingSpinner loading={loading} />

      {/* Settings */}
      <details className="px-4 pt-2">
        <summary className="text-white text-sm cursor-pointer">{t('settings.title')}</summary>
        <div className="mt-2 flex flex-wrap gap-4 text-sm text-white">
          <label htmlFor="cpuDifficulty">
            {t('settings.cpuDifficulty')}
            <select
              id="cpuDifficulty"
              value={memoryConfig.cpuDifficulty}
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
        </div>
      </details>

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Player scores */}
        <div className="my-2 p-2 rounded bg-black/30">
          <div className="text-white text-sm mb-1">{t('scores')}</div>
          <table className="w-full text-sm text-white">
            <thead>
              <tr>
                <th className="text-left">{t('scoresPlayer')}</th>
                <th>{t('scoresPairs')}</th>
              </tr>
            </thead>
            <tbody>
              {state.players.map((p) => (
                <tr key={p.id} className={p.isHuman ? 'text-yellow-300' : ''}>
                  <td>{playerName(p.id, p.isHuman)}</td>
                  <td className="text-center">{p.pairCount}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Board: 4×13 grid */}
        <div className="my-3 p-2 rounded bg-black/40">
          <div className="grid grid-cols-4 sm:grid-cols-7 md:grid-cols-13 gap-1">
            {state.board.map((bc, idx) => (
              <button
                type="button"
                key={`board-${idx.toString()}`}
                disabled={loading || !isHumanTurn || bc.taken || bc.faceUp}
                onClick={() => handleFlip(idx)}
                aria-hidden={bc.taken || undefined}
                className={`relative aspect-[2/3] rounded border ${
                  bc.taken
                    ? 'bg-transparent border-transparent'
                    : bc.faceUp
                      ? 'bg-white border-yellow-400 ring-2 ring-yellow-400'
                      : 'bg-blue-800 border-blue-600 hover:border-yellow-400'
                } transition-all`}
              >
                {bc.faceUp && bc.card && <CardImage card={bc.card} />}
                {!bc.taken && !bc.faceUp && <span className="text-white/40 text-xs">{idx}</span>}
              </button>
            ))}
          </div>
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
        <div className="flex gap-2 items-center">
          {isResult && (
            <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
              {t('nextButton')}
            </button>
          )}
          <button
            type="button"
            className={btnWarning}
            onClick={() =>
              exec('reset', undefined, {
                cpuDifficulty: memoryConfig.cpuDifficulty,
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
