import { useCallback, useMemo, useState } from 'react';
import type { euchreApi } from '../api/gameApi';
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
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useEuchreGame } from '../hooks/useEuchreGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useGameRoundGuard } from '../hooks/useGameRoundGuard';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { EuchreResponse } from '../types/card';
import { EuchrePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { EUCHRE_HELP, parseEuchreCommand } from '../utils/cli/commands/euchreCommands';
import { formatEuchreState } from '../utils/cli/formatters/euchreFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

const SUIT_NAMES: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Euchre tutorial step definitions. */
const EU_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="eu-pickup-controls"]',
    messageKey: 'tutorial.pickUpControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-team-info"]',
    messageKey: 'tutorial.teamInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const EUCHRE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [EuchrePhase.PICK_UP]: 'pickUp',
  [EuchrePhase.CALL_TRUMP]: 'callTrump',
  [EuchrePhase.DISCARD]: 'discard',
  [EuchrePhase.PLAY]: 'play',
  [EuchrePhase.TRICK_END]: 'trickEnd',
  [EuchrePhase.ROUND_END]: 'roundEnd',
  [EuchrePhase.GAME_END]: 'gameEnd',
};

/** Renders the Euchre game page with pick-up, trump calling, trick play, and team scoring. */
export const EuchrePage = withTutorial(EuchrePageContent, 'euchre', EU_TUTORIAL_STEPS);
/** Inner content of the Euchre page, wrapped by TutorialProvider. */
function EuchrePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('euchre');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    retry,
    apiExec,
    euchreConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleOrderUp,
    handlePass,
    handleCallTrump,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useEuchreGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('euchre', state);
  const { cardWidth, isMobile, solitaireMinColWidth } = useCardDimensions();
  const [goAlone, setGoAlone] = useState(false);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('euchre');
  const cliConfig: CliGameConfig<EuchreResponse, Parameters<typeof euchreApi.exec>> = useMemo(
    () => ({
      gameName: 'euchre',
      parseCommand: parseEuchreCommand,
      formatResponse: formatEuchreState,
      helpText: EUCHRE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === EuchrePhase.PLAY;
  const isDiscardPhaseForKbd = state?.phase === EuchrePhase.DISCARD;
  const isHumanTurnForKbd =
    (isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true) ||
    (isDiscardPhaseForKbd && state?.players[state.dealerIdx]?.isHuman === true);
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    } else {
      handlePlay();
    }
  }, [handlePlay, handleDiscard, isDiscardPhaseForKbd]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('euchre', EUCHRE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiExec('reset', undefined, undefined, undefined, {
      cpuDifficulty: euchreConfig.cpuDifficulty,
      pointLimit: euchreConfig.pointLimit,
    });
  }, [apiExec, hideActionLog, euchreConfig.cpuDifficulty, euchreConfig.pointLimit]);

  useGameRoundGuard(!!state && !state.gameEndFlag);

  if (!state)
    return <GameSkeleton gameKey="euchre" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isPickUpPhase = state.phase === EuchrePhase.PICK_UP;
  const isCallTrumpPhase = state.phase === EuchrePhase.CALL_TRUMP;
  const isDiscardPhase = state.phase === EuchrePhase.DISCARD;
  const isPlayPhase = state.phase === EuchrePhase.PLAY;
  const isTrickEnd = state.phase === EuchrePhase.TRICK_END;
  const isRoundEnd = state.phase === EuchrePhase.ROUND_END;
  const isGameEnd = state.phase === EuchrePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = (isPickUpPhase || isCallTrumpPhase) && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanDiscard = isDiscardPhase && state.players[state.dealerIdx]?.isHuman === true;

  const suitName = (suit: number) => (SUIT_NAMES[suit] ? t(SUIT_NAMES[suit]) : '');

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.euchre.bg}`} aria-busy={loading}>
      <GamePageHeading title={tc('nav.euchre')} />
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanBidTurn || isHumanTurn || isHumanDiscard}>
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/euchre" />
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
                    value: euchreConfig.cpuDifficulty,
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
                    value: euchreConfig.pointLimit,
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

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              {state.trumpSuit > 0 && <span>{t('trumpSuit', { suit: suitName(state.trumpSuit) })}</span>}
              {state.trumpSuit === 0 && <span>{t('noTrump')}</span>}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Maker / Going alone info */}
                {state.makerTeam >= 0 && (
                  <div className="text-ds-warning text-center mb-2">
                    <span className="mr-4">{t('maker', { team: state.makerTeam })}</span>
                    {state.goingAlone && <span>{t('goingAlone')}</span>}
                  </div>
                )}

                {/* Pick-up phase instruction */}
                {isHumanBidTurn && isPickUpPhase && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="eu-pickup-controls">
                    {t('pickUpPhase')}
                  </div>
                )}

                {/* Call trump phase instruction */}
                {isHumanBidTurn && isCallTrumpPhase && (
                  <div className="text-ds-warning text-center mb-2">{t('callTrumpPhase')}</div>
                )}

                {/* Discard phase instruction */}
                {isHumanDiscard && <div className="text-ds-warning text-center mb-2">{t('discardPhase')}</div>}

                {/* Face-up card */}
                {state.faceUpCard && (isPickUpPhase || isCallTrumpPhase) && (
                  <div className="my-2 text-center">
                    <div className="text-ds-text-muted text-sm mb-1">{t('faceUpCard')}</div>
                    <div className="inline-block">
                      <AnimatedCard
                        card={state.faceUpCard}
                        width={cardWidth}
                        onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                      />
                    </div>
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="eu-trick-display"
                  onCardDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                />

                {/* Partnership info */}
                {humanPlayer && (
                  <div className="text-ds-text-muted text-sm text-center mb-2" data-tutorial="eu-team-info">
                    {t('partnership', {
                      partner: playerName(
                        state.players.find((p) => !p.isHuman && p.team === humanTeam)?.id ?? -1,
                        false,
                      ),
                    })}
                    {state.dealerIdx === humanPlayer.id ? ` | ${t('dealer')}` : ''}
                  </div>
                )}
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
                            {t('team', { n: p.team })} | {t('trickCount', { count: p.trickCount })}
                            {state.dealerIdx === p.id ? ` | ${t('dealer')}` : ''}
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
                          {t('team', { n: p.team })} | {t('trickCount', { count: p.trickCount })}
                          {state.dealerIdx === p.id ? ` | ${t('dealer')}` : ''}
                        </div>
                      </div>
                    ))
                )}

                {/* Team scores */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30"
                    data-tutorial="eu-score-table"
                    open={isRoundEnd || isGameEnd || undefined}
                  >
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">
                      {t('teamScores')}
                    </summary>
                    <table className="w-full text-sm text-ds-text-muted mt-1">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('team', { n: '' })}
                          </th>
                          <th scope="col">{tc('button.score', { defaultValue: 'Score' })}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, idx) => (
                          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                            <td className="text-center">{score}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="eu-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('team', { n: '' })}
                          </th>
                          <th scope="col">{tc('button.score', { defaultValue: 'Score' })}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, idx) => (
                          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                            <td className="text-center">{score}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </div>
                )}
              </div>
            </div>

            {/* Message */}
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Action log */}
            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.euchre.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <div
                className={isMobile ? 'flex gap-1 overflow-x-auto mb-2' : 'flex flex-wrap gap-1 mb-2'}
                data-tutorial="eu-player-hand"
              >
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
                      ...(isMobile ? { minWidth: solitaireMinColWidth, flexShrink: 0 } : {}),
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

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.cardIndex != null
                  ? `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`
                  : `(${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="eu-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Pick-up phase controls */}
              {isHumanBidTurn && isPickUpPhase && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => handleOrderUp(false)} disabled={loading}>
                    {t('orderUpButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => handleOrderUp(true)} disabled={loading}>
                    {t('orderUpAloneButton')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {/* Call trump phase controls */}
              {isHumanBidTurn && isCallTrumpPhase && (
                <>
                  {[1, 2, 3, 4]
                    .filter(
                      (s) =>
                        state.faceUpCard == null ||
                        s !==
                          ({ SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4, JOKER: 0 } as Record<string, number>)[
                            state.faceUpCard.design
                          ],
                    )
                    .map((s) => (
                      <button
                        key={s}
                        type="button"
                        className={btnPrimary}
                        onClick={() => handleCallTrump(s, goAlone)}
                        disabled={loading}
                      >
                        {t(SUIT_NAMES[s])}
                      </button>
                    ))}
                  <label className="text-ds-text-primary text-sm flex items-center gap-1">
                    <input type="checkbox" checked={goAlone} onChange={(e) => setGoAlone(e.target.checked)} />
                    {t('goAloneCheck')}
                  </label>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {/* Discard phase */}
              {isHumanDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('discardButton')}
                </button>
              )}

              {/* Play phase */}
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

              {/* Trick end */}
              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {/* Round end */}
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}

              {/* Reset */}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="eu-reset-button"
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
