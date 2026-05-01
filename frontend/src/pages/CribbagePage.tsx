import { useCallback, useMemo } from 'react';
import type { cribbageApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetButton } from '../components/GameResetButton';
import { GameResetDialog } from '../components/GameResetDialog';
import { HintTooltip } from '../components/hint/HintTooltip';
import { ManualButton } from '../components/ManualButton';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { CribbageSkeleton } from '../components/skeleton/CribbageSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCribbageGame } from '../hooks/useCribbageGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CribbageResponse } from '../types/card';
import { CribbagePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { CRIBBAGE_HELP, parseCribbageCommand } from '../utils/cli/commands/cribbageCommands';
import { formatCribbageState } from '../utils/cli/formatters/cribbageFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

const CRIBBAGE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CribbagePhase.DISCARD]: 'discard',
  [CribbagePhase.CUT]: 'cut',
  [CribbagePhase.PEGGING]: 'pegging',
  [CribbagePhase.SHOW]: 'show',
  [CribbagePhase.ROUND_END]: 'roundEnd',
  [CribbagePhase.GAME_END]: 'gameEnd',
};

/** Cribbage tutorial step definitions. */
const CB_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cb-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-pegging-area"]',
    messageKey: 'tutorial.peggingArea',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-score-table"]',
    messageKey: 'tutorial.pegBoard',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Cribbage game page with discard, pegging, show, and round phases. */
export function CribbagePage() {
  return (
    <TutorialWrapper gameName="cribbage" steps={CB_TUTORIAL_STEPS}>
      <CribbagePageContent />
    </TutorialWrapper>
  );
}

/** Inline peg board showing score progress as a simple bar. */
function PegBoard({ scores, pointLimit }: { scores: { name: string; score: number }[]; pointLimit: number }) {
  return (
    <section className="my-2 p-2 rounded bg-black/30" aria-label="peg-board">
      {scores.map((p, idx) => {
        const pct = Math.min((p.score / pointLimit) * 100, 100);
        return (
          <div key={idx} className="mb-1">
            <div className="flex justify-between text-ds-text-muted text-xs mb-0.5">
              <span>{p.name}</span>
              <span>
                {p.score}/{pointLimit}
              </span>
            </div>
            <div className="w-full h-3 bg-white/10 rounded-full overflow-hidden">
              <div
                className={`h-full rounded-full transition-all ${idx === 0 ? 'bg-ds-warning' : 'bg-ds-info'}`}
                style={{ width: `${pct}%` }}
              />
            </div>
          </div>
        );
      })}
    </section>
  );
}

