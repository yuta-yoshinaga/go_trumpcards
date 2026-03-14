import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ConfirmDialog } from '../components/ConfirmDialog';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { LoadingSpinner } from '../components/LoadingSpinner';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { useActionLog } from '../hooks/useActionLog';
import { useConfirmDialog } from '../hooks/useConfirmDialog';
import { CPU_DIFFICULTY_OPTIONS, useMemoryGame } from '../hooks/useMemoryGame';
import { btnSuccess, btnWarning } from '../styles/buttonStyles';
import { MEMORY_PHASE } from '../types/card';
import { playerName } from '../utils/playerUtils';

export function MemoryPage() {
  const { t } = useTranslation('memory');
  const { t: tc } = useTranslation('common');
  const { state, loading, error, exec, memoryConfig, handleConfigChange, handleFlip, handleNext } = useMemoryGame();
  const { actionLog, showActionLog, hideActionLog } = useActionLog('memory');
  const { isOpen: confirmOpen, requestConfirm, confirm: confirmReset, cancel: cancelReset } = useConfirmDialog();

  if (!state) return null;

  const isFlip1 = state.phase === MEMORY_PHASE.FLIP1;
  const isFlip2 = state.phase === MEMORY_PHASE.FLIP2;
  const isResult = state.phase === MEMORY_PHASE.RESULT;
  const isGameEnd = state.phase === MEMORY_PHASE.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isFlip1 || isFlip2) && state.players[state.currentPlayerIdx]?.isHuman === true;

  const phaseNameMap: Record<number, string> = {
    [MEMORY_PHASE.FLIP1]: t('phase.flip1'),
    [MEMORY_PHASE.FLIP2]: t('phase.flip2'),
    [MEMORY_PHASE.RESULT]: t('phase.result'),
    [MEMORY_PHASE.GAME_END]: t('phase.gameEnd'),
  };

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-[#1a2c5c]" aria-busy={loading}>
      <LoadingSpinner loading={loading} />

      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNameMap[state.phase] ?? t('phase.flip1')} isHumanTurn={isHumanTurn} />

      {/* Landscape orientation banner (visible on small portrait screens) */}
      <div className="hidden portrait:flex sm:hidden items-center gap-2 px-4 py-2 bg-yellow-500/90 text-black text-sm font-medium">
        <span aria-hidden="true">&#8635;</span>
        <span>{t('landscapeBanner')}</span>
      </div>

      {/* Settings */}
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: memoryConfig.cpuDifficulty,
                options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(`settings.${o.label.toLowerCase()}`),
                })),
                onSelect: (v) => handleConfigChange('cpuDifficulty', v),
              },
            ],
          },
        ]}
      />

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4">
        {/* Player scores */}
        <div className="my-2 p-2 rounded bg-black/30 text-white text-sm">
          <div className="mb-1">{t('scores')}</div>
          <table className="w-full">
            <thead>
              <tr>
                <th scope="col" className="text-left">
                  {t('scoresPlayer')}
                </th>
                <th scope="col">{t('scoresPairs')}</th>
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

        {/* Board: responsive grid (4/6/8/13 columns by breakpoint) */}
        <div className="my-3 p-2 rounded bg-black/40">
          <div className="grid grid-cols-4 sm:grid-cols-6 md:grid-cols-8 lg:grid-cols-13 gap-1">
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
        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
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
              requestConfirm(() => {
                hideActionLog();
                return exec('reset', undefined, { cpuDifficulty: memoryConfig.cpuDifficulty });
              })
            }
            disabled={loading}
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
      <ConfirmDialog
        open={confirmOpen}
        title={tc('button.confirmReset')}
        message={tc('button.confirmResetMessage')}
        confirmLabel={tc('button.confirm')}
        cancelLabel={tc('button.cancel')}
        onConfirm={confirmReset}
        onCancel={cancelReset}
      />
    </div>
  );
}
