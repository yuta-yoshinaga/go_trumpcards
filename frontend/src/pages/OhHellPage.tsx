import { useCallback, useMemo, useState } from 'react';
import type { ohHellApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { OhHellSkeleton } from '../components/skeleton/OhHellSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
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
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
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
import { playerName } from '../utils/playerUtils';

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

const OH_HELL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OhHellPhase.BID]: 'bid',
  [OhHellPhase.PLAY]: 'play',
  [OhHellPhase.TRICK_END]: 'trickEnd',
  [OhHellPhase.ROUND_END]: 'roundEnd',
  [OhHellPhase.GAME_END]: 'gameEnd',
};

/** Renders the Oh Hell game page with bidding, trick play, and scoring. */
export function OhHellPage() {
  return (
    <TutorialWrapper gameName="ohhell" steps={OH_TUTORIAL_STEPS}>
      <OhHellPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Oh Hell page, wrapped by TutorialProvider. */
function OhHellPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ohhell');
  const { playSound } = useSound();
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
  const [bidValue, setBidValue] = useState(0);
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

  if (!state) return <OhHellSkeleton />;

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

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.ohhell.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.ohhell')} />
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanBidTurn || isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/ohhell" />
      </PhaseIndicator>

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
            {/* Round/Trick/Trump info */}
            <div className="text-white text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber, total: state.totalRounds })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('handSize', { n: state.handSize })}</span>
            </div>
            <div className="text-white text-center mb-2">
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
                      <div className="text-orange-300 text-sm">{t('restrictedBid', { n: state.restrictedBid })}</div>
                    )}
                  </div>
                )}

                {/* Trump card display */}
                {state.trumpCard && (
                  <div className="my-2 flex justify-center">
                    <div className="text-center">
                      <AnimatedCard
                        card={state.trumpCard}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </div>
                  </div>
                )}

                {/* Current trick */}
                {state.currentTrick.length > 0 && (
                  <div className="my-3 p-3 rounded bg-black/40" data-tutorial="oh-trick-display">
                    <div className="text-white/70 text-sm mb-1">{t('currentTrick')}</div>
                    <div className="flex gap-2">
                      {state.currentTrick.map((trickCard) => (
                        <div key={`trick-${trickCard.playerIdx}`} className="text-center">
                          <AnimatedCard
                            card={trickCard.card}
                            width={cardWidth}
                            onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                          />
                          <div className="text-game-text-muted text-xs mt-1">
                            {playerName(
                              state.players[trickCard.playerIdx]?.id ?? trickCard.playerIdx,
                              state.players[trickCard.playerIdx]?.isHuman ?? false,
                            )}
                          </div>
                        </div>
                      ))}
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
                      <div className="text-white/70 text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })} |{' '}
                        {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')}
                      </div>
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="oh-score-table">
                  <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
                  <div className="overflow-x-auto -mx-2 px-2">
                    <table className="w-full text-sm text-white/70 min-w-[360px]">
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
                  {isMobile && <ScrollFadeHint />}
                </div>
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
                      className={`transition-transform ${focusRingCard}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        ...selectedCardStyle(selectedCardIndices.includes(idx)),
                        boxSizing: 'border-box',
                      }}
                    >
                      <AnimatedCard
                        card={card}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </button>
                  ))}
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
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}
            <div className="flex gap-2 items-center" data-tutorial="oh-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}
              {isHumanBidTurn && (
                <>
                  <input
                    type="number"
                    min={0}
                    max={state.handSize}
                    value={bidValue}
                    onChange={(e) => setBidValue(Number(e.target.value))}
                    className="w-16 px-2 py-1 rounded bg-white/20 text-white text-center"
                    aria-label="bid-input"
                  />
                  <button type="button" className={btnPrimary} onClick={() => handleBid(bidValue)} disabled={loading}>
                    {t('bidButton')}
                  </button>
                </>
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
              <button
                type="button"
                className={btnOutline}
                data-tutorial="oh-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    return exec('reset', undefined, undefined, {
                      cpuDifficulty: ohHellConfig.cpuDifficulty,
                      maxHandSize: ohHellConfig.maxHandSize,
                      scoringVariant: ohHellConfig.scoringVariant,
                      roundDirection: ohHellConfig.roundDirection,
                    });
                  })
                }
                disabled={loading}
              >
                {tc('button.reset')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={!!state?.gameEndFlag} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
