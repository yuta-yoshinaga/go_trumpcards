import { useCallback, useEffect, useMemo, useRef } from 'react';
import type { doubtApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { CountdownBar } from '../components/doubt/CountdownBar';
import { DoubtCpuArea } from '../components/doubt/DoubtCpuArea';
import { DoubtHandCard } from '../components/doubt/DoubtHandCard';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCardSwipeSelection } from '../hooks/useCardSwipeSelection';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import {
  actionDesc,
  CPU_MEMORY_OPTIONS,
  DOUBT_WINDOW_OPTIONS,
  PENALTY_DRAW_LIMIT_OPTIONS,
  useDoubtGame,
} from '../hooks/useDoubtGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useSound } from '../providers/SoundProvider';
import { btnDanger, btnPrimary, btnSecondary, btnSuccess, focusRingAccent } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DoubtCpuAction, DoubtResponse } from '../types/card';
import { DoubtPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { valueName } from '../utils/cardUtils';
import { DOUBT_HELP, parseDoubtCommand } from '../utils/cli/commands/doubtCommands';
import { formatDoubtState } from '../utils/cli/formatters/doubtFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Doubt tutorial step definitions. */
const DT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="dt-table-area"]',
    messageKey: 'tutorial.tableArea',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-claim-input"]',
    messageKey: 'tutorial.claimInput',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-doubt-window"]',
    messageKey: 'tutorial.doubtWindow',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="dt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Doubt game page with card play, doubt window countdown, and config. */
