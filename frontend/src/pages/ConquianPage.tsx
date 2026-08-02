import { useCallback, useMemo } from 'react';
import type { conquianApi } from '../api/gameApi';
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
import { KbdBadge } from '../components/KbdBadge';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, TARGET_WINS_OPTIONS, useConquianGame } from '../hooks/useConquianGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ConquianPlayerData, ConquianResponse } from '../types/card';
import { ConquianPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CONQUIAN_HELP, parseConquianCommand } from '../utils/cli/commands/conquianCommands';
import { formatConquianState } from '../utils/cli/formatters/conquianFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Total cards a player must lay in table melds to win a Conquian round. */
const CONQUIAN_MELD_TARGET = 11;

const CONQUIAN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ConquianPhase.DRAW]: 'draw',
  [ConquianPhase.MELD]: 'meld',
  [ConquianPhase.ROUND_END]: 'roundEnd',
  [ConquianPhase.GAME_END]: 'gameEnd',
};

/** Conquian tutorial step definitions. */
const CQ_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cq-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="cq-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-meld-button"]',
    messageKey: 'tutorial.meldButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cq-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Conquian game page with draw and meld phases and per-player table melds. */
export const ConquianPage = withTutorial(ConquianPageContent, 'conquian', CQ_TUTORIAL_STEPS);
/** Inner content of the Conquian page, wrapped by TutorialProvider. */
function ConquianPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('conquian');
  const {
    state,
    loading,
    error,
    retry,
    gameExec,
    conquianConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleDrawStock,
    handleDrawDiscard,
    handleMeldSelected,
    handleDiscard,
    handleNextRound,
  } = useConquianGame();

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('conquian', CONQUIAN_PHASE_KEYS);

  const humanPlayer = state?.players.find((p) => p.isHuman);
  const humanCardCount = humanPlayer?.cards?.length ?? 0;

  // Meld-progress indicator: count the cards laid across a player's table melds
  // toward the 11-card win condition, highlighting the final stretch (<=2 left).
  const renderMeldProgress = (p: ConquianPlayerData, testId: string) => {
    const melded = p.melds.reduce((sum, m) => sum + m.cards.length, 0);
    const remaining = CONQUIAN_MELD_TARGET - melded;
    const isClose = remaining > 0 && remaining <= 2;
    const pct = Math.min(100, (melded / CONQUIAN_MELD_TARGET) * 100);
    const label = t('meldProgress', { count: melded, total: CONQUIAN_MELD_TARGET });
    return (
      <div data-testid={testId} className="mt-1">
        <div className="flex items-center justify-between text-xs mb-0.5">
          <span className="text-ds-text-muted">{label}</span>
          {isClose && <span className="text-ds-warning font-bold">{t('meldRemaining', { count: remaining })}</span>}
        </div>
        <div
          role="progressbar"
          aria-label={label}
          aria-valuemin={0}
          aria-valuemax={CONQUIAN_MELD_TARGET}
          aria-valuenow={melded}
          className="h-1.5 w-full rounded-sm bg-white/15 overflow-hidden"
        >
          <div
            className={`h-full rounded-sm ${isClose ? 'bg-ds-warning' : 'bg-ds-accent'}`}
            style={{ width: `${pct}%` }}
          />
        </div>
      </div>
    );
  };
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('conquian');
  const cliConfig: CliGameConfig<ConquianResponse, Parameters<typeof conquianApi.exec>> = useMemo(
    () => ({
      gameName: 'conquian',
      parseCommand: parseConquianCommand,
      formatResponse: formatConquianState,
      helpText: CONQUIAN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDrawPhase = state?.phase === ConquianPhase.DRAW;
  const isMeldPhase = state?.phase === ConquianPhase.MELD;
  const isRoundEnd = state?.phase === ConquianPhase.ROUND_END;
  const isGameEnd = state?.phase === ConquianPhase.GAME_END || !!state?.gameEndFlag;
  const isHumanTurn = (isDrawPhase || isMeldPhase) && state?.players[state.currentPlayerIdx]?.isHuman === true;

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: conquianConfig.cpuDifficulty,
      targetWins: conquianConfig.targetWins,
    });
  }, [gameExec, hideActionLog, conquianConfig.cpuDifficulty, conquianConfig.targetWins]);

  const kbdConfirmAction = useCallback(() => {
    if (isMeldPhase) handleMeldSelected();
  }, [isMeldPhase, handleMeldSelected]);

  useCardKeyboardNav({
    cardCount: humanCardCount,
    onToggle: toggleCard,
    onConfirm: kbdConfirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurn && !loading,
  });

  // DRAW phase: s draws from the stock, d takes the discard top. Number keys
  // and Enter/Escape are handled by useCardKeyboardNav; these letter keys don't
  // collide with it.
  const canDrawForKbd = isDrawPhase && !!isHumanTurn && !loading;
  const drawBindings = useMemo(
    () => [
      { key: 's', action: handleDrawStock, enabled: canDrawForKbd },
      { key: 'd', action: handleDrawDiscard, enabled: canDrawForKbd && !!state?.discardTop },
    ],
    [handleDrawStock, handleDrawDiscard, canDrawForKbd, state?.discardTop],
  );
  useActionKeyboardNav({ bindings: drawBindings, enabled: canDrawForKbd });

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('conquian', state);

  if (!state) {
    return (
      <GameSkeleton
        gameKey="conquian"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 10 }}
      />
    );
  }

  return (
    <GamePageShell
      title={tc('nav.conquian')}
      gameThemeBg={gameTheme.conquian.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/conquian"
      gameEndFlag={isGameEnd}
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
                    value: conquianConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetWins',
                    label: t('settings.targetWins'),
                    value: conquianConfig.targetWins,
                    options: TARGET_WINS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetWins', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

                {/* Discard pile top */}
                {state.discardTop && (
                  <div
                    className="my-3 p-3 rounded bg-black/40 flex items-center gap-3"
                    data-tutorial="cq-draw-area"
                    data-testid="cq-discard-pile"
                  >
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">{t('discardTop')}</div>
                  </div>
                )}

                {/* Per-player table melds */}
                {state.players.map((p) =>
                  p.melds.length === 0 ? null : (
                    <div key={`melds-${p.id}`} className="my-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm mb-1">
                        {playerName(p.id, p.isHuman)} - {t('tableMelds')}
                      </div>
                      {p.melds.map((m, mi) => (
                        <div key={`meld-${p.id}-${mi}`} className="flex flex-wrap gap-1 mb-1">
                          {m.cards.map((card, ci) => (
                            <AnimatedCard key={`meld-${p.id}-${mi}-${ci}`} card={card} width={cardWidth * 0.6} />
                          ))}
                        </div>
                      ))}
                    </div>
                  ),
                )}
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU player */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('wins', { count: p.wins })}
                      </div>
                      {renderMeldProgress(p, `conquian-meld-progress-cpu-${p.id}`)}
                      {/* Reveal CPU cards at round/game end */}
                      {(isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Wins table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="cq-score-table">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresWins')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.wins}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>

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

          <GameFooter className={`${gameTheme.conquian.footer} px-4 py-2.5`}>
            {isMeldPhase && isHumanTurn && state.tookDiscard && (
              <div className="text-xs font-bold mb-1 text-ds-info" data-testid="conquian-forced-use">
                {t('forcedUse')}
              </div>
            )}
            {humanPlayer && (
              <div className="mb-2 max-w-xs">{renderMeldProgress(humanPlayer, 'conquian-meld-progress')}</div>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="cq-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={cardAlt(card)}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`transition-transform ${focusRingCard}`}
                    style={{
                      background: 'none',
                      padding: 0,
                      borderRadius: 8,
                      ...selectedCardStyle(selectedCardIndices.includes(idx)),
                      boxSizing: 'border-box',
                    }}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDrawStock}
                    disabled={loading}
                    aria-keyshortcuts="s"
                  >
                    {t('drawStockButton')}
                    <KbdBadge label={t('kbd.stock')} />
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDrawDiscard}
                    disabled={loading || !state.discardTop}
                    aria-keyshortcuts="d"
                  >
                    {t('drawDiscardButton')}
                    <KbdBadge label={t('kbd.discard')} />
                  </button>
                </div>
              )}
              {isMeldPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeldSelected}
                    disabled={loading || selectedCardIndices.length < 3}
                    data-tutorial="cq-meld-button"
                    data-testid="conquian-meld-button"
                  >
                    {t('meldButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleMeldSelected}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-testid="conquian-layoff-button"
                  >
                    {t('layoffButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="cq-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                </>
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
                dataTutorial="cq-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
