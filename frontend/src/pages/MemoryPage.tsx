import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { MemorySkeleton } from '../components/skeleton/MemorySkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, useMemoryGame } from '../hooks/useMemoryGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnOutline, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import { MemoryPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

/** Memory tutorial step definitions. */
const MEM_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="mem-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mem-board"]',
    messageKey: 'tutorial.board',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mem-next-button"]',
    messageKey: 'tutorial.nextButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="mem-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Memory tutorial configuration. */
const MEM_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'memory',
  steps: MEM_TUTORIAL_STEPS,
};

const MEMORY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MemoryPhase.FLIP1]: 'flip1',
  [MemoryPhase.FLIP2]: 'flip2',
  [MemoryPhase.RESULT]: 'result',
  [MemoryPhase.GAME_END]: 'gameEnd',
};

/** Renders the Memory card matching game page with board grid and scores. */
export function MemoryPage() {
  const { t: tMem } = useTranslation('memory');
  return (
    <TutorialProvider config={MEM_TUTORIAL_CONFIG} translateMessage={tMem}>
      <MemoryPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Memory page, wrapped by TutorialProvider. */
function MemoryPageContent() {
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
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.memory.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.memory')} />
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanTurn}>
        <TutorialButton />
      </PhaseIndicator>

      <LandscapeBanner message={t('landscapeBanner')} />

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

      {/* Content area – scroll on mobile, fit-to-viewport on desktop */}
      <div className="flex-1 overflow-y-auto lg:overflow-hidden lg:flex lg:flex-col pt-3 lg:pt-1 px-4 lg:px-8">
        {/* Player scores – compact inline layout to maximise board visibility */}
        <div
          className="my-1 px-2 py-1 rounded bg-black/30 text-white text-sm flex flex-wrap items-center gap-y-0.5 lg:shrink-0"
          data-tutorial="mem-score-table"
          role="status"
          aria-label={t('scores')}
        >
          {state.players.map((p, idx) => (
            <span key={p.id} className={p.isHuman ? 'text-yellow-300' : ''}>
              {idx > 0 && (
                <span className="text-white/40 mr-3" aria-hidden="true">
                  |
                </span>
              )}
              {t('scoreLine', { name: playerName(p.id, p.isHuman), count: p.pairCount })}
            </span>
          ))}
        </div>

        {/* Board: responsive grid (4/8/13 columns); on lg fills remaining height */}
        <div
          className="my-3 lg:my-1 p-2 lg:p-1 rounded bg-black/40 lg:flex-1 lg:min-h-0 lg:overflow-hidden"
          data-tutorial="mem-board"
        >
          <div className="grid grid-cols-6 md:grid-cols-8 lg:grid-cols-13 gap-0.5 md:gap-1 lg:grid-rows-4 lg:h-full">
            {state.board.map((bc, idx) => (
              <button
                type="button"
                key={`board-${idx.toString()}`}
                data-testid={`board-${idx.toString()}`}
                aria-label={bc.faceUp && bc.card ? cardAlt(bc.card) : t('cardFaceDown', { position: idx + 1 })}
                disabled={loading || !isHumanTurn || bc.taken || bc.faceUp}
                onClick={() => handleFlip(idx)}
                className={`memory-card relative aspect-[2/3] lg:aspect-auto rounded ${focusRingWhite} ${
                  bc.taken
                    ? 'hidden'
                    : bc.faceUp
                      ? 'bg-white ring-2 ring-yellow-400 shadow-lg shadow-yellow-400/30'
                      : 'bg-blue-800 border border-white/10 hover:ring-1 hover:ring-yellow-400'
                } transition-all`}
              >
                <div className={`memory-card-inner${bc.faceUp ? ' flipped' : ''}`}>
                  <div className="memory-card-back">
                    <img src="/images/z01.png" alt="" className="w-full h-full object-contain rounded" />
                  </div>
                  <div className="memory-card-front">
                    {bc.card && <AnimatedCard card={bc.card} width={cardWidth} />}
                  </div>
                </div>
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
      <GameFooter className={`${gameTheme.memory.footer} px-4 py-2.5`}>
        <ErrorAlert message={error} />
        <div className="flex gap-2 items-center">
          {isResult && (
            <div data-tutorial="mem-next-button">
              <button type="button" className={btnSuccess} onClick={handleNext} disabled={loading}>
                {t('nextButton')}
              </button>
            </div>
          )}
          <div data-tutorial="mem-reset-button">
            <button
              type="button"
              className={btnOutline}
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
        </div>
      </GameFooter>
      <WinCelebration show={!!state?.gameEndFlag} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
