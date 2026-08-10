import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { clocksolitaireApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnOutline, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ClockSolitaireResponse } from '../types/card';
import { ClockSolitairePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

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

/** Autoplay animation speed presets. */
type AutoPlaySpeed = 'slow' | 'normal' | 'fast';

/**
 * Delay (ms) between auto-advanced `step` calls per speed preset. A larger delay
 * gives each card's `isFlightTarget` landing highlight more time to play out; a
 * smaller delay races to the finish.
 */
const AUTOPLAY_DELAY_MS: Record<AutoPlaySpeed, number> = {
  slow: 900,
  normal: 450,
  fast: 150,
};

const AUTOPLAY_SPEED_STORAGE_KEY = 'clocksolitaire:autoPlaySpeed';

/** Read the persisted autoplay speed, falling back to `normal` when unset/invalid. */
function loadAutoPlaySpeed(): AutoPlaySpeed {
  try {
    const v = localStorage.getItem(AUTOPLAY_SPEED_STORAGE_KEY);
    if (v === 'slow' || v === 'normal' || v === 'fast') return v;
  } catch {
    // localStorage may be unavailable (private mode / SSR); fall through to default.
  }
  return 'normal';
}

/** Clock Solitaire page content component. */
function ClockSolitairePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('clocksolitaire');
  const { state, loading, error, exec: execApi, retry } = useGameApi(clocksolitaireApi.exec);

  const [autoPlaySpeed, setAutoPlaySpeed] = useState<AutoPlaySpeed>(loadAutoPlaySpeed);
  const [autoPlaying, setAutoPlaying] = useState(false);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  // Undo the last placed card. Stop autoplay first so the timed loop doesn't
  // race the rewind.
  const handleUndo = useCallback(() => {
    setAutoPlaying(false);
    return execApi('undo');
  }, [execApi]);
  // Autoplay is driven client-side as a timed sequence of `step` calls (see the
  // effect below) so each card's landing highlight plays out; the button toggles it.
  const handleAutoPlay = useCallback(() => setAutoPlaying((prev) => !prev), []);
  const handleSelectSpeed = useCallback((v: string) => {
    const speed: AutoPlaySpeed = v === 'slow' || v === 'fast' ? v : 'normal';
    setAutoPlaySpeed(speed);
    try {
      localStorage.setItem(AUTOPLAY_SPEED_STORAGE_KEY, speed);
    } catch {
      // Persistence is best-effort; ignore storage failures.
    }
  }, []);
  const handleReset = useCallback(() => {
    setAutoPlaying(false);
    return execApi('reset');
  }, [execApi]);

  // Latest state/speed read from a self-scheduling autoplay loop without
  // re-subscribing the effect on every render.
  const stateRef = useRef(state);
  stateRef.current = state;
  const speedRef = useRef(autoPlaySpeed);
  speedRef.current = autoPlaySpeed;

  // Drive autoplay client-side: recursively `step` on a speed-scaled timer so each
  // card's `isFlightTarget` landing highlight plays out, stopping as soon as the
  // game reaches GAME_CLEAR / GAME_OVER.
  useEffect(() => {
    if (!autoPlaying) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    const tick = async () => {
      const s = stateRef.current;
      if (!s || s.phase !== ClockSolitairePhase.PLAYING) {
        setAutoPlaying(false);
        return;
      }
      await execApi('step');
      if (cancelled) return;
      timer = setTimeout(() => void tick(), AUTOPLAY_DELAY_MS[speedRef.current]);
    };
    timer = setTimeout(() => void tick(), AUTOPLAY_DELAY_MS[speedRef.current]);
    return () => {
      cancelled = true;
      clearTimeout(timer);
    };
  }, [autoPlaying, execApi]);

  useMountReset(execApi);
  const { cardWidth, cardHeight } = useCardDimensions();
  const { hint, hintEnabled, setHintEnabled } = useGameHint('clocksolitaire', state);

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
        if (cmd === 'undo' || cmd === 'u') return { args: ['undo'] };
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
        'u/undo     - Undo the last step',
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
  if (!state) return <GameSkeleton gameKey="clocksolitaire" layout={{ kind: 'centered', rows: [4, 5, 4] }} />;

  const radius = Math.min(cardWidth * 5, 180);

  // Resolve the same text GameMessageBox shows so it can also be announced in a
  // dedicated live region (the visible box is not guaranteed to be a live region).
  const liveMessage = (() => {
    const parts: string[] = [];
    if (state.messageCode) {
      const key = `messageCode.${state.messageCode}`;
      const translated = tc(key, state.messageParams ?? {});
      parts.push(translated === key ? (state.message ?? '') : translated);
    } else {
      parts.push(state.message ?? '');
    }
    // The pulsing ring on the destination pile is visual only. The CUI presenter
    // already spells the same thing out every turn, so mirror it here (#4785).
    const value = state.currentCard?.value ?? 0;
    if (value >= 1 && value <= 12) parts.push(t('placementHint', { hour: value }));
    else if (value === 13) parts.push(t('placementKing'));
    return parts.filter(Boolean).join(' ');
  })();

  return (
    <GamePageShell
      title={tc('nav.clocksolitaire')}
      gameThemeBg={theme.bg}
      phaseName={phaseName}
      gamePath="/clocksolitaire"
      gameEndFlag={isEnded}
      winShow={isGameClear}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('stepCount')}: {state.stepCount}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
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
              {/* Map the waiting card to its destination pile index once per render:
                  values 1..12 land on the matching hour (index 0..11) and a King (13) lands
                  on the center pile (index 12). When no card is waiting we use -1 so no pile
                  matches. */}
              {(() => {
                const currentValue = state.currentCard?.value ?? 0;
                const targetIdx =
                  currentValue >= 1 && currentValue <= 12 ? currentValue - 1 : currentValue === 13 ? 12 : -1;
                return CLOCK_POSITIONS.map((pos, i) => {
                  const pile = state.piles[i];
                  const faceUpCount = state.faceUpCount[i];
                  const isComplete = faceUpCount >= 4;
                  const topCard = pile?.[pile.length - 1]?.card ?? null;
                  const cx = radius + cardWidth / 2 + 8 + pos.x * radius;
                  const cy = radius + cardHeight / 2 + 8 + pos.y * radius;
                  const isFlightTarget = targetIdx === i;

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
                      <span className="mb-0.5 text-xs font-bold text-ds-text-muted">{CLOCK_LABELS[i]}</span>
                      {pile && pile.length > 0 ? (
                        <div
                          className={`relative rounded ${isFlightTarget ? 'ring-2 ring-ds-warning animate-pulse' : ''}`}
                          data-flight-target={isFlightTarget ? 'true' : undefined}
                          style={{ width: cardWidth }}
                        >
                          {isComplete && topCard ? (
                            <AnimatedCard
                              card={topCard}
                              width={cardWidth}
                              className="rounded border-2 border-ds-success"
                            />
                          ) : (
                            <AnimatedCardBack width={cardWidth} />
                          )}
                          <span className="absolute -bottom-4 left-1/2 -translate-x-1/2 text-[10px] text-ds-text-muted">
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
                });
              })()}

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
                <span className="mb-0.5 text-xs font-bold text-ds-warning">K</span>
                {(() => {
                  const centerPile = state.piles[12];
                  const centerTopCard = centerPile?.[centerPile.length - 1]?.card ?? null;
                  const isCenterFlightTarget = state.currentCard?.value === 13;
                  return centerPile && centerPile.length > 0 ? (
                    <div
                      className={`relative rounded ${isCenterFlightTarget ? 'ring-2 ring-ds-warning animate-pulse' : ''}`}
                      data-flight-target={isCenterFlightTarget ? 'true' : undefined}
                      style={{ width: cardWidth }}
                    >
                      {state.faceUpCount[12] >= 4 && centerTopCard ? (
                        <AnimatedCard
                          card={centerTopCard}
                          width={cardWidth}
                          className="rounded border-2 border-ds-warning"
                        />
                      ) : (
                        <AnimatedCardBack width={cardWidth} />
                      )}
                      <span className="absolute -bottom-4 left-1/2 -translate-x-1/2 text-[10px] text-ds-text-muted">
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
                <span className="text-sm text-ds-text-primary">{t('currentCard')}</span>
                <AnimatedCard card={state.currentCard} width={cardWidth} className="rounded border-2 border-white/50" />
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {/* Announce each step's result (and game clear/over) to screen readers. */}
            <div className="sr-only" role="status" aria-live="polite" data-testid="cs-live-region">
              {liveMessage}
            </div>

            <ActionLogSection
              isEndPhase={isEnded}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}

          {/* Settings */}
          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'autoPlaySpeed',
                    testId: 'autoplay-speed-select',
                    label: t('settings.speed'),
                    tooltip: t('settings.speedHelp'),
                    value: autoPlaySpeed,
                    options: [
                      { value: 'slow', label: t('settings.speedSlow') },
                      { value: 'normal', label: t('settings.speedNormal') },
                      { value: 'fast', label: t('settings.speedFast') },
                    ],
                    onSelect: handleSelectSpeed,
                  },
                  hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                ],
              },
            ]}
          />

          <GameFooter className={`${theme.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap items-center gap-2">
              {isPlaying && (
                <div data-tutorial="clock-controls" className="flex gap-2">
                  <button
                    type="button"
                    className={`${btnPrimary} ${focusRingWhite}`}
                    onClick={handleStep}
                    disabled={loading || autoPlaying}
                    data-testid="cs-step-button"
                  >
                    {t('step')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSuccess} ${focusRingWhite}`}
                    onClick={handleAutoPlay}
                    aria-pressed={autoPlaying}
                    data-testid="cs-autoplay-button"
                  >
                    {autoPlaying ? t('stopAutoplay') : t('autoplay')}
                  </button>
                </div>
              )}
              <button
                type="button"
                className={`${btnOutline} ${focusRingWhite}`}
                onClick={handleUndo}
                disabled={!state.canUndo || loading || autoPlaying}
                data-testid="cs-undo-button"
              >
                {t('undo')}
              </button>
              <GameResetButton
                isGameEnd={isEnded}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="clock-reset"
                className={focusRingWhite}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Clock Solitaire game page wrapped with tutorial support. */
export const ClockSolitairePage = withTutorial(ClockSolitairePageContent, 'clocksolitaire', TUTORIAL_STEPS);
