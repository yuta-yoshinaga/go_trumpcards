import { useCallback, useMemo } from 'react';
import type { ohHellApi } from '../api/gameApi';
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
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import {
  CPU_DIFFICULTY_OPTIONS,
  MAX_HAND_SIZE_OPTIONS,
  ROUND_DIRECTION_OPTIONS,
  SCORING_VARIANT_OPTIONS,
  useOhHellGame,
} from '../hooks/useOhHellGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeErrorColors, badgeInfoColors, badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { OhHellResponse } from '../types/card';
import { OhHellPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { OHHELL_HELP, parseOhhellCommand } from '../utils/cli/commands/ohhellCommands';
import { formatOhhellState } from '../utils/cli/formatters/ohhellFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { ohHellBidSummary } from '../utils/ohHellBid';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Oh Hell tutorial step definitions. */
const OH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="oh-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="oh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Pick the badge color tokens for the bid-progress chip: green on target, yellow overshot, red unreachable. */
function progressChipColors(bid: number, won: number, remainingTricks: number): string {
  if (won === bid) return badgeSuccessColors;
  if (won > bid) return badgeWarningColors;
  return bid - won > remainingTricks ? badgeErrorColors : badgeInfoColors;
}

const OH_HELL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OhHellPhase.BID]: 'bid',
  [OhHellPhase.PLAY]: 'play',
  [OhHellPhase.TRICK_END]: 'trickEnd',
  [OhHellPhase.ROUND_END]: 'roundEnd',
  [OhHellPhase.GAME_END]: 'gameEnd',
};

