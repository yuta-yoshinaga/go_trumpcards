import { useMemo } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { MemorySkeleton } from '../components/skeleton/MemorySkeleton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, useMemoryGame } from '../hooks/useMemoryGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { MemoryPhase } from '../types/phases';
import { playerName } from '../utils/playerUtils';

const MEMORY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MemoryPhase.FLIP1]: 'flip1',
  [MemoryPhase.FLIP2]: 'flip2',
  [MemoryPhase.RESULT]: 'result',
  [MemoryPhase.GAME_END]: 'gameEnd',
};

/** Renders the Memory card matching game page with board grid and scores. */
export function MemoryPage() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('memory');
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec, memoryConfig, handleConfigChange, handleFlip, handleNext } = useMemoryGame();

  const isResultForKbd = state?.phase === MemoryPhase.RESULT;

  const actionBindings = useMemo(
    () => [{ key: 'n', action: handleNext, enabled: isResultForKbd }],
    [handleNext, isResultForKbd],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!state && !loading,
  });

  const phaseNames = usePhaseNames('memory', MEMORY_PHASE_KEYS);

  if (!state) return <MemorySkeleton />;

  const isFlip1 = state.phase === MemoryPhase.FLIP1;
  const isFlip2 = state.phase === MemoryPhase.FLIP2;
  const isResult = state.phase === MemoryPhase.RESULT;
  const isGameEnd = state.phase === MemoryPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isFlip1 || isFlip2) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-blue" aria-busy={loading}>
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanTurn} />

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
                className={`relative aspect-[2/3] rounded border ${focusRingWhite} ${
                  bc.taken
                    ? 'bg-transparent border-transparent'
                    : bc.faceUp
                      ? 'bg-white border-yellow-400 ring-2 ring-yellow-400'
                      : 'bg-blue-800 border-blue-600 hover:border-yellow-400'
                } transition-all`}
              >
                {bc.faceUp && bc.card && <AnimatedCard card={bc.card} width={cardWidth} />}
                {!bc.taken && !bc.faceUp && <span className="text-game-text-muted text-xs">{idx}</span>}
              </button>
            ))}
          </div>
        </div>

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Action log */}
        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Footer */}
      <GameFooter className="bg-game-bg-blue-dark border-white/20 px-4 py-2.5">
        <ErrorAlert message={error} />
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
      <WinCelebration show={!!state?.gameEndFlag} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
