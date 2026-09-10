import { useCallback, useEffect, useMemo } from 'react';
import type { marriageApi } from '../api/games/marriage';
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
  useMarriageGame,
} from '../hooks/useMarriageGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import type { MarriageResponse } from '../types/games/marriage';
import { MarriagePhase } from '../types/games/marriage';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { MARRIAGE_HELP, parseMarriageCommand } from '../utils/cli/commands/marriageCommands';
import { formatMarriageState } from '../utils/cli/formatters/marriageFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { evaluateMarriageDeclare, MARRIAGE_HAND_SIZE } from '../utils/marriageDeclare';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

const MARRIAGE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [MarriagePhase.DRAW]: 'draw',
  [MarriagePhase.DISCARD]: 'discard',
  [MarriagePhase.ROUND_END]: 'roundEnd',
  [MarriagePhase.GAME_END]: 'gameEnd',
};
const marriageTheme = gameTheme.ginrummy;

/** Marriage tutorial step definitions. */
const MARRIAGE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="marriage-wild-joker"]',
    messageKey: 'tutorial.wildJoker',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marriage-draw-area"]',
    messageKey: 'tutorial.drawArea',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marriage-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marriage-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marriage-declare-button"]',
    messageKey: 'tutorial.declareButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marriage-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="marriage-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Marriage game page with draw, discard, and declare phases. */
export const MarriagePage = withTutorial(MarriagePageContent, 'marriage', MARRIAGE_TUTORIAL_STEPS);
/** Inner content of the Marriage page, wrapped by TutorialProvider. */
function MarriagePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('marriage');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    marriageConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleDeclare,
    handleNextRound,
  } = useMarriageGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('marriage', state);
  const { cardWidth } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('marriage');
  const cliConfig: CliGameConfig<MarriageResponse, Parameters<typeof marriageApi.exec>> = useMemo(
    () => ({
      gameName: 'marriage',
      parseCommand: parseMarriageCommand,
      formatResponse: formatMarriageState,
      helpText: MARRIAGE_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  // biome-ignore lint/correctness/useExhaustiveDependencies: reset only on initial page mount
  useEffect(() => {
    void gameExec('reset', undefined, {
      playerCount: marriageConfig.playerCount,
      cpuDifficulty: marriageConfig.cpuDifficulty,
      targetRounds: marriageConfig.targetRounds,
    });
  }, [gameExec]);

  const isDiscardPhaseForKbd = state?.phase === MarriagePhase.DISCARD;
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

  const phaseNames = usePhaseNames('marriage', MARRIAGE_PHASE_KEYS);

  // Client-side declaration preview: when a finish card is selected during the
  // discard phase, evaluate whether the remaining 21 cards form a valid declare.
  // The backend still re-validates on submit; this only powers a non-blocking hint.
  const declarePreview = useMemo(() => {
    if (!state || state.phase !== MarriagePhase.DISCARD) return null;
    if (state.players[state.currentPlayerIdx]?.isHuman !== true) return null;
    if (selectedCardIndices.length !== 1) return null;
    const human = state.players.find((p) => p.isHuman);
    if (!human) return null;
    const finishIdx = selectedCardIndices[0];
    const remaining = human.cards.filter((_, i) => i !== finishIdx);
    if (remaining.length !== MARRIAGE_HAND_SIZE) return null;
    return evaluateMarriageDeclare(remaining, state.wildRank);
  }, [state, selectedCardIndices]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      playerCount: marriageConfig.playerCount,
      cpuDifficulty: marriageConfig.cpuDifficulty,
      targetRounds: marriageConfig.targetRounds,
    });
  }, [gameExec, hideActionLog, marriageConfig.playerCount, marriageConfig.cpuDifficulty, marriageConfig.targetRounds]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="marriage"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, trickArea: true, footerHandSize: 21 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === MarriagePhase.DRAW;
  const isDiscardPhase = state.phase === MarriagePhase.DISCARD;
  const isRoundEnd = state.phase === MarriagePhase.ROUND_END;
  const isGameEnd = state.phase === MarriagePhase.GAME_END || state.gameEndFlag;
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
        data-testid="marriage-wild-badge"
      >
        {t('wildBadge')}
      </span>
    ) : null;

  return (
    <GamePageShell
      title={tc('nav.marriage')}
      gameThemeBg={marriageTheme.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/marriage"
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
                    value: marriageConfig.playerCount,
                    options: PLAYER_COUNT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('playerCount', v),
                  },
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: marriageConfig.cpuDifficulty,
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
                    value: marriageConfig.targetRounds,
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
                    data-tutorial="marriage-wild-joker"
                    data-testid="marriage-wild-joker"
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
                            {p.maal > 0 && (
                              <span className="ml-1" data-testid="marriage-maal">
                                {t('maalShort', { score: p.maal })}
                              </span>
                            )}
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
                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="marriage-score-table">
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
                          <td className="text-center">
                            {p.cumulativeScore}
                            {p.maal > 0 && (
                              <span className="ml-1" data-testid="marriage-maal">
                                {t('maalShort', { score: p.maal })}
                              </span>
                            )}
                          </td>
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

          <GameFooter className={`${marriageTheme.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="marriage-player-hand">
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
              <div className="mt-2 text-sm text-ds-text-muted" data-testid="marriage-hand-status">
                {t('deadwoodShort', { score: humanPlayer.deadwood })}
                {humanPlayer.maal > 0 && (
                  <span className="ml-2" data-testid="marriage-maal">
                    {t('maalShort', { score: humanPlayer.maal })}
                  </span>
                )}
                <span className={`ml-2 ${humanPlayer.hasPureSequence ? 'text-ds-success' : 'text-ds-warning'}`}>
                  {humanPlayer.hasPureSequence ? t('pureSequenceBadge') : t('pureSequenceMissing')}
                </span>
                {/* **A も 10 点。** ジンラミー系に慣れたプレイヤーほど A=1 を
                    期待するので、合計だけ見せられると数字を逆算できない。
                    ワイルドが 0 点であることも同時に言う (#5501)。 */}
                <div className="text-xs" data-testid="marriage-points-legend">
                  {t('pointsLegend')}
                </div>
              </div>
            )}

            {isDiscardPhase && isHumanTurn && declarePreview && (
              <div
                role="status"
                aria-live="polite"
                data-testid="marriage-declare-preview"
                className={`mb-2 px-3 py-2 rounded text-sm ${
                  declarePreview.valid ? badgeSuccessColors : badgeWarningColors
                }`}
              >
                {declarePreview.valid ? (
                  <span data-testid="marriage-declare-preview-valid">{t('declarePreview.valid')}</span>
                ) : (
                  <div data-testid="marriage-declare-preview-invalid">
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
                <div className="flex gap-2" data-tutorial="marriage-draw-area">
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
                    data-tutorial="marriage-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handleDeclare}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="marriage-declare-button"
                    data-testid="marriage-declare-button"
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
                dataTutorial="marriage-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="indian-rummy-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