/** Renders the Oh Hell game page with bidding, trick play, and scoring. */
export const OhHellPage = withTutorial(OhHellPageContent, 'ohhell', OH_TUTORIAL_STEPS);
/** Inner content of the Oh Hell page, wrapped by TutorialProvider. */
function OhHellPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ohhell');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    ohHellConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useOhHellGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ohhell', state);
  const { cardWidth, isMobile } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ohhell');
  const cliConfig: CliGameConfig<OhHellResponse, Parameters<typeof ohHellApi.exec>> = useMemo(
    () => ({
      gameName: 'ohhell',
      parseCommand: parseOhhellCommand,
      formatResponse: formatOhhellState,
      helpText: OHHELL_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === OhHellPhase.PLAY;
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

  const phaseNames = usePhaseNames('ohhell', OH_HELL_PHASE_KEYS);

  const gameAction = exec;
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameAction('reset', undefined, undefined, {
      cpuDifficulty: ohHellConfig.cpuDifficulty,
      maxHandSize: ohHellConfig.maxHandSize,
      scoringVariant: ohHellConfig.scoringVariant,
      roundDirection: ohHellConfig.roundDirection,
    });
  }, [
    gameAction,
    hideActionLog,
    ohHellConfig.cpuDifficulty,
    ohHellConfig.maxHandSize,
    ohHellConfig.scoringVariant,
    ohHellConfig.roundDirection,
  ]);

  if (!state)
    return <GameSkeleton gameKey="ohhell" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === OhHellPhase.BID;
  const isPlayPhase = state.phase === OhHellPhase.PLAY;
  const isTrickEnd = state.phase === OhHellPhase.TRICK_END;
  const isRoundEnd = state.phase === OhHellPhase.ROUND_END;
  const isGameEnd = state.phase === OhHellPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  const dealerName = playerName(
    state.players[state.dealerIdx]?.id ?? state.dealerIdx,
    state.players[state.dealerIdx]?.isHuman ?? false,
  );

  // A card already played into the unresolved trick still counts as a winnable trick,
  // even though cardCount has been decremented for it.
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const humanInCurrentTrick = state.currentTrick.some((tc) => tc.playerIdx === humanIdx);
  const showProgressChip = !isBidPhase && !isGameEnd && humanPlayer !== undefined && humanPlayer.bid >= 0;
  // During bidding, summarize the table's placed bids vs the hand size.
  const bidSummary = isBidPhase
    ? ohHellBidSummary(
        state.players.filter((p) => p.bid >= 0).map((p) => p.bid),
        state.handSize,
      )
    : null;

  return (
    <GamePageShell
      title={tc('nav.ohhell')}
      gameThemeBg={gameTheme.ohhell.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/ohhell"
      gameEndFlag={!!state.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          {showProgressChip && (
            <span
              data-testid="bid-progress-chip"
              className={`rounded-full px-2.5 py-1 text-xs font-medium ${progressChipColors(
                humanPlayer.bid,
                humanPlayer.trickCount,
                humanPlayer.cardCount + (humanInCurrentTrick ? 1 : 0),
              )}`}
            >
              {t('bidProgress', { bid: humanPlayer.bid, won: humanPlayer.trickCount })}
            </span>
          )}
          {bidSummary && (
            <span
              data-testid="bid-total-chip"
              className={`rounded-full px-2.5 py-1 text-xs font-medium ${
                bidSummary.kind === 'over'
                  ? badgeWarningColors
                  : bidSummary.kind === 'exact'
                    ? badgeSuccessColors
                    : badgeInfoColors
              }`}
            >
              {t('bidTotal', { total: bidSummary.total, handSize: state.handSize })}{' '}
              {t(`bidOverUnder.${bidSummary.kind}`)}
            </span>
          )}
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          {/* Settings */}
          <SettingsPanel
            title={t('settings.title')}
            groups={[
              {
                items: [
                  {
                    type: 'select',
                    id: 'cpuDifficulty',
                    label: t('settings.cpuDifficulty'),
                    value: ohHellConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'maxHandSize',
                    label: t('settings.maxHandSize'),
                    value: ohHellConfig.maxHandSize,
                    options: MAX_HAND_SIZE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('maxHandSize', v),
                  },
                  {
                    type: 'select',
                    id: 'scoringVariant',
                    label: t('settings.scoringVariant'),
                    value: ohHellConfig.scoringVariant,
                    options: SCORING_VARIANT_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.labelKey}`),
                    })),
                    onSelect: (v) => handleConfigChange('scoringVariant', v),
                  },
                  {
                    type: 'select',
                    id: 'roundDirection',
                    label: t('settings.roundDirection'),
                    value: ohHellConfig.roundDirection,
                    options: ROUND_DIRECTION_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.labelKey}`),
                    })),
                    onSelect: (v) => handleConfigChange('roundDirection', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick/Trump info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, total: state.totalRounds })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('handSize', { n: state.handSize })}</span>
            </div>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('trump', { suit: t(`suitName.${state.trumpSuit}`) })}</span>
              <span>{t('dealer', { name: dealerName })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Bid phase instruction */}
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="oh-bid-controls">
                    <div>{t('bidPhase', { max: state.handSize })}</div>
                    {state.restrictedBid >= 0 && (
                      <div className="text-ds-warning text-sm">{t('restrictedBid', { n: state.restrictedBid })}</div>
                    )}
                  </div>
                )}

                {/* Trump card display */}
                {state.trumpCard && (
                  <div className="my-2 flex justify-center">
                    <div className="text-center">
                      <AnimatedCard card={state.trumpCard} width={cardWidth} />
                    </div>
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="oh-trick-display"
                  winnerIdx={isTrickEnd ? state.leadPlayerIdx : undefined}
                  winnerLabel={t('trickWinnerBadge')}
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* CPU players */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {tc('label.cpuOpponents', { count: state.players.filter((p) => !p.isHuman).length })}
                    </summary>
                    <div className="mt-1">
                      {state.players
                        .filter((p) => !p.isHuman)
                        .map((p) => (
                          <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                            {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                            {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                            {t('roundScore', { score: p.roundScore })} |{' '}
                            {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                          </div>
                        ))}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => (
                      <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                        <div className="text-ds-text-muted text-sm">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                          {t('roundScore', { score: p.roundScore })} |{' '}
                          {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                        </div>
                      </div>
                    ))
                )}

                {/* Score table */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30 relative"
                    data-tutorial="oh-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('scores')}</summary>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[360px] mt-1">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresPlayer')}
                            </th>
                            <th scope="col">{t('scoresBid')}</th>
                            <th scope="col">{t('scoresTricks')}</th>
                            <th scope="col">{t('scoresRound')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {state.players.map((p) => (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                              <td className="text-center">{p.trickCount}</td>
                              <td className="text-center">{p.roundScore}</td>
                              <td className="text-center">{p.cumulativeScore}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                    <ScrollFadeHint />
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="oh-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('scores')}</div>
                    <div className="overflow-x-auto -mx-2 px-2">
                      <table className="w-full text-sm text-ds-text-muted min-w-[360px]">
                        <thead>
                          <tr>
                            <th scope="col" className="text-left">
                              {t('scoresPlayer')}
                            </th>
                            <th scope="col">{t('scoresBid')}</th>
                            <th scope="col">{t('scoresTricks')}</th>
                            <th scope="col">{t('scoresRound')}</th>
                            <th scope="col">{t('scoresTotal')}</th>
                          </tr>
                        </thead>
                        <tbody>
                          {state.players.map((p) => (
                            <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                              <td>{playerName(p.id, p.isHuman)}</td>
                              <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                              <td className="text-center">{p.trickCount}</td>
                              <td className="text-center">{p.roundScore}</td>
                              <td className="text-center">{p.cumulativeScore}</td>
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}
                <RoundScoreAnnouncement
                  active={isRoundEnd || isGameEnd}
                  entries={state.players.map((p) => ({
                    name: playerName(p.id, p.isHuman),
                    roundScore: p.roundScore,
                    cumulativeScore: p.cumulativeScore,
                  }))}
                />
              </div>
            </div>

            {/* Message */}
            <div>
              <GameMessageBox
                message={state.message}
                messageCode={state.messageCode}
                messageParams={state.messageParams}
              />
            </div>

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.ohhell.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer &&
              (isMobile ? (
                <MobileHandGrid
                  cards={humanPlayer.cards}
                  selectedIndices={selectedCardIndices}
                  onToggle={toggleCard}
                  cardWidth={cardWidth}
                  dataTutorial="oh-player-hand"
                />
              ) : (
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="oh-player-hand">
                  {humanPlayer.cards.map((card, idx) => (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      // プレイヒントの [N] もテキストだけだった。TwoTenJack と
                      // 同じく、該当する札を光らせる。
                      className={`transition-transform ${focusRingCard}${
                        isHumanTurn && hint?.cardIndex === idx ? ' rounded-lg ring-2 ring-ds-warning' : ''
                      }`}
                      data-hint-suggested={isHumanTurn && hint?.cardIndex === idx ? 'true' : undefined}
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
              ))}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
                付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
                読み上げられないことがある (#5955, #6663)。 */}
            <div data-testid="ohhell-hint-live" role="status" aria-live="polite">
              {hint && (
                <div className="text-ds-warning text-sm mb-2">
                  {hint.bid != null
                    ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
                    : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            <div className="flex gap-2 items-center" data-tutorial="oh-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                <div className="flex flex-wrap gap-1.5">
                  {Array.from({ length: state.handSize + 1 }, (_, i) => i).map((i) => {
                    const isRestricted = state.restrictedBid === i;
                    return (
                      <button
                        key={i}
                        type="button"
                        // Restricted bids use aria-disabled (not the HTML disabled
                        // attribute) so they stay focusable and a screen reader can
                        // read why they can't be chosen; the click is guarded instead.
                        // Mirrors the Cribbage pegRestricted / Call Break pattern.
                        // Neutralize btnPrimary's interactive feedback (press-scale,
                        // hover shadow) when restricted so it doesn't feel clickable.
                        // **ヒントは数字をテキストで言うだけで、どのボタンを
                        // 押せばよいか視覚的に示していなかった。**制限ビッドとは
                        // 別状態なので、強調は制限が無いときだけ付ける。
                        className={`${btnPrimary}${
                          isRestricted ? ' opacity-50 cursor-not-allowed active:scale-100 hover:shadow-none' : ''
                        }${!isRestricted && hint?.bid === i ? ' ring-2 ring-ds-warning' : ''}`}
                        onClick={() => {
                          if (!isRestricted) handleBid(i);
                        }}
                        disabled={loading}
                        aria-disabled={isRestricted || undefined}
                        title={isRestricted ? t('restrictedBidTooltip') : undefined}
                        aria-label={isRestricted ? t('restrictedBidAria', { n: i }) : t('bid', { n: i })}
                        data-testid={isRestricted ? 'ohhell-restricted-bid' : undefined}
                        data-hint-suggested={!isRestricted && hint?.bid === i ? 'true' : undefined}
                      >
                        {i}
                      </button>
                    );
                  })}
                </div>
              )}
              {isHumanTurn && (
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
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="oh-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="oh-hell-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
