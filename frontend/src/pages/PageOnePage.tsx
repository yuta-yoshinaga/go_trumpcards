import { useCallback, useMemo } from 'react';
import type { pageoneApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, usePageOneGame } from '../hooks/usePageOneGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { PageOneResponse } from '../types/card';
import { PageOnePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt, suitSymbol } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { PAGEONE_HELP, parsePageoneCommand } from '../utils/cli/commands/pageoneCommands';
import { formatPageoneState } from '../utils/cli/formatters/pageoneFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isPageOnePlayable } from '../utils/hints/pageoneHint';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const PAGEONE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PageOnePhase.PLAY]: 'play',
  [PageOnePhase.MUST_DECLARE]: 'mustDeclare',
  [PageOnePhase.ROUND_END]: 'roundEnd',
  [PageOnePhase.GAME_END]: 'gameEnd',
};

/** Page One tutorial step definitions. */
const PAGEONE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="po-discard-pile"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="po-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="po-play-draw"]', messageKey: 'tutorial.playDraw', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="po-declare"]', messageKey: 'tutorial.declare', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="po-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Page One game page with card play and declaration mechanic. */
export const PageOnePage = withTutorial(PageOnePageContent, 'pageone', PAGEONE_TUTORIAL_STEPS);
/** Inner content of the Page One page, wrapped by TutorialProvider. */
function PageOnePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('pageone');
  const {
    state,
    loading,
    error,
    exec: gameCall,
    retry,
    pageOneConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleDeclare,
    handleSkipDeclare,
    handleNextRound,
  } = usePageOneGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('pageone', state);
  const { cardWidth } = useCardDimensions();
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('pageone');
  const cliConfig: CliGameConfig<PageOneResponse, Parameters<typeof pageoneApi.exec>> = useMemo(
    () => ({
      gameName: 'pageone',
      parseCommand: parsePageoneCommand,
      formatResponse: formatPageoneState,
      helpText: PAGEONE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameCall, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === PageOnePhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('pageone', PAGEONE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameCall('reset', undefined, {
      cpuDifficulty: pageOneConfig.cpuDifficulty,
      pointLimit: pageOneConfig.pointLimit,
    });
  }, [gameCall, hideActionLog, pageOneConfig.cpuDifficulty, pageOneConfig.pointLimit]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="pageone"
        layout={{ kind: 'trick-taking', centerCard: true, trickArea: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === PageOnePhase.PLAY;
  const isMustDeclare = state.phase === PageOnePhase.MUST_DECLARE;
  const isRoundEnd = state.phase === PageOnePhase.ROUND_END;
  const isGameEnd = state.phase === PageOnePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanMustDeclare = isMustDeclare && state.players[state.currentPlayerIdx]?.isHuman === true;
  // Last-card alert: easy to miss the "Page One!" declaration window without strong feedback.
  // Use cardCount for both human + CPU so the "1 card left" check stays consistent across roles.
  const isGameActive = !isGameEnd && !isRoundEnd;
  const humanAtOneCard = (humanPlayer?.cardCount ?? 0) === 1 && !humanPlayer?.hasDeclared;
  const showLastCardBanner = isGameActive && humanAtOneCard;

  return (
    <GamePageShell
      title={tc('nav.pageone')}
      gameThemeBg={gameTheme.pageone.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isHumanMustDeclare}
      gamePath="/pageone"
      gameEndFlag={!!isGameEnd}
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
                    value: pageOneConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'pointLimit',
                    label: t('settings.pointLimit'),
                    value: pageOneConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
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
              <div>
                {state.discardTop && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3" data-tutorial="po-discard-pile">
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('discardTop')}</div>
                      <div className="text-xs text-ds-accent mt-0.5">
                        {t('playCondition', {
                          suit: suitSymbol(state.discardTop.design),
                          rank: valueName(state.discardTop.value),
                        })}
                      </div>
                    </div>
                  </div>
                )}

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

              <div>
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => {
                    const cpuAtOne = p.cardCount === 1 && !p.hasDeclared && isGameActive;
                    return (
                      <div
                        key={p.id}
                        data-testid={`po-cpu-${p.id}`}
                        className={`mb-2 p-2 rounded ${
                          cpuAtOne ? 'bg-ds-warning/15 ring-2 ring-ds-warning' : 'bg-black/30'
                        }`}
                      >
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })}
                          {p.hasDeclared ? ` • ${t('declaredBadge')}` : ''}
                          {cpuAtOne && (
                            <span
                              data-testid={`po-cpu-${p.id}-last-card-badge`}
                              role="status"
                              aria-live="polite"
                              aria-label={t('cpuLastCardAnnounce', { name: playerName(p.id, p.isHuman) })}
                              className={`ml-2 inline-flex items-center gap-1 px-2 py-0.5 rounded text-xs font-bold ${badgeWarningColors}`}
                            >
                              <span
                                aria-hidden="true"
                                className="inline-block w-2 h-2 rounded-full bg-ds-warning animate-pulse"
                              />
                              {t('cpuLastCardBadge')}
                            </span>
                          )}
                        </div>
                      </div>
                    );
                  })}

                <div className="my-3 p-2 rounded bg-black/30">
                  <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                  <table className="w-full text-sm text-ds-text-muted">
                    <thead>
                      <tr>
                        <th scope="col" className="text-left">
                          {t('scoresPlayer')}
                        </th>
                        <th scope="col">{t('scoresRound')}</th>
                        <th scope="col">{t('scoresTotal')}</th>
                      </tr>
                    </thead>
                    <tbody>
                      {state.players.map((p) => (
                        <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                          <td>{playerName(p.id, p.isHuman)}</td>
                          <td className="text-center">{p.roundScore}</td>
                          <td className="text-center">{p.cumulativeScore}</td>
                        </tr>
                      ))}
                    </tbody>
                  </table>
                </div>
              </div>
            </div>
          </div>

          <GameFooter className={`${gameTheme.pageone.footer} px-4 py-2.5`}>
            {showLastCardBanner && (
              <div
                role="alert"
                aria-live="assertive"
                data-testid="po-last-card-banner"
                className="mb-2 px-3 py-2 rounded bg-ds-surface border-2 border-ds-warning text-ds-warning text-sm font-bold flex items-center gap-2"
              >
                <span aria-hidden="true" className="inline-block w-2 h-2 rounded-full bg-ds-warning animate-pulse" />
                {t('lastCardBanner')}
              </div>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="po-player-hand">
                {humanPlayer.cards.map((card, idx) => {
                  // The CUI lists the legal indices every turn; the web player had
                  // to compare each card against the discard top by eye (#4744).
                  const playable = isHumanTurn && isPageOnePlayable(card, state.discardTop ?? null);
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={playable ? `${cardAlt(card)} (${t('playableAria')})` : cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      data-playable={playable || undefined}
                      className={`transition-transform ${focusRingCard} ${
                        playable && !selectedCardIndices.includes(idx)
                          ? 'ring-2 ring-ds-success motion-safe:animate-pulse rounded'
                          : ''
                      }`}
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
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
                <div className="flex gap-2" data-tutorial="po-play-draw">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                    {t('drawButton')}
                  </button>
                </div>
              )}
              {isHumanMustDeclare && (
                <div className="flex gap-2" data-tutorial="po-declare">
                  <button
                    type="button"
                    className={`${btnSuccess} ring-2 ring-ds-warning animate-pulse`}
                    onClick={handleDeclare}
                    disabled={loading}
                    data-testid="po-declare-btn"
                  >
                    {t('declareButton')}
                  </button>
                  <button type="button" className={btnWarning} onClick={handleSkipDeclare} disabled={loading}>
                    {t('skipButton')}
                  </button>
                </div>
              )}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="po-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="page-one-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
