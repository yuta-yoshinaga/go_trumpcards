import { useCallback, useEffect, useMemo } from 'react';
import { clocksolitaireApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ClockSolitaireResponse } from '../types/card';
import { ClockSolitairePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

/** Clock Solitaire tutorial step definitions. */
const TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="clock-face"]',
    messageKey: 'tutorial.clockFace',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="clock-center"]',
    messageKey: 'tutorial.centerPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="clock-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="clock-reset"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Clock position angles: index 0 = 1 o'clock, index 11 = 12 o'clock */
const CLOCK_POSITIONS = Array.from({ length: 12 }, (_, i) => {
  const angle = ((i * 30 + 30 - 90) * Math.PI) / 180;
  return { x: Math.cos(angle), y: Math.sin(angle) };
});

/** Labels for clock positions (1-12). */
const CLOCK_LABELS = ['1', '2', '3', '4', '5', '6', '7', '8', '9', '10', '11', '12'];

/** Clock Solitaire page content component. */
function ClockSolitairePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('clocksolitaire');
  const { state, loading, error, exec: execApi, retry } = useGameApi(clocksolitaireApi.exec);
  const handleReset = useCallback(() => execApi('reset'), [execApi]);
  const handleStep = useCallback(() => execApi('step'), [execApi]);
  const handleAutoPlay = useCallback(() => execApi('autoplay'), [execApi]);

  useEffect(() => {
    execApi('reset');
  }, [execApi]);
  const { cardWidth, cardHeight } = useCardDimensions();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('clocksolitaire');
  type CsArgs = Parameters<typeof clocksolitaireApi.exec>;
  const cliConfig: CliGameConfig<ClockSolitaireResponse, CsArgs> = useMemo(
    () => ({
      gameName: 'clocksolitaire',
      parseCommand: (input: string): CliParseResult<CsArgs> => {
        const cmd = input.trim().toLowerCase();
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'step' || cmd === 's') return { args: ['step'] };
        if (cmd === 'autoplay' || cmd === 'auto' || cmd === 'a') return { args: ['autoplay'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: ClockSolitaireResponse) => {
        const lines: string[] = [];
        const phase =
          s.phase === ClockSolitairePhase.GAME_CLEAR
            ? 'CLEAR'
            : s.phase === ClockSolitairePhase.GAME_OVER
              ? 'OVER'
              : 'Playing';
        lines.push(`Phase: ${phase} | Steps: ${s.stepCount}`);
        const completed = s.faceUpCount.filter((c) => c >= 4).length;
        lines.push(`Completed: ${completed}/13`);
        if (s.currentCard) {
          lines.push(`Current: ${s.currentCard.design[0]}${s.currentCard.value}`);
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        's/step     - Place one card',
        'a/autoplay - Auto-play to end',
        'l/log      - Show action log',
        'r/reset    - Reset game',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });

  const theme = gameTheme.clocksolitaire;
  const isPlaying = state?.phase === ClockSolitairePhase.PLAYING;
  const isGameClear = state?.phase === ClockSolitairePhase.GAME_CLEAR;
  const isEnded = state?.phase === ClockSolitairePhase.GAME_CLEAR || state?.phase === ClockSolitairePhase.GAME_OVER;

  const phaseName = isGameClear
    ? t('phase.gameClear')
    : state?.phase === ClockSolitairePhase.GAME_OVER
      ? t('phase.gameOver')
      : t('phase.playing');

  if (error) return <ErrorAlert message={error} onRetry={retry} />;
  if (!state) return null;

  const radius = Math.min(cardWidth * 5, 180);

  return (
    <div className={`flex min-h-screen flex-1 flex-col ${theme.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.clocksolitaire')} />
      <PhaseIndicator phaseName={phaseName}>
        <span className="text-sm text-white/70">
          {t('stepCount')}: {state.stepCount}
        </span>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/clocksolitaire" />
      </PhaseIndicator>

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <LandscapeBanner message={t('landscapeBanner')} />

          <div className="flex-1 overflow-y-auto px-4 pt-3 lg:px-8">
            {/* Clock face layout */}
            <div
              data-tutorial="clock-face"
              className="relative mx-auto"
              style={{ width: radius * 2 + cardWidth + 16, height: radius * 2 + cardHeight + 16 }}
            >
              {/* 12 clock positions */}
              {CLOCK_POSITIONS.map((pos, i) => {
                const pile = state.piles[i];
                const faceUpCount = state.faceUpCount[i];
                const isComplete = faceUpCount >= 4;
                const topCard = pile?.[pile.length - 1]?.card ?? null;
                const cx = radius + cardWidth / 2 + 8 + pos.x * radius;
                const cy = radius + cardHeight / 2 + 8 + pos.y * radius;

                return (
                  <div
                    key={i}
                    className="absolute flex flex-col items-center"
                    style={{
                      left: cx - cardWidth / 2,
                      top: cy - cardHeight / 2,
                      width: cardWidth,
                    }}
                  >
                    <span className="mb-0.5 text-xs font-bold text-white/70">{CLOCK_LABELS[i]}</span>
                    {pile && pile.length > 0 ? (
                      <div className="relative" style={{ width: cardWidth }}>
                        {isComplete && topCard ? (
                          <AnimatedCard
                            card={topCard}
                            width={cardWidth}
                            className="rounded border-2 border-green-400"
                          />
                        ) : (
                          <AnimatedCardBack width={cardWidth} />
                        )}
                        <span className="absolute -bottom-4 left-1/2 -translate-x-1/2 text-[10px] text-white/60">
                          {faceUpCount}/4
                        </span>
                      </div>
                    ) : (
                      <div
                        className="rounded border border-dashed border-white/30"
                        style={{ width: cardWidth, height: cardHeight }}
                      />
                    )}
                  </div>
                );
              })}

              {/* Center pile (K) */}
              <div
                data-tutorial="clock-center"
                className="absolute flex flex-col items-center"
                style={{
                  left: radius + 8,
                  top: radius + 8,
                  width: cardWidth,
                }}
              >
                <span className="mb-0.5 text-xs font-bold text-yellow-400">K</span>
                {(() => {
                  const centerPile = state.piles[12];
                  const centerTopCard = centerPile?.[centerPile.length - 1]?.card ?? null;
                  return centerPile && centerPile.length > 0 ? (
                    <div className="relative" style={{ width: cardWidth }}>
                      {state.faceUpCount[12] >= 4 && centerTopCard ? (
                        <AnimatedCard
                          card={centerTopCard}
                          width={cardWidth}
                          className="rounded border-2 border-yellow-400"
                        />
                      ) : (
                        <AnimatedCardBack width={cardWidth} />
                      )}
                      <span className="absolute -bottom-4 left-1/2 -translate-x-1/2 text-[10px] text-white/60">
                        {state.faceUpCount[12]}/4
                      </span>
                    </div>
                  ) : (
                    <div
                      className="rounded border border-dashed border-white/30"
                      style={{ width: cardWidth, height: cardHeight }}
                    />
                  );
                })()}
              </div>
            </div>

            {/* Current card */}
            {state.currentCard && (
              <div className="mt-6 flex flex-col items-center gap-1">
                <span className="text-sm text-white/80">{t('currentCard')}</span>
                <AnimatedCard card={state.currentCard} width={cardWidth} className="rounded border-2 border-white/50" />
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${theme.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap items-center gap-2">
              {isPlaying && (
                <div data-tutorial="clock-controls" className="flex gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleStep}
                    disabled={loading}
                  >
                    {t('step')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleAutoPlay}
                    disabled={loading}
                  >
                    {t('autoplay')}
                  </button>
                </div>
              )}
              <div data-tutorial="clock-reset">
                <button
                  type="button"
                  className={`${btnDanger} ${focusRingWhite}`}
                  onClick={() => requestConfirm(handleReset)}
                >
                  {tc('reset')}
                </button>
              </div>
            </div>
          </GameFooter>
        </>
      )}

      <WinCelebration show={state.phase === ClockSolitairePhase.GAME_CLEAR} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}

/** Clock Solitaire game page wrapped with tutorial support. */
export function ClockSolitairePage() {
  return (
    <TutorialWrapper gameName="clocksolitaire" steps={TUTORIAL_STEPS}>
      <ClockSolitairePageContent />
    </TutorialWrapper>
  );
}
