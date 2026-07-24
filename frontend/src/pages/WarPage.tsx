import { useCallback, useEffect, useMemo, useRef, useState } from 'react';
import { warApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
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
import { useSound } from '../providers/SoundProvider';
import { gameTheme } from '../styles/gameTheme';
import type { WarResponse } from '../types/card';
import { WarPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import type { CliGameConfig, CliParseResult } from '../utils/cli/types';

type WarArgs = Parameters<typeof warApi.exec>;

const DEFAULT_MAX_ROUNDS = 500;

/** Autoplay animation speed presets. */
type AutoPlaySpeed = 'slow' | 'normal' | 'fast';

/**
 * Delay (ms) between auto-advanced `step` calls per speed preset. A larger delay
 * gives each round's reveal/war animation more time to play out; a smaller delay
 * races to the finish.
 */
const AUTOPLAY_DELAY_MS: Record<AutoPlaySpeed, number> = {
  slow: 900,
  normal: 450,
  fast: 150,
};

const AUTOPLAY_SPEED_STORAGE_KEY = 'war:autoPlaySpeed';

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

/** Tutorial steps for the War game. */
const WR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="wr-cpu-pile"]', messageKey: 'tutorial.cpuPile', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="wr-arena"]', messageKey: 'tutorial.arena', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="wr-player-pile"]',
    messageKey: 'tutorial.playerPile',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wr-step-button"]',
    messageKey: 'tutorial.stepButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wr-autoplay-button"]',
    messageKey: 'tutorial.autoplayButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/**
 * Pick the emphasis classes for a revealed card: green ring for the round winner
 * and a dimmed loser when resolved, yellow rings on both cards during a war.
 */
function revealedCardEmphasis(phase: number, lastWinnerIdx: number, isPlayer: boolean): string {
  if (phase === WarPhase.RESOLVED) {
    const won = isPlayer ? lastWinnerIdx === 0 : lastWinnerIdx === 1;
    return won ? 'ring-2 ring-ds-success' : 'opacity-60';
  }
  if (phase === WarPhase.WAR_BURY) {
    return 'ring-2 ring-ds-warning';
  }
  return '';
}

