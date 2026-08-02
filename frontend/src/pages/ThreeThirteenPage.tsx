import { useCallback, useMemo } from 'react';
import type { threethirteenApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, PLAYER_COUNT_OPTIONS, useThreeThirteenGame } from '../hooks/useThreeThirteenGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, ThreeThirteenResponse } from '../types/card';
import { ThreeThirteenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseThreeThirteenCommand, THREETHIRTEEN_HELP } from '../utils/cli/commands/threethirteenCommands';
import { formatThreeThirteenState } from '../utils/cli/formatters/threethirteenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { bestThreeThirteenDeadwoodValue, bestThreeThirteenDiscardValue } from '../utils/threethirteenDeadwood';

const THREETHIRTEEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [ThreeThirteenPhase.DRAW]: 'draw',
  [ThreeThirteenPhase.DISCARD]: 'discard',
  [ThreeThirteenPhase.ROUND_END]: 'roundEnd',
  [ThreeThirteenPhase.GAME_END]: 'gameEnd',
};

/** Three Thirteen tutorial step definitions. */
const TT_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tt-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="tt-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-knock-button"]',
    messageKey: 'tutorial.knockButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tt-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Three Thirteen game page with draw, discard, and knock phases. */
export const ThreeThirteenPage = withTutorial(ThreeThirteenPageContent, 'threethirteen', TT_TUTORIAL_STEPS);
/** Inner content of the Three Thirteen page, wrapped by TutorialProvider. */
function ThreeThirteenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('threethirteen');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    threeThirteenConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleNextRound,
  } = useThreeThirteenGame();
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('threethirteen');
  const cliConfig: CliGameConfig<ThreeThirteenResponse, Parameters<typeof threethirteenApi.exec>> = useMemo(
    () => ({
      gameName: 'threethirteen',
      parseCommand: parseThreeThirteenCommand,
      formatResponse: formatThreeThirteenState,
      helpText: THREETHIRTEEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === ThreeThirteenPhase.DISCARD;
  const isHumanTurnForKbd = isDiscardPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    }
  }, [isDiscardPhaseForKbd, handleDiscard]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('threethirteen', THREETHIRTEEN_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: threeThirteenConfig.cpuDifficulty,
      playerCount: threeThirteenConfig.playerCount,
    });
  }, [gameExec, hideActionLog, threeThirteenConfig.cpuDifficulty, threeThirteenConfig.playerCount]);

  // Predicted post-discard deadwood, so the player can weigh a discard before
  // committing. When one card is selected we project that exact discard;
  // otherwise we show the best value reachable across every single discard.
  // Wild-rank cards (state.wildRank) are matched into melds, so this mirrors the
  // server's own scoring. Computed unconditionally (before the early return) to
  // keep hook order stable; only surfaced during the DISCARD phase.
  const predictedDeadwood = useMemo<number | null>(() => {
    if (!state || state.phase !== ThreeThirteenPhase.DISCARD) return null;
    const human = state.players.find((p) => p.isHuman);
    if (!human || human.cards.length === 0) return null;
    const cards = human.cards;
    if (selectedCardIndices.length === 1) {
      return bestThreeThirteenDeadwoodValue(
        cards.filter((_, i) => i !== selectedCardIndices[0]),
        state.wildRank,
      );
    }
    return bestThreeThirteenDiscardValue(cards, state.wildRank);
  }, [state, selectedCardIndices]);

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('threethirteen', state);

  if (!state)
    return (
      <GameSkeleton
        gameKey="threethirteen"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 7 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === ThreeThirteenPhase.DRAW;
  const isDiscardPhase = state.phase === ThreeThirteenPhase.DISCARD;
  const isRoundEnd = state.phase === ThreeThirteenPhase.ROUND_END;
  const isGameEnd = state.phase === ThreeThirteenPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isDrawPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  // The wild rank changes each round; flag matching cards so players don't have to cross-check the header.
  const isWildCard = (card: Card): boolean => card.value === state.wildRank;
  const wildBadge = (card: Card) =>
    isWildCard(card) ? (
      <span
        aria-hidden="true"
        className="absolute top-0.5 right-0.5 px-1 rounded bg-ds-accent text-ds-text-on-accent text-[8px] font-extrabold tracking-wider shadow-md pointer-events-none"
        data-testid="tt-wild-badge"
      >
        {t('wildBadge')}
      </span>
    ) : null;

  // Prefer the projected post-discard deadwood (predictedDeadwood) during the
  // discard phase; fall back to the server's current-hand value otherwise.
  const humanDeadwood = predictedDeadwood ?? humanPlayer?.deadwood ?? null;

  return (
    <GamePageShell
      title={tc('nav.threethirteen')}
      gameThemeBg={gameTheme.threethirteen.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/threethirteen"
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
                    value: threeThirteenConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: threeThirteenConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-testid="threethirteen-round-banner">
              <span className="mr-4">{t('round', { n: state.round })}</span>
              <span className="mr-4 font-bold text-ds-accent">{t('wildRank', { rank: state.wildRank })}</span>
              <span className="mr-4">{t('dealCount', { count: state.dealCount })}</span>
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top */}
                {state.discardTop && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3" data-tutorial="tt-draw-area">
                    <div
                      className={`relative inline-block rounded ${isWildCard(state.discardTop) ? 'ring-2 ring-ds-accent' : ''}`}
                    >
                      <AnimatedCard card={state.discardTop} width={cardWidth} />
                      {wildBadge(state.discardTop)}
                    </div>
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('discardTop')}</div>
                      {isWildCard(state.discardTop) && (
                        <div className="text-ds-accent font-semibold">{t('wildAria')}</div>
                      )}
                    </div>
                  </div>
                )}
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
                      {/* Show CPU cards during round end / game end */}
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

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="tt-score-table">
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

          <GameFooter className={`${gameTheme.threethirteen.footer} px-4 py-2.5`}>
            {isDiscardPhase && humanDeadwood != null && (
              <div data-testid="threethirteen-deadwood-indicator" className="text-xs font-bold mb-1 text-ds-text-muted">
                {t('deadwoodLabel', { score: humanDeadwood })}
              </div>
            )}
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="tt-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={`${cardAlt(card)}${isWildCard(card) ? ` ${t('wildAria')}` : ''}`}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`relative transition-transform ${focusRingCard} ${
                      isWildCard(card) ? 'ring-2 ring-ds-accent' : ''
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
                    {wildBadge(card)}
                  </button>
                ))}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2">
                  <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                    {t('drawStockButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDrawDiscard}
                    disabled={loading || !state.discardTop}
                  >
                    {t('drawDiscardButton')}
                  </button>
                </div>
              )}
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

              {isDiscardPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="tt-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleKnock}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="tt-knock-button"
                    data-testid="threethirteen-knock-button"
                  >
                    {t('knockButton')}
                  </button>
                </>
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
                dataTutorial="tt-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="three-thirteen-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