export const DoubtPage = withTutorial(DoubtPageContent, 'doubt', DT_TUTORIAL_STEPS);
/** Inner content of the Doubt page, wrapped by TutorialProvider. */
function DoubtPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('doubt');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    countdown,
    doubtConfig,
    selectedCardIndices,
    toggleCard,
    claimedValue,
    setClaimedValue,
    handleConfigChange,
    handleConfigToggle,
    handlePlay,
    handleDoubt,
    handleSkip,
    handleCpuDoubtConfirm,
    clearSelection,
  } = useDoubtGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('doubt', state);

  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('doubt');
  const cliConfig: CliGameConfig<DoubtResponse, Parameters<typeof doubtApi.exec>> = useMemo(
    () => ({
      gameName: 'doubt',
      parseCommand: parseDoubtCommand,
      formatResponse: formatDoubtState,
      helpText: DOUBT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isHumanTurn = !state?.gameEndFlag && state?.players[state.currentTurn]?.isHuman === true;
  const showClaimInput = selectedCardIndices.length > 0 && isHumanTurn && state?.phase === 0;

  const isHumanPlayTurn = isHumanTurn && state?.phase === 0;
  // The "honest" next value is the last claim + 1, wrapping 13 → 1 (or 1 at the start).
  const honestValue = state?.lastAction != null ? (state.lastAction.claimedValue % 13) + 1 : 1;
  // Preset the claim selection to the honest value when the human's play turn begins,
  // so honest plays don't require hunting for the right button each turn.
  useEffect(() => {
    if (isHumanPlayTurn) setClaimedValue(honestValue);
  }, [isHumanPlayTurn, honestValue, setClaimedValue]);
  const humanPlayer = state?.players.find((p) => p.isHuman);

  useCardKeyboardNav({
    cardCount: humanPlayer?.cards?.length ?? 0,
    onToggle: toggleCard,
    onConfirm: handlePlay,
    onClear: clearSelection,
    enabled: isHumanPlayTurn && !loading,
  });

  const { onPointerDown: handleCardSwipeStart } = useCardSwipeSelection({
    selected: selectedCardIndices,
    toggle: toggleCard,
    enabled: isHumanPlayTurn && !loading,
  });

  // Keyboard shortcuts for the doubt decision window: Space = Doubt, Escape = Skip.
  // Only active while a CPU has just played and the human is being asked to judge.
  const isDoubtDecisionPhase =
    state?.phase === DoubtPhase.DOUBT &&
    state?.lastAction !== null &&
    !state?.players[state.lastAction.playerIdx]?.isHuman &&
    !state?.gameEndFlag;
  const doubtKeyRef = useRef({ handleDoubt, handleSkip });
  doubtKeyRef.current = { handleDoubt, handleSkip };
  useEffect(() => {
    if (!isDoubtDecisionPhase || loading) return;
    const onKey = (e: KeyboardEvent) => {
      const target = e.target as HTMLElement | null;
      if (
        target &&
        (target.tagName === 'INPUT' ||
          target.tagName === 'TEXTAREA' ||
          target.tagName === 'SELECT' ||
          target.isContentEditable)
      ) {
        return;
      }
      if (e.ctrlKey || e.altKey || e.metaKey) return;
      if (e.key === ' ' || e.key === 'Spacebar') {
        e.preventDefault();
        doubtKeyRef.current.handleDoubt();
      } else if (e.key === 'Escape' || e.key === 'Esc') {
        e.preventDefault();
        doubtKeyRef.current.handleSkip();
      }
    };
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [isDoubtDecisionPhase, loading]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void exec('reset', undefined, undefined, undefined, doubtConfig);
  }, [exec, hideActionLog, doubtConfig]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="doubt"
        layout={{
          kind: 'trick-taking',
          titleBar: false,
          opponents: 3,
          opponentStyle: 'hand',
          opponentHandSize: 4,
          trickArea: true,
          footerHandSize: 5,
        }}
      />
    );

  const cpuPlayers = state.players.filter((p) => !p.isHuman);
  const isDoubtPhase = state.phase === DoubtPhase.DOUBT;
  const cpuPlayed = isDoubtPhase && state.lastAction !== null && !state.players[state.lastAction.playerIdx]?.isHuman;

  const cpuTells = new Set(
    [...state.cpuActions, state.lastAction]
      .filter((a): a is DoubtCpuAction => a !== null && a.hasTell === true)
      .map((a) => a.playerIdx),
  );

  return (
    <GamePageShell
      title={tc('nav.doubt')}
      gameThemeBg={gameTheme.doubt.bg}
      phaseName={state.gameEndFlag ? t('phase.end') : state.phase === 1 ? t('phase.doubt') : t('phase.play')}
      isHumanTurn={isHumanTurn}
      gamePath="/doubt"
      gameEndFlag={!!state.gameEndFlag}
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
          {/* Settings panel */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'doubtWindowSec',
                    label: t('settings.doubtTime'),
                    value: doubtConfig.doubtWindowSec,
                    options: DOUBT_WINDOW_OPTIONS.map((sec) => ({ value: sec, label: t('settings.sec', { sec }) })),
                    onSelect: (v) => handleConfigChange('doubtWindowSec', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuMemoryLevel',
                    label: t('settings.cpuMemory'),
                    value: doubtConfig.cpuMemoryLevel,
                    options: CPU_MEMORY_OPTIONS.map((opt) => ({ value: opt.value, label: opt.label })),
                    onSelect: (v) => handleConfigChange('cpuMemoryLevel', v),
                  },
                  {
                    type: 'select',
                    id: 'penaltyDrawLimit',
                    label: t('settings.penaltyDrawLimit'),
                    value: doubtConfig.penaltyDrawLimit,
                    options: PENALTY_DRAW_LIMIT_OPTIONS.map((v) => ({
                      value: v,
                      label: v === 0 ? t('settings.unlimited') : t('settings.cards', { count: v }),
                    })),
                    onSelect: (v) => handleConfigChange('penaltyDrawLimit', v),
                  },
                  {
                    type: 'checkbox',
                    id: 'cpuHesitation',
                    label: t('settings.cpuHesitation'),
                    checked: doubtConfig.cpuHesitationEnabled,
                    onToggle: (checked) => handleConfigToggle('cpuHesitationEnabled', checked),
                  },
                  {
                    type: 'checkbox',
                    id: 'cpuMetaAI',
                    label: t('settings.cpuMetaAI'),
                    checked: doubtConfig.cpuMetaAI,
                    onToggle: (checked) => handleConfigToggle('cpuMetaAI', checked),
                  },
                  {
                    type: 'checkbox',
                    id: 'frontendHint',
                    label: tc('hint.toggle', { ns: 'tutorial' }),
                    checked: frontendHintEnabled,
                    onToggle: setFrontendHintEnabled,
                  },
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Table area */}
                <div className="bg-black/30 rounded-[10px] py-2.5 px-3.5 my-2" data-tutorial="dt-table-area">
                  <div className="text-ds-text-primary font-bold mb-1">{t('table')}</div>
                  <div className="text-game-text-muted text-sm">{t('tableCards', { count: state.tableCardCount })}</div>
                  {state.lastAction && (
                    <div className="text-ds-warning text-xs mt-1">{actionDesc(state.lastAction, state.players, t)}</div>
                  )}
                </div>

                {/* Doubt/Skip UI */}
                {isDoubtPhase && !state.gameEndFlag && (
                  <div className="bg-black/40 rounded-[10px] py-3 px-4 my-2" data-tutorial="dt-doubt-window">
                    {cpuPlayed ? (
                      <>
                        {/* Assertive one-shot alert so a screen-reader user is told a
                            time-limited decision has begun (it auto-skips otherwise).
                            Uses the fixed window length, not the live countdown, so the
                            alert text is stable and fires once per window rather than
                            re-announcing every second. */}
                        {state.lastAction && (
                          // Key on the running table count so two consecutive CPU turns
                          // with an identical claim still remount → re-announce the alert.
                          <div
                            key={state.tableCardCount}
                            className="sr-only"
                            role="alert"
                            data-testid="doubt-window-alert"
                          >
                            {t('doubtWindowAlert', {
                              action: actionDesc(state.lastAction, state.players, t),
                              sec: state.doubtWindowSec,
                            })}
                          </div>
                        )}
                        <div className="text-ds-text-primary font-bold mb-2">{t('doubtQuestion')}</div>
                        {state.lastAction && (
                          <div
                            className="bg-ds-warning/20 border-2 border-ds-warning rounded-lg py-2 px-3 mb-2 text-center animate-pulse"
                            data-testid="doubt-last-action-highlight"
                          >
                            <div className="text-ds-warning font-bold text-base sm:text-lg">
                              {actionDesc(state.lastAction, state.players, t)}
                            </div>
                          </div>
                        )}
                        {countdown !== null && (
                          <CountdownBar
                            remaining={countdown}
                            total={state.doubtWindowSec}
                            label={t('countdown', { sec: countdown })}
                          />
                        )}
                        {state.cpuDoubters.length > 0 && (
                          <div className="text-game-text-muted text-xs mb-2">
                            {t('cpuDoubters', {
                              names: state.cpuDoubters.map((idx) => playerName(idx, false)).join(', '),
                            })}
                          </div>
                        )}
                        <div className="flex gap-2">
                          <button type="button" className={btnDanger} disabled={loading} onClick={handleDoubt}>
                            {t('doubtButton')}
                          </button>
                          <button type="button" className={btnSecondary} disabled={loading} onClick={handleSkip}>
                            {t('skipButton')}
                          </button>
                        </div>
                        <p className="text-game-text-muted text-xs mt-2" data-testid="doubt-key-hints">
                          {t('keyHints')}
                        </p>
                      </>
                    ) : (
                      <>
                        <div className="text-ds-text-primary font-bold mb-2">{t('cpuJudging')}</div>
                        {state.cpuDoubters.length > 0 && (
                          <div className="text-ds-error text-sm mb-2">
                            {t('cpuDoubtExclaim', {
                              names: state.cpuDoubters.map((idx) => playerName(idx, false)).join(', '),
                            })}
                          </div>
                        )}
                        <button type="button" className={btnPrimary} disabled={loading} onClick={handleCpuDoubtConfirm}>
                          {t('confirmButton')}
                        </button>
                      </>
                    )}
                  </div>
                )}

                {/* Last doubt result */}
                {state.lastDoubtResult && (
                  <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-xs">
                    <div className="text-ds-text-primary font-bold mb-1">{t('doubtResult.title')}</div>
                    <div className={state.lastDoubtResult.wasLying ? 'text-ds-error' : 'text-ds-success'}>
                      {state.lastDoubtResult.wasLying ? t('doubtResult.wasLying') : t('doubtResult.wasTruth')}
                    </div>
                    <div className="text-game-text-muted">
                      {t('doubtResult.loserTook', {
                        name: playerName(
                          state.players[state.lastDoubtResult.loserIdx]?.id ?? state.lastDoubtResult.loserIdx,
                          state.players[state.lastDoubtResult.loserIdx]?.isHuman ?? false,
                        ),
                        count: state.lastDoubtResult.cardCount,
                      })}
                    </div>
                    {state.lastDoubtResult.discardedCount > 0 && (
                      <div className="text-ds-warning">
                        {t('doubtResult.discarded', { count: state.lastDoubtResult.discardedCount })}
                      </div>
                    )}
                    {state.lastDoubtResult.revealedCards.length > 0 && (
                      <div className="flex flex-wrap gap-1 mt-1">
                        {state.lastDoubtResult.revealedCards.map((card, i) => (
                          <AnimatedCard key={`${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                        ))}
                      </div>
                    )}
                  </div>
                )}

                {/* Human/CPU action logs */}
                {state.humanAction && !isDoubtPhase && (
                  <div className="bg-black/40 rounded-lg text-game-text-highlight py-2 px-3.5 my-2 text-xs">
                    {actionDesc(state.humanAction, state.players, t)}
                  </div>
                )}
                {state.cpuActions && state.cpuActions.length > 0 && (
                  <div className="bg-black/40 rounded-lg text-game-text-muted py-2 px-3.5 my-2 whitespace-pre-line text-xs">
                    {[tc('label.cpuActions'), ...state.cpuActions.map((a) => actionDesc(a, state.players, t))].join(
                      '\n',
                    )}
                  </div>
                )}

                {/* Result message */}
                <GameMessageBox
                  message={state.message}
                  messageCode={state.messageCode}
                  messageParams={state.messageParams}
                />

                {/* Action log */}
                <ActionLogSection
                  isEndPhase={state.gameEndFlag}
                  actionLog={actionLog}
                  showActionLog={showActionLog}
                  hideActionLog={hideActionLog}
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU player areas */}
                <div className="flex gap-2 flex-wrap mb-3 lg:flex-col">
                  {cpuPlayers.map((player) => (
                    <DoubtCpuArea
                      key={player.id}
                      player={player}
                      isCurrentTurn={state.currentTurn === player.id}
                      hasTell={cpuTells.has(player.id)}
                    />
                  ))}
                </div>

                {/* Meta-AI info */}
                {state.metaAI?.enabled && (
                  <div className="bg-black/40 rounded-lg py-2 px-3.5 my-2 text-xs">
                    <div className="text-ds-text-primary font-bold mb-1">{t('metaAI.title')}</div>
                    <div className="text-game-text-muted">
                      {t('metaAI.gamesPlayed', { count: state.metaAI.gamesPlayed })}
                    </div>
                    <div className="text-game-text-muted">
                      {t('metaAI.bluffRate', { rate: `${(state.metaAI.bluffRate * 100).toFixed(0)}%` })}
                    </div>
                    <div className="text-game-text-muted">
                      {t('metaAI.doubtAccuracy', { rate: `${(state.metaAI.doubtAccuracy * 100).toFixed(0)}%` })}
                    </div>
                    {state.metaAI.hesitationMean > 0 && (
                      <div className="text-game-text-muted">
                        {t('metaAI.hesitationMean', { ms: Math.round(state.metaAI.hesitationMean) })}
                      </div>
                    )}
                  </div>
                )}
              </div>
            </div>
          </div>

          {/* Sticky footer: human player hand + action buttons */}
          <GameFooter className={`${gameTheme.doubt.footer} px-4 py-2.5`}>
            {/* Human player info */}
            {humanPlayer && (
              <div className="mb-2" data-tutorial="dt-player-hand">
                <div className="text-ds-text-primary font-bold text-sm mb-1">
                  {t('yourCards', { count: humanPlayer.cardCount })}
                  {isHumanTurn && state.phase === 0 && (
                    <span className="text-ds-success text-xs ml-2">{t('selectPrompt')}</span>
                  )}
                </div>
                {/* Human cards */}
                <div className="flex flex-wrap gap-1">
                  {humanPlayer.cards?.map((card, i) => (
                    <DoubtHandCard
                      key={`${card.design}-${card.value}-${i}`}
                      card={card}
                      index={i}
                      selected={selectedCardIndices.includes(i)}
                      selectable={isHumanTurn && state.phase === 0 && !loading}
                      onToggle={toggleCard}
                      onSwipeStart={handleCardSwipeStart}
                    />
                  ))}
                </div>

                {/* Claimed value buttons (shown when cards are selected) */}
                {showClaimInput && (
                  <fieldset className="m-0 border-0 p-0 mt-2" data-tutorial="dt-claim-input">
                    <legend className="text-ds-text-primary text-sm mb-1">{t('claimedValue')}</legend>
                    <div className="flex flex-wrap gap-1.5">
                      {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => (
                        <button
                          key={v}
                          type="button"
                          className={
                            claimedValue === v
                              ? `${btnSecondary} min-w-[44px] ring-2 ring-ds-warning ${focusRingAccent}`
                              : v === honestValue
                                ? `${btnSecondary} min-w-[44px] ring-2 ring-ds-success ${focusRingAccent}`
                                : `${btnSecondary} min-w-[44px] ${focusRingAccent}`
                          }
                          aria-pressed={claimedValue === v}
                          data-testid={v === honestValue ? 'doubt-honest-value' : undefined}
                          onClick={() => setClaimedValue(v)}
                          disabled={loading}
                        >
                          {valueName(v)}
                        </button>
                      ))}
                    </div>
                  </fieldset>
                )}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            {/* Action buttons */}
            <div className="text-center">
              <GameResetButton
                isGameEnd={!!state.gameEndFlag}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="dt-reset-button"
                className="min-w-[90px]"
              />
              {isHumanTurn && state.phase === 0 && (
                <button
                  type="button"
                  className={`${btnSuccess} min-w-[90px]`}
                  disabled={loading || selectedCardIndices.length === 0}
                  onClick={handlePlay}
                  data-tutorial="dt-play-button"
                >
                  {t('playButton')}
                </button>
              )}
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
