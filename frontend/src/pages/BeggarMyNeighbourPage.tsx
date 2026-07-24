import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { beggarmyneighbourApi } from '../api/gameApi';
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
import { gameTheme } from '../styles/gameTheme';
import type { BeggarMyNeighbourResponse } from '../types/card';
import { BeggarMyNeighbourPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BEGGARMYNEIGHBOUR_HELP, parseBeggarMyNeighbourCommand } from '../utils/cli/commands/beggarmyneighbourCommands';
import { formatBeggarMyNeighbourState } from '../utils/cli/formatters/beggarmyneighbourFormatter';
import type { CliGameConfig } from '../utils/cli/types';

type BeggarMyNeighbourArgs = Parameters<typeof beggarmyneighbourApi.exec>;

const DEFAULT_MAX_ROUNDS = 2000;

/** Autoplay playback speed presets. */
type AutoPlaySpeed = 'slow' | 'normal' | 'fast';

/**
 * Delay (ms) between auto-advanced `step` calls per speed preset. A larger delay
 * gives each play's central-pile stack / penalty reveal more time to be seen; a
 * smaller delay races to the finish.
 */
const AUTOPLAY_DELAY_MS: Record<AutoPlaySpeed, number> = {
  slow: 900,
  normal: 450,
  fast: 150,
};

const AUTOPLAY_SPEED_STORAGE_KEY = 'beggarmyneighbour:autoPlaySpeed';

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

/** Tutorial steps for the Beggar-My-Neighbour game. */
const BMN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bmn-cpu-pile"]',
    messageKey: 'tutorial.cpuPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-central-pile"]',
    messageKey: 'tutorial.centralPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-player-pile"]',
    messageKey: 'tutorial.playerPile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-step-button"]',
    messageKey: 'tutorial.stepButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-autoplay-button"]',
    messageKey: 'tutorial.autoplayButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bmn-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Returns emphasis ring classes for the central pile based on the current phase. */
function centralPileEmphasis(phase: number): string {
  if (phase === BeggarMyNeighbourPhase.PAY_PENALTY) return 'ring-2 ring-ds-warning';
  if (phase === BeggarMyNeighbourPhase.COLLECT) return 'ring-2 ring-ds-success';
  return '';
}

/** Renders the Beggar-My-Neighbour game page. */
export const BeggarMyNeighbourPage = withTutorial(
  BeggarMyNeighbourPageContent,
  'beggarmyneighbour',
  BMN_TUTORIAL_STEPS,
);

