import { useCallback, useMemo } from 'react';
import type { indianRummyApi } from '../api/gameApi';
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
import {
  CPU_DIFFICULTY_OPTIONS,
  PLAYER_COUNT_OPTIONS,
  TARGET_ROUNDS_OPTIONS,
  useIndianRummyGame,
} from '../hooks/useIndianRummyGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, IndianRummyResponse } from '../types/card';
import { IndianRummyPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { INDIANRUMMY_HELP, parseIndianrummyCommand } from '../utils/cli/commands/indianrummyCommands';
import { formatIndianrummyState } from '../utils/cli/formatters/indianrummyFormatter';
import { hintCliText, isHintCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { evaluateIndianRummyDeclare, INDIAN_RUMMY_HAND_SIZE } from '../utils/indianRummyDeclare';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const INDIANRUMMY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [IndianRummyPhase.DRAW]: 'draw',
  [IndianRummyPhase.DISCARD]: 'discard',
  [IndianRummyPhase.ROUND_END]: 'roundEnd',
  [IndianRummyPhase.GAME_END]: 'gameEnd',
};

/** Indian Rummy tutorial step definitions. */
const IR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ir-wild-joker"]',
    messageKey: 'tutorial.wildJoker',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ir-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ir-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ir-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ir-declare-button"]',
    messageKey: 'tutorial.declareButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ir-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ir-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Indian Rummy game page with draw, discard, and declare phases. */
export const IndianRummyPage = withTutorial(IndianRummyPageContent, 'indianrummy', IR_TUTORIAL_STEPS);
/** Inner content of the Indian Rummy page, wrapped by TutorialProvider. */
function IndianRummyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('indianrummy');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    indianRummyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleDeclare,
    handleNextRound,
  } = useIndianRummyGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('indianrummy', state);
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('indianrummy');
  const cliConfig: CliGameConfig<IndianRummyResponse, Parameters<typeof indianRummyApi.exec>> = useMemo(
    () => ({
      gameName: 'indianrummy',
      parseCommand: parseIndianrummyCommand,
      formatResponse: formatIndianrummyState,
      helpText: INDIANRUMMY_HELP,
      localCommand: (input: string) => (isHintCommand(input) ? hintCliText(frontendHint) : null),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === IndianRummyPhase.DISCARD;
  const isHumanTurnForKbd = isDiscardPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) handleDiscard();
  }, [isDiscardPhaseForKbd, handleDiscard]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('indianrummy', INDIANRUMMY_PHASE_KEYS);

  // Client-side declaration preview: when a finish card is selected during the
  // discard phase, evaluate whether the remaining 13 cards form a valid declare.
  // The backend still re-validates on submit; this only powers a non-blocking hint.
  const declarePreview = useMemo(() => {
    if (!state || state.phase !== IndianRummyPhase.DISCARD) return null;
    if (state.players[state.currentPlayerIdx]?.isHuman !== true) return null;
    if (selectedCardIndices.length !== 1) return null;
    const human = state.players.find((p) => p.isHuman);
    if (!human) return null;
    const finishIdx = selectedCardIndices[0];
    const remaining = human.cards.filter((_, i) => i !== finishIdx);
    if (remaining.length !== INDIAN_RUMMY_HAND_SIZE) return null;
    return evaluateIndianRummyDeclare(remaining, state.wildRank);
  }, [state, selectedCardIndices]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      playerCount: indianRummyConfig.playerCount,
      cpuDifficulty: indianRummyConfig.cpuDifficulty,
      targetRounds: indianRummyConfig.targetRounds,
    });
  }, [
    gameExec,
    hideActionLog,
    indianRummyConfig.playerCount,
    indianRummyConfig.cpuDifficulty,
    indianRummyConfig.targetRounds,
  ]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="indianrummy"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 13 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === IndianRummyPhase.DRAW;
  const isDiscardPhase = state.phase === IndianRummyPhase.DISCARD;
  const isRoundEnd = state.phase === IndianRummyPhase.ROUND_END;
  const isGameEnd = state.phase === IndianRummyPhase.GAME_END || state.gameEndFlag;
  const revealCpu = isRoundEnd || isGameEnd;
  const isHumanTurn = (isDrawPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  // A card counts as wild when it is a printed joker or shares the rank of the round's wild joker.
  const wildValue = state.wildJoker?.value;
  const isWildCard = (card: Card): boolean =>
    card.design === 'JOKER' || (wildValue !== undefined && card.value === wildValue);
  const wildBadge = (card: Card) =>
    isWildCard(card) ? (
      <span
        aria-hidden="true"
        className="absolute top-0.5 right-0.5 px-1 rounded bg-ds-info text-ds-text-on-accent text-[8px] font-extrabold tracking-wider shadow-md pointer-events-none"
        data-testid="ir-wild-badge"
      >
        {t('wildBadge')}
      </span>
    ) : null;

  return (
    <GamePageShell
      title={tc('nav.indianrummy')}
      gameThemeBg={gameTheme.indianrummy.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/indianrummy"
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
                    id: 'playerCount',
                    label: t('settings.playerCount'),
                    value: indianRummyConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: indianRummyConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: indianRummyConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, total: state.targetRounds })}</span>
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Wild joker indicator */}
                {state.wildJoker && (
                  <div
                    className="my-3 p-3 rounded bg-black/40 flex items-center gap-3"
                    data-tutorial="ir-wild-joker"
                    data-testid="indianrummy-wild-joker"
                  >
                    <AnimatedCard card={state.wildJoker} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('wildJoker')}</div>
                    </div>
                  </div>
                )}

                {/* Discard pile top */}
                {state.discardTop && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('discardTop')}</div>
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
                        {t('cumulativeScore', { score: p.cumulativeScore })}
                        {revealCpu && (
                          <>
                            {' '}
                            | {t('deadwoodShort', { score: p.deadwood })}
                            {p.hasPureSequence && (
                              <span className="ml-1 text-ds-success">{t('pureSequenceBadge')}</span>
                            )}
                          </>
                        )}
                      </div>
                      {revealCpu && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <div
                              key={`cpu-${p.id}-${card.design}-${card.value}-${idx}`}
                              className={`relative inline-block rounded ${
                                isWildCard(card) ? 'ring-2 ring-ds-info' : ''
                              }`}
                            >
                              <AnimatedCard card={card} width={cardWidth * 0.8} />
                              {isWildCard(card) && <span className="sr-only">{t('wildAria')}</span>}
                              {wildBadge(card)}
                            </div>
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="ir-score-table">
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

          <GameFooter className={`${gameTheme.indianrummy.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ir-player-hand">
                {humanPlayer.cards.map((card, idx) => (
                  <button
                    type="button"
                    key={`${card.design}-${card.value}-${idx}`}
                    onClick={() => toggleCard(idx)}
                    aria-label={`${cardAlt(card)}${isWildCard(card) ? ` ${t('wildAria')}` : ''}`}
                    aria-pressed={selectedCardIndices.includes(idx)}
                    className={`relative transition-transform ${focusRingCard} ${
                      isWildCard(card) ? 'ring-2 ring-ds-info' : ''
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

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {/* **CUI は毎ターン出している。**カードを選んでいない思案中でも、いまの
                デッドウッドとピュアシーケンス充足を確認できるようにする (#4824)。 */}
            {isDiscardPhase && isHumanTurn && humanPlayer && (
              // 読み上げは declarePreview 側の live region に任せる。隣り合う
              // live region が 2 つあると同じ変化を二重に告知する。
              <div className="mt-2 text-sm text-ds-text-muted" data-testid="indianrummy-hand-status">
                {t('deadwoodShort', { score: humanPlayer.deadwood })}
                <span className={`ml-2 ${humanPlayer.hasPureSequence ? 'text-ds-success' : 'text-ds-warning'}`}>
                  {humanPlayer.hasPureSequence ? t('pureSequenceBadge') : t('pureSequenceMissing')}
                </span>
              </div>
            )}

            {isDiscardPhase && isHumanTurn && declarePreview && (
              <div
                role="status"
                aria-live="polite"
                data-testid="indianrummy-declare-preview"
                className={`mb-2 px-3 py-2 rounded text-sm ${
                  declarePreview.valid ? badgeSuccessColors : badgeWarningColors
                }`}
              >
                {declarePreview.valid ? (
                  <span data-testid="indianrummy-declare-preview-valid">{t('declarePreview.valid')}</span>
                ) : (
                  <div data-testid="indianrummy-declare-preview-invalid">
                    <div className="font-semibold">{t('declarePreview.title')}</div>
                    <ul className="list-disc list-inside">
                      {!declarePreview.hasPureSequence && <li>{t('declarePreview.noPureSequence')}</li>}
                      {declarePreview.unmeldedCount > 0 && (
                        <li>
                          {t('declarePreview.unmelded', {
                            count: declarePreview.unmeldedCount,
                            points: declarePreview.unmeldedPoints,
                          })}
                        </li>
                      )}
                      {declarePreview.hasPureSequence && declarePreview.unmeldedCount === 0 && (
                        <li>{t('declarePreview.incomplete')}</li>
                      )}
                    </ul>
                    <div>{t('declarePreview.penalty', { penalty: declarePreview.penalty })}</div>
                  </div>
                )}
              </div>
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2" data-tutorial="ir-draw-area">
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
              {isDiscardPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDiscard}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="ir-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDeclare}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="ir-declare-button"
                    data-testid="indianrummy-declare-button"
                  >
                    {t('declareButton')}
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
                dataTutorial="ir-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="indian-rummy-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
