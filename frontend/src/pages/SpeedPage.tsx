import { useCallback, useEffect, useMemo, useRef } from 'react';
import type { speedApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { AutoFlipCountdown } from '../components/AutoFlipCountdown';
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
import { SpeedSkeleton } from '../components/skeleton/SpeedSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { AUTO_FLIP_DELAY_MS, CPU_DIFFICULTY_OPTIONS, useSpeedGame } from '../hooks/useSpeedGame';
import { useSound } from '../providers/SoundProvider';
import { btnOutline } from '../styles/buttonStyles';
import { focusRingCard, playableRingStyle, selectedCardStyle } from '../styles/cardStyles';
import type { SpeedResponse } from '../types/card';
import { SpeedPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSpeedCommand, SPEED_HELP } from '../utils/cli/commands/speedCommands';
import { formatSpeedState } from '../utils/cli/formatters/speedFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isSpeedPlayable } from '../utils/hints/speedHint';

const SPEED_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sp-center-piles"]',
    messageKey: 'tutorial.centerPiles',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sp-draw-pile"]', messageKey: 'tutorial.drawPile', placement: 'left', advanceOn: 'next' },
];
/** Renders the Speed game page. */
export const SpeedPage = withTutorial(SpeedPageContent, 'speed', SPEED_TUTORIAL_STEPS);
/** Inner content of the Speed page. */
function SpeedPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('speed');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    speedConfig,
    selectedCardIndices,
    handlePlay,
    handleSmartClick,
    handleFlip,
    handleHint,
    handleConfigChange,
    handleToggle,
  } = useSpeedGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('speed', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('speed');
  const cliConfig: CliGameConfig<SpeedResponse, Parameters<typeof speedApi.exec>> = useMemo(
    () => ({
      gameName: 'speed',
      parseCommand: parseSpeedCommand,
      formatResponse: formatSpeedState,
      helpText: SPEED_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // Keyboard shortcuts: 1..N select a hand card; ←/→ play selected card to left/right center pile;
  // Space/Enter triggers Flip during STUCK phase.
  // Active only when CLI mode is off and a play phase or stuck phase is in progress. Respects
  // input focus and skips when modifier keys (Ctrl/Alt/Meta) are held to avoid hijacking browser
  // shortcuts. Latest state and handlers are read via refs so the listener is registered only once
  // per enable/disable transition rather than on every state update (e.g. each CPU move).
  const isPlayPhaseForKeys = state?.phase === SpeedPhase.PLAY;
  const isStuckForKeys = state?.phase === SpeedPhase.STUCK;
  const keyboardActive = isPlayPhaseForKeys || isStuckForKeys;
  const keyHandlersRef = useRef({ state, handleSmartClick, handlePlay, selectedCardIndices, handleFlip });
  keyHandlersRef.current = { state, handleSmartClick, handlePlay, selectedCardIndices, handleFlip };
  useEffect(() => {
    if (cliEnabled || !keyboardActive || loading) return;
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      if (target && (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA' || target.isContentEditable)) return;
      if (e.ctrlKey || e.altKey || e.metaKey) return;
      const cur = keyHandlersRef.current;
      if (!cur.state) return;
      if (cur.state.phase === SpeedPhase.STUCK) {
        if (e.key === ' ' || e.key === 'Spacebar' || e.key === 'Enter') {
          e.preventDefault();
          cur.handleFlip();
        }
        return;
      }
      if (e.key >= '1' && e.key <= '9') {
        const idx = Number.parseInt(e.key, 10) - 1;
        if (idx < cur.state.players[0].cards.length) {
          e.preventDefault();
          cur.handleSmartClick(idx, cur.state.players[0].cards, cur.state.centerPiles);
        }
        return;
      }
      if (e.key === 'ArrowLeft' || e.key === 'ArrowRight') {
        if (cur.selectedCardIndices.length === 1) {
          e.preventDefault();
          cur.handlePlay(e.key === 'ArrowLeft' ? 0 : 1);
        }
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [cliEnabled, keyboardActive, loading]);

  const handleManualReset = useCallback(() => {
    void gameExec('reset', undefined, undefined, speedConfig);
  }, [gameExec, speedConfig]);

  if (!state || state.players.length < 2) return <SpeedSkeleton />;

  const isPlayPhase = state.phase === SpeedPhase.PLAY;
  const isStuck = state.phase === SpeedPhase.STUCK;
  const isGameEnd = state.phase === SpeedPhase.GAME_END || state.gameEndFlag;
  const humanPlayer = state.players[0];
  const cpuPlayer = state.players[1];
  const humanWon = state.winnerIdx === 0;

  const phaseName = state.gameEndFlag
    ? t('phase.gameEnd')
    : state.phase === SpeedPhase.STUCK
      ? t('phase.stuck')
      : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.speed')}
      gameThemeBg="gap-2 p-2"
      phaseName={phaseName}
      isHumanTurn={isPlayPhase}
      gamePath="/speed"
      gameEndFlag={isGameEnd}
      winShow={!!isGameEnd && humanWon}
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
          {error && <ErrorAlert message={error} onRetry={retry} />}
          <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
          <div className="flex-1 flex flex-col gap-3 min-h-0">
            {/* CPU area */}
            <div className="flex items-center justify-center gap-2">
              <span className="text-sm text-ds-text-muted">
                {t('cpuHand')}: {cpuPlayer.cardCount}
              </span>
              <div className="flex gap-1">
                {Array.from({ length: cpuPlayer.cardCount }).map((_, i) => (
                  <AnimatedCardBack key={i} width={cardWidth * 0.7} />
                ))}
              </div>
              <span className="text-sm text-ds-text-muted">
                {t('drawPile')}: {cpuPlayer.drawPileSize}
              </span>
            </div>

            {/* Center piles — clickable for play (normal) or flip (stuck) */}
            <div className="relative flex items-center justify-center gap-6" data-tutorial="sp-center-piles">
              {state.centerPiles.map((card, pi) => (
                <button
                  type="button"
                  key={pi}
                  onClick={isStuck ? handleFlip : () => handlePlay(pi)}
                  disabled={isStuck ? loading : !isPlayPhase || selectedCardIndices.length !== 1 || loading}
                  className={`transition-transform hover:scale-105 disabled:opacity-50 ${focusRingCard}${isStuck && !loading ? ' animate-pulse cursor-pointer' : ''}`}
                  aria-label={
                    isStuck
                      ? t('flipCenterPile', { n: pi + 1 })
                      : t('centerPileCard', { n: pi + 1, card: cardAlt(card) })
                  }
                >
                  {card && <AnimatedCard card={card} width={cardWidth * 1.2} />}
                </button>
              ))}
              {isStuck && (
                // Centering container: keeps the popup glued to the center
                // even while animate-bounce drives its own transform on the
                // button. Pointer-events-none on the wrapper + auto on the
                // button so the wrapper doesn't intercept clicks on siblings.
                <div className="pointer-events-none absolute inset-0 z-10 flex items-center justify-center">
                  <button
                    type="button"
                    onClick={handleFlip}
                    disabled={loading}
                    data-testid="speed-stuck-flip-popup"
                    className="pointer-events-auto rounded-full bg-ds-accent px-5 py-3 text-base font-bold text-ds-text-on-accent shadow-2xl ring-4 ring-ds-accent/40 motion-safe:animate-bounce disabled:opacity-50 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ds-accent focus-visible:ring-offset-1 focus-visible:ring-offset-transparent"
                    aria-label={t('flipButton')}
                  >
                    ↻ {t('flipButton')}
                  </button>
                </div>
              )}
            </div>

            {/* Human hand */}
            <div className="flex flex-col items-center gap-1" data-tutorial="sp-player-hand">
              <div className="flex items-center gap-2">
                <span className="text-sm">{t('yourHand')}</span>
                <span className="text-sm text-ds-text-muted" data-tutorial="sp-draw-pile">
                  {t('drawPile')}: {humanPlayer.drawPileSize}
                </span>
              </div>
              <div className="flex gap-1 flex-wrap justify-center">
                {humanPlayer.cards.map((card, idx) => {
                  // Highlight cards playable right now (rank ±1 of either center
                  // pile, K↔A wrap). Ring-only: the button stays clickable so the
                  // backend still validates the play.
                  const playable = isPlayPhase && isSpeedPlayable(card.value, state.centerPiles);
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => handleSmartClick(idx, humanPlayer.cards, state.centerPiles)}
                      disabled={!isPlayPhase || loading}
                      aria-label={playable ? `${cardAlt(card)} (${t('playable')})` : cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      data-playable={playable ? 'true' : undefined}
                      className={`transition-transform ${focusRingCard}`}
                      style={{
                        ...selectedCardStyle(selectedCardIndices.includes(idx)),
                        ...(playable ? playableRingStyle() : {}),
                      }}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                    </button>
                  );
                })}
              </div>
            </div>

            {/* Stuck message, flip button, and inline auto-flip toggle */}
            {isStuck && (
              <div
                className="flex flex-col items-center gap-2 bg-ds-warning/10 ring-2 ring-ds-warning rounded-lg p-3"
                data-testid="stuck-emphasis-container"
                role="status"
                aria-live="polite"
              >
                <p className="text-ds-warning font-bold text-base sm:text-lg">{t('stuckMessage')}</p>
                {speedConfig.autoFlip && !loading && (
                  <div className="text-ds-warning">
                    <AutoFlipCountdown
                      key={`stuck-${state.message}-${state.centerPiles.map((c) => c?.value ?? 0).join(',')}`}
                      durationMs={AUTO_FLIP_DELAY_MS}
                      ariaLabel={t('autoFlipCountdownAria')}
                      formatRemaining={(n) => t('autoFlipCountdownRemaining', { n })}
                    />
                  </div>
                )}
                <button
                  type="button"
                  onClick={handleFlip}
                  disabled={loading}
                  className={`px-4 py-2 bg-ds-warning text-white rounded hover:bg-ds-warning-hover disabled:opacity-50 ring-2 ring-ds-warning${!loading ? ' animate-pulse' : ''}`}
                  data-testid="flip-button"
                >
                  {t('flipButton')}
                </button>
                <label className="flex items-center gap-2 text-sm text-ds-text-primary cursor-pointer min-h-[44px] px-1">
                  <input
                    type="checkbox"
                    checked={speedConfig.autoFlip}
                    onChange={(e) => handleToggle('autoFlip', e.target.checked)}
                    data-testid="inline-auto-flip-toggle"
                  />
                  {t('autoFlipInline')}
                </label>
              </div>
            )}

            {/* Hint */}
            {state.hint?.found && isPlayPhase && (
              <p className="text-center text-sm text-ds-info">
                {t('hint.play', { cardIndex: state.hint.cardIndex, pileIndex: state.hint.pileIndex })}
              </p>
            )}
            {frontendHintEnabled && frontendHint && (
              <div className="flex justify-center">
                <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
              </div>
            )}
          </div>

          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select' as const,
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: speedConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label}`),
                    })),
                    onSelect: (v: string) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'checkbox' as const,
                    id: 'autoFlip',
                    label: t('settings.autoFlip'),
                    checked: speedConfig.autoFlip,
                    onToggle: (v: boolean) => handleToggle('autoFlip', v),
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
          <GameFooter>
            <button type="button" onClick={handleHint} disabled={loading || !isPlayPhase} className={btnOutline}>
              {tc('button.hint')}
            </button>
            <ActionLogSection
              isEndPhase={!!isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
            <GameResetButton
              isGameEnd={!!isGameEnd}
              onReset={handleManualReset}
              requestConfirm={requestConfirm}
              loading={loading}
            />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
