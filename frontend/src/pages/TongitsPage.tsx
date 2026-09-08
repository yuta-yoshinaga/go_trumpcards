import { useCallback, useMemo } from 'react';
import type { tongitsApi } from '../api/gameApi';
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
import { TongitsOnDealCelebration } from '../components/TongitsOnDealCelebration';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useTongitsGame } from '../hooks/useTongitsGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, playableCardStyle, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TongitsResponse } from '../types/card';
import { TongitsPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseTongitsCommand, TONGITS_HELP } from '../utils/cli/commands/tongitsCommands';
import { formatTongitsState } from '../utils/cli/formatters/tongitsFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { tongitsMeldIndices } from '../utils/tongitsMeldIndices';

const TONGITS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [TongitsPhase.DRAW]: 'draw',
  [TongitsPhase.DISCARD]: 'discard',
  [TongitsPhase.ROUND_END]: 'roundEnd',
  [TongitsPhase.GAME_END]: 'gameEnd',
};

/** Tongits tutorial step definitions. */
const TONGITS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tongits-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="tongits-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-knock-button"]',
    messageKey: 'tutorial.knockButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="tongits-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Tongits game page with draw, discard, and knock phases. */
export const TongitsPage = withTutorial(TongitsPageContent, 'tongits', TONGITS_TUTORIAL_STEPS);
/** Inner content of the Tongits page, wrapped by TutorialProvider. */
function TongitsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tongits');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    tongitsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleNextRound,
  } = useTongitsGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tongits', state);
  const { cardWidth } = useCardDimensions();
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tongits');
  const cliConfig: CliGameConfig<TongitsResponse, Parameters<typeof tongitsApi.exec>> = useMemo(
    () => ({
      gameName: 'tongits',
      parseCommand: parseTongitsCommand,
      formatResponse: formatTongitsState,
      helpText: TONGITS_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === TongitsPhase.DISCARD;
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

  const phaseNames = usePhaseNames('tongits', TONGITS_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, {
      cpuDifficulty: tongitsConfig.cpuDifficulty,
      pointLimit: tongitsConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, tongitsConfig.cpuDifficulty, tongitsConfig.pointLimit]);

  if (!state)
    return (
      <GameSkeleton
        gameKey="tongits"
        layout={{ kind: 'trick-taking', opponents: 1, centerCard: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === TongitsPhase.DRAW;
  const isDiscardPhase = state.phase === TongitsPhase.DISCARD;
  const isRoundEnd = state.phase === TongitsPhase.ROUND_END;
  const isGameEnd = state.phase === TongitsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isDrawPhase || isDiscardPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  // When exactly one card is selected to discard, highlight which of the remaining
  // four cards already form a meld (set or run) so the player can knock with confidence.
  const meldHighlight = (() => {
    if (!humanPlayer || !isDiscardPhase || !isHumanTurn || selectedCardIndices.length !== 1) {
      return new Set<number>();
    }
    const discardIdx = selectedCardIndices[0];
    const remaining = humanPlayer.cards.map((card, idx) => ({ card, idx })).filter((x) => x.idx !== discardIdx);
    const meldedPositions = tongitsMeldIndices(remaining.map((x) => x.card));
    const result = new Set<number>();
    for (const pos of meldedPositions) result.add(remaining[pos].idx);
    return result;
  })();
  // Undercut early-warning: if any opponent has 2 or fewer cards, calling Knock is
  // disproportionately risky (they're likely about to go out themselves, flipping the
  // result). We add a warning ring + ⚠️ glyph + tooltip so the player notices the
  // trap before committing. See issue #1939.
  const minOpponentCards = state.players
    .filter((p) => !p.isHuman)
    .reduce((m, p) => Math.min(m, p.cardCount), Number.POSITIVE_INFINITY);
  // 閾値はサーバから。画面に 2 を書くと、変えたとき CUI と警告の出る局面がずれる (#5582)。
  const undercutRisk = Number.isFinite(minOpponentCards) && minOpponentCards <= state.undercutRiskMax;
  const knockBtnClass = undercutRisk ? `${btnPrimary} ring-2 ring-ds-warning motion-safe:animate-pulse` : btnPrimary;

  return (
    <GamePageShell
      title={tc('nav.tongits')}
      gameThemeBg={gameTheme.tongits.bg}
      phaseName={phaseNames[state.phase] ?? ''}
      isHumanTurn={isHumanTurn}
      gamePath="/tongits"
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
                    value: tongitsConfig.cpuDifficulty,
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
                    value: tongitsConfig.pointLimit,
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
                {state.discardTop &&
                  (() => {
                    // The discard pile top card doubles as a draw target: on the human's
                    // draw turn, clicking it draws from the discard pile (same action as the
                    // footer "draw from discard" button). Outside that phase it stays inert
                    // but focusable-disabled so it never traps keyboard/tap focus.
                    const canDrawDiscard = isDrawPhase && isHumanTurn;
                    return (
                      <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                        <button
                          type="button"
                          onClick={handleDrawDiscard}
                          disabled={!canDrawDiscard || loading}
                          aria-label={t('drawDiscardCardLabel', { card: cardAlt(state.discardTop) })}
                          data-testid="tongits-discard-pile"
                          className={`transition-transform ${focusRingCard} ${
                            canDrawDiscard
                              ? 'cursor-pointer ring-2 ring-ds-info motion-safe:hover:scale-105'
                              : 'cursor-default'
                          }`}
                          style={{ background: 'none', padding: 0, border: 'none' }}
                        >
                          <AnimatedCard card={state.discardTop} width={cardWidth} />
                        </button>
                        <div className="text-ds-text-muted text-sm">
                          <div>{t('discardTop')}</div>
                        </div>
                      </div>
                    );
                  })()}

                {state.knockerMelds.length > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm mb-1">{t('knockerMelds')}</div>
                    {state.knockerMelds.map((meld, meldIdx) => (
                      <div key={`meld-${meldIdx}`} className="flex flex-wrap gap-1 mb-1">
                        {meld.cards.map((card, cardIdx) => (
                          <AnimatedCard
                            key={`meld-${meldIdx}-${card.design}-${card.value}-${cardIdx}`}
                            card={card}
                            width={cardWidth * 0.7}
                          />
                        ))}
                      </div>
                    ))}
                  </div>
                )}

                {/* ラウンドの点差はアンダーカット判定（両者のデッドウッド比較）から
                    来るのに、比較の相手側が画面に出ていなかった。 */}
                {state.knockerDeadwood.length > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30" data-testid="tongits-knocker-deadwood">
                    <div className="text-ds-text-muted text-sm mb-1">{t('knockerDeadwood')}</div>
                    <div className="flex flex-wrap gap-1">
                      {state.knockerDeadwood.map((card, cardIdx) => (
                        <AnimatedCard
                          key={`kdw-${card.design}-${card.value}-${cardIdx}`}
                          card={card}
                          width={cardWidth * 0.7}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {state.opponentMelds.length > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30" data-testid="tongits-opponent-melds">
                    <div className="text-ds-text-muted text-sm mb-1">{t('opponentMelds')}</div>
                    {state.opponentMelds.map((meld, meldIdx) => (
                      <div key={`opp-meld-${meldIdx}`} className="flex flex-wrap gap-1 mb-1">
                        {meld.cards.map((card, cardIdx) => (
                          <AnimatedCard
                            key={`opp-meld-${meldIdx}-${card.design}-${card.value}-${cardIdx}`}
                            card={card}
                            width={cardWidth * 0.7}
                          />
                        ))}
                      </div>
                    ))}
                  </div>
                )}

                {state.opponentDeadwood.length > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30" data-testid="tongits-opponent-deadwood">
                    <div className="text-ds-text-muted text-sm mb-1">{t('opponentDeadwood')}</div>
                    <div className="flex flex-wrap gap-1">
                      {state.opponentDeadwood.map((card, cardIdx) => (
                        <AnimatedCard
                          key={`odw-${card.design}-${card.value}-${cardIdx}`}
                          card={card}
                          width={cardWidth * 0.7}
                        />
                      ))}
                    </div>
                  </div>
                )}
              </div>

              <div>
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
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

                <div className="my-3 p-2 rounded bg-black/30" data-tutorial="tongits-score-table">
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

          <GameFooter className={`${gameTheme.tongits.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="tongits-player-hand">
                {humanPlayer.cards.map((card, idx) => {
                  const isCardSelected = selectedCardIndices.includes(idx);
                  const isMeldCard = !isCardSelected && meldHighlight.has(idx);
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={cardAlt(card)}
                      aria-pressed={isCardSelected}
                      data-meld={isMeldCard ? 'true' : undefined}
                      data-testid={`tongits-hand-${idx.toString()}`}
                      className={`transition-transform ${focusRingCard}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        // Selection wins; otherwise green-highlight cards that already form a meld.
                        ...(isCardSelected
                          ? selectedCardStyle(true)
                          : isMeldCard
                            ? playableCardStyle(true)
                            : selectedCardStyle(false)),
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
              {isDrawPhase && isHumanTurn && (
                <div className="flex gap-2" data-tutorial="tongits-draw-area">
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
                    data-tutorial="tongits-discard-button"
                  >
                    {t('discardButton')}
                  </button>
                  {/* **CUI は毎ターン「ノック可能/不可」を出しているのに、Web は
                      プレイヤーの手計算に任せていた。**ノックボタンは1枚選択されて
                      いれば常に活性化するので、押すまで合法か分からなかった。
                      -1 は「まだ聞くべき場面でない」印なので出さない。 */}
                  {state.bestDeadwood >= 0 && (
                    <span
                      className={`self-center text-xs ${
                        state.bestDeadwood <= state.knockThreshold ? 'text-ds-success' : 'text-ds-text-muted'
                      }`}
                      role="status"
                      aria-live="polite"
                      data-testid="tongits-deadwood"
                      data-knockable={state.bestDeadwood <= state.knockThreshold ? 'true' : undefined}
                    >
                      {t('deadwood.current', { value: state.bestDeadwood })}{' '}
                      {state.bestDeadwood <= state.knockThreshold
                        ? t('deadwood.knockable')
                        : t('deadwood.notKnockable', { threshold: state.knockThreshold })}
                    </span>
                  )}
                  <button
                    type="button"
                    className={knockBtnClass}
                    onClick={handleKnock}
                    disabled={loading || selectedCardIndices.length !== 1}
                    data-tutorial="tongits-knock-button"
                    data-undercut-risk={undercutRisk ? 'true' : undefined}
                    title={undercutRisk ? t('knockUndercutWarning') : undefined}
                  >
                    {t('knockButton')}
                    {undercutRisk && (
                      <span className="ml-1" aria-hidden="true">
                        ⚠️
                      </span>
                    )}
                  </button>
                  {undercutRisk && (
                    <div
                      className="text-ds-warning text-xs w-full text-center"
                      role="status"
                      data-testid="tongits-undercut-warning"
                    >
                      <span aria-hidden="true">⚠️ </span>
                      {t('knockUndercutWarning')}
                    </div>
                  )}
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
                dataTutorial="tongits-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="tongits-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
      {(() => {
        // Tongits-on-deal sets winnerIdx >= 0; guard defensively so we never index
        // players[-1] (which would surface as "CPU 0 wins" via playerName).
        const winner =
          state && state.winnerIdx >= 0 && state.winnerIdx < state.players.length
            ? state.players[state.winnerIdx]
            : undefined;
        return (
          <TongitsOnDealCelebration
            show={!!state?.isTongits && (isRoundEnd || isGameEnd) && winner !== undefined}
            winnerCards={winner?.cards ?? []}
            winnerName={winner ? playerName(winner.id, winner.isHuman) : undefined}
          />
        );
      })()}
    </GamePageShell>
  );
}
