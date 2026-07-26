import { useEffect, useMemo, useRef, useState } from 'react';
import type { spoilFiveApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, useSpoilFiveGame } from '../hooks/useSpoilFiveGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SpoilFiveResponse } from '../types/card';
import { SpoilFivePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSpoilFiveCommand, SPOIL_FIVE_HELP } from '../utils/cli/commands/spoilFiveCommands';
import { formatSpoilFiveState } from '../utils/cli/formatters/spoilFiveFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Suit number of Hearts (the ♥A is always a trump, so it needs special handling). */
const HEART_SUIT = 3;

/**
 * Builds the fixed named top-trump cards in strength order (high → low) for a given
 * trump suit, mirroring the Go domain `spoilRank`: trump 5 > trump J > ♥A > trump A >
 * trump K > trump Q. When Hearts is trump the ♥A *is* the trump ace, so it is not
 * duplicated.
 */
function spoilFiveTopTrumps(trumpSuit: number): string[] {
  const t = SUIT_SYMBOLS[trumpSuit] ?? '?';
  const cards = [`5${t}`, `J${t}`, 'A♥'];
  if (trumpSuit !== HEART_SUIT) cards.push(`A${t}`);
  cards.push(`K${t}`, `Q${t}`);
  return cards;
}

/** Spoil Five tutorial step definitions. */
const SPOIL_FIVE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="spoilfive-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spoilfive-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="spoilfive-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="spoilfive-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="spoilfive-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SPOIL_FIVE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpoilFivePhase.PLAY]: 'play',
  [SpoilFivePhase.TRICK_END]: 'trickEnd',
  [SpoilFivePhase.ROUND_END]: 'roundEnd',
  [SpoilFivePhase.GAME_END]: 'gameEnd',
};

/** Renders the Spoil Five game page: an Irish play-only 5-player trick-taker with a pot, fixed top trumps, and Reneging. */
export const SpoilFivePage = withTutorial(SpoilFivePageContent, 'spoilfive', SPOIL_FIVE_TUTORIAL_STEPS);

/** Inner content of the Spoil Five page, wrapped by TutorialProvider. */
function SpoilFivePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spoilfive');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    spoilFiveConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useSpoilFiveGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spoilfive');
  const spoilFiveCliConfig: CliGameConfig<SpoilFiveResponse, Parameters<typeof spoilFiveApi.exec>> = useMemo(
    () => ({
      gameName: 'spoilfive',
      parseCommand: parseSpoilFiveCommand,
      formatResponse: formatSpoilFiveState,
      helpText: SPOIL_FIVE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, spoilFiveCliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('spoilfive', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('spoilfive', SPOIL_FIVE_PHASE_KEYS);

  // Transient "+NN" feedback whenever the pot grows (e.g. a spoiled round carries
  // the pot forward). Cleared after a short delay so the pulse is momentary.
  const [potDelta, setPotDelta] = useState(0);
  const prevPotRef = useRef<number | null>(null);
  useEffect(() => {
    const pot = state?.pot;
    if (pot == null) return;
    const prev = prevPotRef.current;
    prevPotRef.current = pot;
    if (prev != null && pot > prev) {
      setPotDelta(pot - prev);
      const id = setTimeout(() => setPotDelta(0), 2500);
      return () => clearTimeout(id);
    }
  }, [state?.pot]);

  if (!state)
    return <GameSkeleton gameKey="spoilfive" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === SpoilFivePhase.PLAY;
  const isTrickEnd = state.phase === SpoilFivePhase.TRICK_END;
  const isRoundEnd = state.phase === SpoilFivePhase.ROUND_END;
  const isGameEnd = state.phase === SpoilFivePhase.GAME_END || state.gameEndFlag;
  const isSpoil = isRoundEnd && state.roundWinnerIdx < 0;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';
  const topTrumps = spoilFiveTopTrumps(state.trumpSuit);

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  /** Renders one player panel showing round tricks, match score, and leader / round-winner state. */
  const renderPlayerPanel = (p: SpoilFiveResponse['players'][number]) => {
    const isLeader = !isGameEnd && state.leadPlayerIdx === p.id;
    const isRoundWinner = (isRoundEnd || isGameEnd) && state.roundWinnerIdx === p.id;
    return (
      <div key={p.id} className="py-0.5 flex items-center gap-2">
        <span className="text-ds-text-muted">
          {playerName(p.id, p.isHuman)}
          {` — ${t('roundTricks', { count: p.roundTricks })} · ${t('score', { count: p.score })}`}
        </span>
        {isLeader && (
          <span className="px-1.5 py-0.5 rounded bg-white/20 text-ds-text-primary text-xs">{t('leader')}</span>
        )}
        {isRoundWinner && (
          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('roundWinner')}</span>
        )}
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.spoilfive')}
      gameThemeBg={gameTheme.spoilfive.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/spoilfive"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerPlayer === humanIdx}
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
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: spoilFiveConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span className={potDelta > 0 ? 'font-semibold text-ds-warning motion-safe:animate-pulse' : ''}>
                {t('pot', { n: state.pot })}
              </span>
              {potDelta > 0 && (
                <span
                  data-testid="spoilfive-pot-delta"
                  className="ml-1 text-sm font-semibold text-ds-warning motion-safe:animate-pulse"
                  aria-hidden="true"
                >
                  {t('potIncrease', { n: potDelta })}
                </span>
              )}
              {/* Screen-reader announcement of the pot change: a self-contained live
                  region so a spoiled round carrying the pot forward is spoken even
                  though the visual "+NN" pulse is transient and colour-only. */}
              <span className="sr-only" role="status" aria-live="polite" data-testid="spoilfive-pot-announce">
                {potDelta > 0 ? t('potIncreaseAnnounce', { delta: potDelta, total: state.pot }) : ''}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="spoilfive-trick-display"
                />
              </div>

              {/* Right: info sidebar — 5 player panels */}
              <div data-tutorial="spoilfive-info">
                {/* Top-trump ordering legend (collapsible, follows the current trump suit) */}
                <details className="mb-2 p-2 rounded bg-black/30" data-testid="spoilfive-trump-legend">
                  <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                    {t('trumpLegend.title')}
                  </summary>
                  <div className="mt-1 text-ds-text-muted text-xs">
                    <div className="mb-1">{t('trumpLegend.caption', { suit: trumpSymbol })}</div>
                    <div className="flex flex-wrap items-center gap-x-1 gap-y-0.5">
                      {topTrumps.map((card) => (
                        <span key={card} className="inline-flex items-center gap-1">
                          <span className="font-mono text-ds-text-primary">{card}</span>
                          <span aria-hidden="true">&gt;</span>
                        </span>
                      ))}
                      <span className="text-ds-text-primary">{t('trumpLegend.otherTrumps')}</span>
                      <span aria-hidden="true">&gt;</span>
                      <span className="text-ds-text-primary">{t('trumpLegend.nonTrump')}</span>
                    </div>
                  </div>
                </details>

                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1 text-sm">{state.players.map(renderPlayerPanel)}</div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30 text-sm">{state.players.map(renderPlayerPanel)}</div>
                )}

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {isSpoil ? (
                      <div>{t('roundResult.spoil')}</div>
                    ) : (
                      state.roundWinnerIdx >= 0 && (
                        <div>
                          {t('roundResult.winner', {
                            name: playerName(
                              state.roundWinnerIdx,
                              state.players[state.roundWinnerIdx]?.isHuman ?? false,
                            ),
                          })}
                        </div>
                      )
                    )}
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.spoilfive.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="spoilfive"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="spoilfive-action-buttons">
              {canPlay && (
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
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="spoilfive-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