/** Inner content of the Beggar-My-Neighbour page. */
function BeggarMyNeighbourPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('beggarmyneighbour');
  const { state, loading, error, exec: execApi, retry } = useGameApi(beggarmyneighbourApi.exec);
  const { cardWidth } = useCardDimensions();
  const [maxRounds, setMaxRounds] = useState(DEFAULT_MAX_ROUNDS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('beggarmyneighbour', state);

  const [autoPlaySpeed, setAutoPlaySpeed] = useState<AutoPlaySpeed>(loadAutoPlaySpeed);
  const [autoPlaying, setAutoPlaying] = useState(false);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  // Autoplay is driven client-side as a timed sequence of `step` calls (see the
  // effect below) so each play's stack/penalty reveal is watchable; the button toggles it.
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
    return execApi('reset', { maxRounds });
  }, [execApi, maxRounds]);

  // Latest state/speed read from a self-scheduling autoplay loop without
  // re-subscribing the effect on every render.
  const stateRef = useRef(state);
  stateRef.current = state;
  const speedRef = useRef(autoPlaySpeed);
  speedRef.current = autoPlaySpeed;

  // Drive autoplay client-side: recursively `step` on a speed-scaled timer so each
  // play's animation is visible, stopping as soon as the game ends.
  useEffect(() => {
    if (!autoPlaying) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    const tick = async () => {
      const s = stateRef.current;
      if (!s || s.gameEndFlag || s.phase === BeggarMyNeighbourPhase.GAME_END) {
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

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } =
    useCliMode('beggarmyneighbour');
  const cliConfig: CliGameConfig<BeggarMyNeighbourResponse, BeggarMyNeighbourArgs> = useMemo(
    () => ({
      gameName: 'beggarmyneighbour',
      parseCommand: parseBeggarMyNeighbourCommand,
      formatResponse: formatBeggarMyNeighbourState,
      helpText: BEGGARMYNEIGHBOUR_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state || state.players.length < 2)
    return <GameSkeleton gameKey="beggarmyneighbour" layout={{ kind: 'centered', rows: [2], gap: 'wide' }} />;

  const isGameEnd = state.gameEndFlag || state.phase === BeggarMyNeighbourPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  // Held-card totals and the who-is-winning ratio (out of the 52-card deck).
  const heldTotal = human.totalCards + cpu.totalCards;
  const youPct = heldTotal > 0 ? Math.round((human.totalCards / heldTotal) * 100) : 50;
  const cpuPct = 100 - youPct;
  // Round progress toward the draw-cutoff cap.
  const roundCap = state.config.maxRounds;
  const roundPct = roundCap > 0 ? Math.min(100, Math.round((state.roundsPlayed / roundCap) * 100)) : 0;

  const phaseName = isGameEnd
    ? t('phase.end')
    : state.phase === BeggarMyNeighbourPhase.PAY_PENALTY
      ? t('phase.payPenalty')
      : state.phase === BeggarMyNeighbourPhase.COLLECT
        ? t('phase.collect')
        : t('phase.play');

  // Phase transitions (and the penalty countdown) are conveyed only by the
  // central-pile ring color, so mirror them into an sr-only live region.
  const phaseAnnouncement =
    state.phase === BeggarMyNeighbourPhase.PAY_PENALTY
      ? t('phaseAnnouncePenalty', { phase: phaseName, count: state.penaltyRemaining })
      : t('phaseAnnounce', { phase: phaseName });

  return (
    <GamePageShell
      title={tc('nav.beggarmyneighbour')}
      gameThemeBg={gameTheme.beggarmyneighbour.bg}
      phaseName={phaseName}
      isHumanTurn={!isGameEnd}
      gamePath="/beggarmyneighbour"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-3">
            <ErrorAlert message={error} onRetry={retry} />

            {/* Announce the phase (and penalty countdown) to screen readers. */}
            <div className="sr-only" role="status" aria-live="polite" data-testid="bmn-phase-announce">
              {phaseAnnouncement}
            </div>

            {/* Held-card totals + round progress so the standings and how close the
                draw-cutoff is are visible without mental arithmetic. */}
            <div className="space-y-2 px-2" data-testid="bmn-summary">
              {/* Round progress toward the max-rounds cap. */}
              <div className="flex items-center gap-2">
                <div className="text-xs text-ds-text-muted whitespace-nowrap" data-testid="bmn-round-progress">
                  {t('readout.roundProgress', { played: state.roundsPlayed, max: roundCap })}
                </div>
                <div className="flex-1 h-1.5 bg-black/30 rounded-full overflow-hidden" aria-hidden="true">
                  <div className="h-full bg-ds-info" style={{ width: `${roundPct}%` }} />
                </div>
              </div>

              {/* Held-card counts and the who-is-ahead ratio bar. */}
              <div className="flex justify-between text-xs text-ds-text-primary" data-testid="bmn-card-counts">
                <span>
                  {tc('player.you')}: {t('readout.cardCount', { count: human.totalCards })}
                </span>
                <span>
                  {tc('player.cpu', { id: 1 })}: {t('readout.cardCount', { count: cpu.totalCards })}
                </span>
              </div>
              <div
                className="flex h-2 rounded-full overflow-hidden"
                role="img"
                aria-label={t('readout.countBarLabel', { you: human.totalCards, cpu: cpu.totalCards })}
                data-testid="bmn-count-bar"
                data-you-pct={youPct}
              >
                <div className="bg-ds-success h-full" style={{ width: `${youPct}%` }} />
                <div className="bg-ds-warning h-full" style={{ width: `${cpuPct}%` }} />
              </div>
            </div>

            {/* CPU pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="bmn-cpu-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.cpu', { id: 1 })} — {t('label.drawPile')}: {cpu.drawPileSize} / {t('label.discardPile')}:{' '}
                  {cpu.discardPileSize}
                </div>
                {/* Decorative: the count is already conveyed by the text line above,
                    so hide the card back from AT to avoid a redundant double read. */}
                <div aria-hidden="true">
                  {cpu.drawPileSize > 0 ? (
                    <AnimatedCardBack width={cardWidth * 0.9} />
                  ) : (
                    <div
                      className="rounded border border-white/20"
                      style={{ width: cardWidth * 0.9, height: cardWidth * 0.9 * 1.4 }}
                    />
                  )}
                </div>
              </div>
            </div>

            {/* Central pile area */}
            <div
              className={`flex items-center justify-center gap-8 py-3 bg-black/20 rounded-lg ${centralPileEmphasis(state.phase)}`}
              data-tutorial="bmn-central-pile"
            >
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.centralPile')}: {state.centralPileSize}
                </div>
                {state.phase === BeggarMyNeighbourPhase.PAY_PENALTY && (
                  <div className="text-xs text-ds-warning mt-1">
                    {t('label.penaltyRemaining')}: {state.penaltyRemaining}
                  </div>
                )}
                {state.centralPileSize > 0 ? (
                  <div
                    className="relative mx-auto mt-1"
                    aria-hidden="true"
                    data-testid="bmn-central-pile-stack"
                    data-pile-size={state.centralPileSize}
                    style={{
                      width: cardWidth * 0.8,
                      height: cardWidth * 0.8 * 1.4 + Math.min(state.centralPileSize, 10) * 2,
                    }}
                  >
                    {Array.from({ length: Math.min(state.centralPileSize, 10) }, (_, i) => (
                      <div
                        key={i}
                        className="absolute left-0 top-0"
                        style={{ transform: `translate(${i * 1.5}px, ${i * 2}px)` }}
                      >
                        <AnimatedCardBack width={cardWidth * 0.8} />
                      </div>
                    ))}
                  </div>
                ) : (
                  <div
                    className="rounded border border-dashed border-white/30 mx-auto mt-1"
                    style={{ width: cardWidth * 0.8, height: cardWidth * 0.8 * 1.4 }}
                  />
                )}
                {state.lastCardPlayed && (
                  <div className="mt-2">
                    <div className="text-xs text-ds-text-muted mb-1">{t('label.lastCard')}</div>
                    <AnimatedCard card={state.lastCardPlayed} width={cardWidth * 1.1} />
                  </div>
                )}
              </div>
            </div>

            {/* Player pile */}
            <div className="flex items-center justify-center gap-4" data-tutorial="bmn-player-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.you')} — {t('label.drawPile')}: {human.drawPileSize} / {t('label.discardPile')}:{' '}
                  {human.discardPileSize}
                </div>
                {/* Decorative: count is in the text line above; hide from AT. */}
                <div aria-hidden="true">
                  {human.drawPileSize > 0 ? (
                    <AnimatedCardBack width={cardWidth * 0.9} />
                  ) : (
                    <div
                      className="rounded border border-white/20"
                      style={{ width: cardWidth * 0.9, height: cardWidth * 0.9 * 1.4 }}
                    />
                  )}
                </div>
              </div>
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
          </div>

          <SettingsPanel
            title={tc('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'maxRounds',
                    label: t('settings.maxRounds'),
                    tooltip: t('settings.maxRoundsHelp'),
                    value: String(maxRounds),
                    options: [
                      { value: '500', label: '500' },
                      { value: '1000', label: '1000' },
                      { value: '2000', label: '2000' },
                      { value: '5000', label: '5000' },
                      { value: '10000', label: '10000' },
                    ],
                    onSelect: (v: string) => setMaxRounds(Number.parseInt(v, 10)),
                  },
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
                  {
                    type: 'checkbox' as const,
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          <GameFooter className={`${gameTheme.beggarmyneighbour.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd || autoPlaying}
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="bmn-step-button"
              >
                {t('button.step')}
              </button>
              <button
                type="button"
                onClick={handleAutoPlay}
                disabled={isGameEnd}
                aria-pressed={autoPlaying}
                className="px-6 py-2 rounded-lg bg-ds-success hover:bg-ds-success-hover text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="autoplay-button"
                data-tutorial="bmn-autoplay-button"
              >
                {autoPlaying ? t('button.stopAutoplay') : t('button.autoplay')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="bmn-reset-button"
              />
              <ActionLogSection
                isEndPhase={isGameEnd}
                actionLog={actionLog}
                showActionLog={showActionLog}
                hideActionLog={hideActionLog}
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
