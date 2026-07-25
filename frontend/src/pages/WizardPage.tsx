import { useCallback, useMemo } from 'react';
import type { wizardApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, useWizardGame } from '../hooks/useWizardGame';
import { useSound } from '../providers/SoundProvider';
import { badgeErrorColors, badgeInfoColors, badgeSuccessColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { WizardResponse } from '../types/card';
import { WizardPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseWizardCommand, WIZARD_HELP } from '../utils/cli/commands/wizardCommands';
import { formatWizardState } from '../utils/cli/formatters/wizardFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { ohHellBidSummary } from '../utils/ohHellBid';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';
import { type WizardBidOutcome, wizardBidAccuracy } from '../utils/wizardBidAccuracy';
import { isWizardLegalPlay } from '../utils/wizardLegal';

/** DOM id linking each illegal card button to the shared screen-reader reason text. */
const WIZARD_ILLEGAL_REASON_ID = 'wiz-illegal-reason';

/** Wizard tutorial step definitions. */
const WIZARD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="wiz-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wiz-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wiz-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wiz-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wiz-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="wiz-reset-button"]',
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

/** Badge color tokens for a bid-accuracy outcome pill: green made, yellow overshot, red undershot. */
function bidAccuracyPillColors(outcome: WizardBidOutcome): string {
  if (outcome === 'made') return badgeSuccessColors;
  return outcome === 'over' ? badgeWarningColors : badgeErrorColors;
}

const WIZARD_PHASE_KEYS: Readonly<Record<number, string>> = {
  [WizardPhase.BID]: 'bid',
  [WizardPhase.PLAY]: 'play',
  [WizardPhase.TRICK_END]: 'trickEnd',
  [WizardPhase.ROUND_END]: 'roundEnd',
  [WizardPhase.GAME_END]: 'gameEnd',
};

/** Renders the Wizard game page with bidding, trick play, and scoring. */
export const WizardPage = withTutorial(WizardPageContent, 'wizard', WIZARD_TUTORIAL_STEPS);
/** Inner content of the Wizard page, wrapped by TutorialProvider. */
function WizardPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('wizard');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    wizardConfig,
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
  } = useWizardGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('wizard', state);
  const { cardWidth, isMobile } = useCardDimensions();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('wizard');
  const cliConfig: CliGameConfig<WizardResponse, Parameters<typeof wizardApi.exec>> = useMemo(
    () => ({
      gameName: 'wizard',
      parseCommand: parseWizardCommand,
      formatResponse: formatWizardState,
      helpText: WIZARD_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === WizardPhase.PLAY;
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

  const phaseNames = usePhaseNames('wizard', WIZARD_PHASE_KEYS);

  const gameAction = exec;
  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameAction('reset', undefined, undefined, {
      cpuDifficulty: wizardConfig.cpuDifficulty,
    });
  }, [gameAction, hideActionLog, wizardConfig.cpuDifficulty]);

  if (!state)
    return <GameSkeleton gameKey="wizard" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === WizardPhase.BID;
  const isPlayPhase = state.phase === WizardPhase.PLAY;
  const isTrickEnd = state.phase === WizardPhase.TRICK_END;
  const isRoundEnd = state.phase === WizardPhase.ROUND_END;
  const isGameEnd = state.phase === WizardPhase.GAME_END || state.gameEndFlag;
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

  // On the human's play turn, compute which hand cards are legal to play so the
  // UI can highlight them. Mirrors Wizard.validatePlay in internal/domain/Wizard.go:
  // Wizard/Jester are always legal; a normal card must follow the led suit while
  // the player still holds it. `undefined` off-turn leaves every card unstyled.
  const trickCards = state.currentTrick.map((tc) => tc.card);
  const humanHand = humanPlayer?.cards ?? [];
  const legalIndices =
    isHumanTurn && humanPlayer
      ? humanHand.reduce<number[]>((acc, card, idx) => {
          if (isWizardLegalPlay(card, trickCards, humanHand)) acc.push(idx);
          return acc;
        }, [])
      : undefined;

  // At round/game end, `state.players` still carries the finished round's bid and
  // trickCount, so we can summarize how far each player's actual tricks landed
  // from their declared bid before the next deal resets them.
  const showBidAccuracy = isRoundEnd || isGameEnd;
  const bidAccuracyEntries = showBidAccuracy
    ? wizardBidAccuracy(
        state.players.map((p) => ({ name: playerName(p.id, p.isHuman), bid: p.bid, trickCount: p.trickCount })),
      )
    : [];

  return (
    <GamePageShell
      title={tc('nav.wizard')}
      gameThemeBg={gameTheme.wizard.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/wizard"
      gameEndFlag={!!state.gameEndFlag}
      onCelebrate={() => playSound('winFanfare')}
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
                    value: wizardConfig.cpuDifficulty,
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
                  <div className="text-ds-warning text-center mb-2" data-tutorial="wiz-bid-controls">
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
                  dataTutorial="wiz-trick-display"
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
                    data-tutorial="wiz-score-table"
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
                  <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="wiz-score-table">
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

                {/* Bid-vs-actual accuracy summary for the just-finished round. */}
                {showBidAccuracy && bidAccuracyEntries.length > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30" data-testid="wiz-bid-accuracy">
                    <div className="text-ds-text-muted text-sm mb-1">{t('bidAccuracy.title')}</div>
                    <ul className="flex flex-col gap-1">
                      {bidAccuracyEntries.map((e) => (
                        <li
                          key={e.name}
                          className="flex items-center justify-between gap-2 text-sm text-ds-text-muted"
                          data-testid={`wiz-bid-accuracy-row-${e.outcome}`}
                        >
                          <span className="truncate">
                            {e.name}: {t('scoresBid')} {e.bid} / {t('scoresTricks')} {e.trickCount}
                          </span>
                          <span
                            className={`shrink-0 rounded-full px-2 py-0.5 text-xs font-medium ${bidAccuracyPillColors(e.outcome)}`}
                          >
                            {e.outcome === 'made'
                              ? t('bidAccuracy.made')
                              : t(`bidAccuracy.${e.outcome}`, { delta: e.delta })}
                          </span>
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
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
          <GameFooter className={`${gameTheme.wizard.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer &&
              (isMobile ? (
                <MobileHandGrid
                  cards={humanPlayer.cards}
                  selectedIndices={selectedCardIndices}
                  onToggle={toggleCard}
                  cardWidth={cardWidth}
                  dataTutorial="wiz-player-hand"
                  validIndices={legalIndices}
                  restrictedTooltip={t('illegalHint')}
                />
              ) : (
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="wiz-player-hand">
                  {/* Shared screen-reader reason, referenced by every illegal card via
                      aria-describedby so the "why" is announced (title alone is skipped by SRs). */}
                  <span id={WIZARD_ILLEGAL_REASON_ID} className="sr-only">
                    {t('illegalHint')}
                  </span>
                  {humanPlayer.cards.map((card, idx) => {
                    // On the human's play turn, ring the legal cards and dim the rest with a
                    // reason tooltip so the follow-suit obligation is visible before playing.
                    const legal = legalIndices == null || legalIndices.includes(idx);
                    const showLegal = legalIndices != null;
                    return (
                      <button
                        type="button"
                        key={`${card.design}-${card.value}-${idx}`}
                        onClick={() => toggleCard(idx)}
                        aria-label={cardAlt(card)}
                        aria-pressed={selectedCardIndices.includes(idx)}
                        title={showLegal && !legal ? t('illegalHint') : undefined}
                        aria-describedby={showLegal && !legal ? WIZARD_ILLEGAL_REASON_ID : undefined}
                        data-legal={showLegal ? legal : undefined}
                        className={`transition-transform ${focusRingCard} ${
                          showLegal && legal ? 'rounded-lg ring-2 ring-ds-success' : ''
                        } ${showLegal && !legal ? 'opacity-50' : ''}`}
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
              ))}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.bid != null
                  ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
                  : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            <div className="flex gap-2 items-center" data-tutorial="wiz-play-button">
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
                        className={btnPrimary}
                        onClick={() => handleBid(i)}
                        disabled={loading || isRestricted}
                        title={isRestricted ? t('restrictedBidTooltip') : undefined}
                        aria-label={t('bid', { n: i })}
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
                dataTutorial="wiz-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