/** Renders the War (戦争) game page. */
export const WarPage = withTutorial(WarPageContent, 'war', WR_TUTORIAL_STEPS);
/** Inner content of the War page. */
function WarPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('war');
  const { state, loading, error, exec: execApi, retry } = useGameApi(warApi.exec);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  const [maxRounds, setMaxRounds] = useState(DEFAULT_MAX_ROUNDS);
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('war', state);

  const [autoPlaySpeed, setAutoPlaySpeed] = useState<AutoPlaySpeed>(loadAutoPlaySpeed);
  const [autoPlaying, setAutoPlaying] = useState(false);

  const handleStep = useCallback(() => execApi('step'), [execApi]);
  // Autoplay is driven client-side as a timed sequence of `step` calls (see the
  // effect below) so each round's reveal/war animation plays out; the button toggles it.
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
    playSound('shuffle');
    return execApi('reset', { maxRounds });
  }, [execApi, maxRounds, playSound]);

  // Play a card-resolution SFX each time a battle settles (RESOLVED) and a
  // tension cue when a tie triggers a war (WAR_BURY). Tracks the previous phase
  // so the sound fires on the transition, not on every re-render, and skips the
  // initial mount. Muting is honored inside `playSound`.
  const prevPhaseRef = useRef<number | undefined>(undefined);
  useEffect(() => {
    const phase = state?.phase;
    const prev = prevPhaseRef.current;
    prevPhaseRef.current = phase;
    if (phase === undefined || prev === undefined || prev === phase) return;
    if (phase === WarPhase.RESOLVED) playSound('cardPlace');
    else if (phase === WarPhase.WAR_BURY) playSound('chipClick');
  }, [state?.phase, playSound]);

  // Buzz once when a new error surfaces (e.g. a failed step/reset request).
  const prevErrorRef = useRef<string | null>(null);
  useEffect(() => {
    if (error && error !== prevErrorRef.current) playSound('errorBuzz');
    prevErrorRef.current = error;
  }, [error, playSound]);

  // Latest state/speed read from a self-scheduling autoplay loop without
  // re-subscribing the effect on every render.
  const stateRef = useRef(state);
  stateRef.current = state;
  const speedRef = useRef(autoPlaySpeed);
  speedRef.current = autoPlaySpeed;

  // Drive autoplay client-side: recursively `step` on a speed-scaled timer so each
  // round's reveal/war animation plays out, stopping as soon as the game ends.
  useEffect(() => {
    if (!autoPlaying) return;
    let cancelled = false;
    let timer: ReturnType<typeof setTimeout>;
    const tick = async () => {
      const s = stateRef.current;
      if (!s || s.gameEndFlag || s.phase === WarPhase.GAME_END) {
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
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('war');
  const cliConfig: CliGameConfig<WarResponse, WarArgs> = useMemo(
    () => ({
      gameName: 'war',
      parseCommand: (input: string): CliParseResult<WarArgs> => {
        const cmd = input.trim().toLowerCase();
        if (cmd === 'reset' || cmd === 'r') return { args: ['reset'] };
        if (cmd === 'step' || cmd === 's') return { args: ['step'] };
        if (cmd === 'autoplay' || cmd === 'a') return { args: ['autoplay'] };
        if (cmd === 'log' || cmd === 'l') return { args: ['log'] };
        return { error: `Unknown command: ${cmd}` };
      },
      formatResponse: (s: WarResponse) => {
        const lines: string[] = [];
        const phaseName = s.phase === WarPhase.WAR_BURY ? 'WAR' : s.phase === WarPhase.RESOLVED ? 'Resolved' : 'Reveal';
        lines.push(`Phase: ${phaseName} | Pot: ${s.warPotSize} | Round: ${s.roundsPlayed}`);
        for (const p of s.players) {
          const tag = p.isHuman ? 'You' : 'CPU';
          lines.push(`${tag}: draw=${p.drawPileSize} discard=${p.discardPileSize} total=${p.totalCards}`);
        }
        if (s.message) lines.push(s.message);
        return lines.join('\n');
      },
      helpText: [
        's/step     - Flip next card',
        'a/autoplay - Auto play to end',
        'r/reset    - Reset game',
        'l/log      - Show action log',
      ],
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  if (!state || state.players.length < 2)
    return <GameSkeleton gameKey="war" layout={{ kind: 'centered', rows: [2], gap: 'wide' }} />;

  const isGameEnd = state.gameEndFlag || state.phase === WarPhase.GAME_END;
  const humanWon = isGameEnd && state.winnerIdx === 0;
  const human = state.players[0];
  const cpu = state.players[1];

  const phaseName = isGameEnd
    ? t('phase.end')
    : state.phase === WarPhase.WAR_BURY
      ? t('phase.war')
      : state.phase === WarPhase.RESOLVED
        ? t('phase.resolved')
        : t('phase.reveal');

  return (
    <GamePageShell
      title={tc('nav.war')}
      gameThemeBg={gameTheme.war.bg}
      phaseName={phaseName}
      isHumanTurn={!isGameEnd}
      gamePath="/war"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      onCelebrate={() => playSound('winFanfare')}
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
            {error && (
              <button type="button" onClick={retry} className="text-ds-error underline">
                {error}
              </button>
            )}

            {/* CPU area */}
            <div className="flex items-center justify-center gap-4" data-tutorial="wr-cpu-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.cpu', { id: 1 })} — {t('label.drawPile')}: {cpu.drawPileSize} / {t('label.discardPile')}:{' '}
                  {cpu.discardPileSize}
                </div>
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

            {/* Revealed cards + pot */}
            <div
              className="flex items-center justify-center gap-8 py-3 bg-black/20 rounded-lg"
              data-tutorial="wr-arena"
            >
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">CPU</div>
                {state.cpuRevealed ? (
                  <AnimatedCard
                    card={state.cpuRevealed}
                    width={cardWidth * 1.1}
                    className={revealedCardEmphasis(state.phase, state.lastWinnerIdx, false)}
                  />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/30"
                    style={{ width: cardWidth * 1.1, height: cardWidth * 1.1 * 1.4 }}
                  />
                )}
              </div>
              <div className="text-center">
                <div className="text-sm text-ds-text-primary font-semibold">
                  {t('label.potCount', { count: state.warPotSize })}
                </div>
                {state.warPotSize > 2 && (
                  <div
                    // Purely decorative stacked card backs; the count is already
                    // conveyed by the adjacent label.potCount text, so hide from AT.
                    aria-hidden="true"
                    className="relative mx-auto mt-1"
                    data-testid="war-pot-stack"
                    data-pot-size={state.warPotSize}
                    style={{
                      width: cardWidth * 0.6,
                      height: cardWidth * 0.6 * 1.4 + Math.min(state.warPotSize, 10) * 2,
                    }}
                  >
                    {Array.from({ length: Math.min(state.warPotSize, 10) }, (_, i) => (
                      <div
                        key={i}
                        className="absolute left-0 top-0"
                        style={{ transform: `translate(${i * 1.5}px, ${i * 2}px)` }}
                      >
                        <AnimatedCardBack width={cardWidth * 0.6} />
                      </div>
                    ))}
                  </div>
                )}
                <div className="text-xs text-ds-text-muted mt-1">
                  {t('label.rounds')}: {state.roundsPlayed} / {state.config.maxRounds}
                </div>
              </div>
              <div className="text-center">
                <div className="text-xs text-ds-text-muted mb-1">{tc('player.you')}</div>
                {state.playerRevealed ? (
                  <AnimatedCard
                    card={state.playerRevealed}
                    width={cardWidth * 1.1}
                    className={revealedCardEmphasis(state.phase, state.lastWinnerIdx, true)}
                  />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/30"
                    style={{ width: cardWidth * 1.1, height: cardWidth * 1.1 * 1.4 }}
                  />
                )}
              </div>
            </div>

            {/* Player area */}
            <div className="flex items-center justify-center gap-4" data-tutorial="wr-player-pile">
              <div className="text-center">
                <div className="text-xs text-ds-text-muted">
                  {tc('player.you')} — {t('label.drawPile')}: {human.drawPileSize} / {t('label.discardPile')}:{' '}
                  {human.discardPileSize}
                </div>
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
                      { value: '100', label: '100' },
                      { value: '250', label: '250' },
                      { value: '500', label: '500' },
                      { value: '1000', label: '1000' },
                      { value: '2000', label: '2000' },
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

          <GameFooter className={`${gameTheme.war.footer} px-4 py-2.5`}>
            <div className="flex gap-2 justify-center">
              <button
                type="button"
                onClick={handleStep}
                disabled={loading || isGameEnd || autoPlaying}
                className="px-6 py-2 rounded-lg bg-ds-info hover:bg-ds-info text-white font-medium disabled:opacity-40 disabled:cursor-not-allowed"
                data-testid="step-button"
                data-tutorial="wr-step-button"
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
                data-tutorial="wr-autoplay-button"
              >
                {autoPlaying ? t('button.stopAutoplay') : t('button.autoplay')}
              </button>
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="wr-reset-button"
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