/** Inner content of the Cribbage page, wrapped by TutorialProvider. */
function CribbagePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cribbage');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    cribbageConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDiscard,
    handlePeg,
    handleGo,
    handleShowNext,
    handleNextRound,
  } = useCribbageGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cribbage', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cribbage');
  const cliConfig: CliGameConfig<CribbageResponse, Parameters<typeof cribbageApi.exec>> = useMemo(
    () => ({
      gameName: 'cribbage',
      parseCommand: parseCribbageCommand,
      formatResponse: formatCribbageState,
      helpText: CRIBBAGE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isDiscardPhaseForKbd = state?.phase === CribbagePhase.DISCARD;
  const isPeggingPhaseForKbd = state?.phase === CribbagePhase.PEGGING;
  const isHumanTurnForKbd =
    (isDiscardPhaseForKbd || isPeggingPhaseForKbd) && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    } else if (isPeggingPhaseForKbd) {
      handlePeg();
    }
  }, [isDiscardPhaseForKbd, isPeggingPhaseForKbd, handleDiscard, handlePeg]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('cribbage', CRIBBAGE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, undefined, {
      cpuDifficulty: cribbageConfig.cpuDifficulty,
      pointLimit: cribbageConfig.pointLimit,
    });
  }, [gameExec, hideActionLog, cribbageConfig.cpuDifficulty, cribbageConfig.pointLimit]);

  // Issue #1609: warn before tab close / reload while a round is in progress.

  useGameRoundGuard(!!state && !state.gameEndFlag);

  if (!state) return <CribbageSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDiscardPhase = state.phase === CribbagePhase.DISCARD;
  const isPeggingPhase = state.phase === CribbagePhase.PEGGING;
  const isShowPhase = state.phase === CribbagePhase.SHOW;
  const isRoundEnd = state.phase === CribbagePhase.ROUND_END;
  const isGameEnd = state.phase === CribbagePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = (isDiscardPhase || isPeggingPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;
  const canHumanPeg =
    isPeggingPhase &&
    isHumanTurn &&
    humanPlayer?.cards?.some((c) => {
      const cv = c.value >= 10 ? 10 : c.value;
      return state.pegCount + cv <= 31;
    });

  const nonDealerIsHuman = state.players[1 - state.dealerIdx]?.isHuman === true;
  const scoreLabels = [
    nonDealerIsHuman ? t('handScoreLabels.you') : t('handScoreLabels.cpu'),
    nonDealerIsHuman ? t('handScoreLabels.cpu') : t('handScoreLabels.you'),
    t('handScoreLabels.crib'),
  ];

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.cribbage.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.cribbage')} />
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanTurn}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/cribbage" />
      </PhaseIndicator>

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
                    value: cribbageConfig.cpuDifficulty,
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
                    value: cribbageConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
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

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span>{t('dealer', { name: playerName(state.dealerIdx, state.players[state.dealerIdx]?.isHuman) })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Starter card */}
                {state.starter && (
                  <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
                    <AnimatedCard
                      card={state.starter}
                      width={cardWidth}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                    <div className="text-ds-text-muted text-sm">
                      <div>{t('starter')}</div>
                    </div>
                  </div>
                )}

                {/* Pegging area */}
                {(isPeggingPhase || state.pegPlayedCards.length > 0) && (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="cb-pegging-area">
                    <div className="text-ds-text-muted text-sm mb-1">
                      {t('pegPlayedCards')} - {t('pegCount', { count: state.pegCount })}
                    </div>
                    <div className="flex flex-wrap gap-1">
                      {state.pegPlayedCards.map((card, idx) => (
                        <AnimatedCard
                          key={`peg-${card.design}-${card.value}-${idx}`}
                          card={card}
                          width={cardWidth * 0.8}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Crib (shown during show/round end/game end) */}
                {state.crib.length > 0 && (isShowPhase || isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm mb-1">{t('crib')}</div>
                    <div className="flex flex-wrap gap-1">
                      {state.crib.map((card, idx) => (
                        <AnimatedCard
                          key={`crib-${card.design}-${card.value}-${idx}`}
                          card={card}
                          width={cardWidth * 0.8}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Hand score details (show phase) */}
                {(isShowPhase || isRoundEnd || isGameEnd) && state.handScoreDetails.some((d) => d !== null) && (
                  <div className="my-3 p-2 rounded bg-black/30">
                    <div className="text-ds-text-muted text-sm mb-1">{t('score')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left" />
                          <th scope="col">{t('scoreDetail.fifteens')}</th>
                          <th scope="col">{t('scoreDetail.pairs')}</th>
                          <th scope="col">{t('scoreDetail.runs')}</th>
                          <th scope="col">{t('scoreDetail.flush')}</th>
                          <th scope="col">{t('scoreDetail.nobs')}</th>
                          <th scope="col">{t('scoreDetail.total')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.handScoreDetails.map((detail, idx) =>
                          detail ? (
                            <tr key={idx}>
                              <td>{scoreLabels[idx]}</td>
                              <td className="text-center">{detail.fifteens}</td>
                              <td className="text-center">{detail.pairs}</td>
                              <td className="text-center">{detail.runs}</td>
                              <td className="text-center">{detail.flush}</td>
                              <td className="text-center">{detail.nobs}</td>
                              <td className="text-center font-bold">{detail.total}</td>
                            </tr>
                          ) : null,
                        )}
                      </tbody>
                    </table>
                  </div>
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
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })}
                      </div>
                      {/* Show CPU cards during show/round end/game end */}
                      {(isShowPhase || isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                        <div className="flex flex-wrap gap-1 mt-1">
                          {p.cards.map((card, idx) => (
                            <AnimatedCard
                              key={`cpu-${card.design}-${card.value}-${idx}`}
                              card={card}
                              width={cardWidth * 0.8}
                              onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                            />
                          ))}
                        </div>
                      )}
                    </div>
                  ))}

                {/* Peg board */}
                <div data-tutorial="cb-score-table">
                  <PegBoard
                    scores={state.players.map((p) => ({
                      name: playerName(p.id, p.isHuman),
                      score: p.cumulativeScore,
                    }))}
                    pointLimit={state.config.pointLimit}
                  />
                </div>

                {/* Score table */}
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

          <GameFooter className={`${gameTheme.cribbage.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="cb-player-hand">
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
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {isDiscardPhase && isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 2}
                  data-tutorial="cb-discard-button"
                >
                  {t('discardButton')}
                </button>
              )}
              {isPeggingPhase && isHumanTurn && (
                <>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePeg}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('pegButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleGo} disabled={loading || !!canHumanPeg}>
                    {t('goButton')}
                  </button>
                </>
              )}
              {isShowPhase && (
                <button type="button" className={btnPrimary} onClick={handleShowNext} disabled={loading}>
                  {t('showNextButton')}
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
                dataTutorial="cb-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
      <WinCelebration show={!!state?.gameEndFlag} onCelebrate={() => playSound('winFanfare')} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
