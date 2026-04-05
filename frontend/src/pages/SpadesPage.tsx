import { useCallback, useMemo, useState } from 'react';
import type { spadesApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { SpadesSkeleton } from '../components/skeleton/SpadesSkeleton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useSpadesGame } from '../hooks/useSpadesGame';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SpadesResponse } from '../types/card';
import { SpadesPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSpadesCommand, SPADES_HELP } from '../utils/cli/commands/spadesCommands';
import { formatSpadesState } from '../utils/cli/formatters/spadesFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Spades tutorial step definitions. */
const SP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sp-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-bags-info"]',
    messageKey: 'tutorial.bagsInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sp-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SPADES_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SpadesPhase.BID]: 'bid',
  [SpadesPhase.PLAY]: 'play',
  [SpadesPhase.TRICK_END]: 'trickEnd',
  [SpadesPhase.ROUND_END]: 'roundEnd',
  [SpadesPhase.GAME_END]: 'gameEnd',
};

/** Renders the Spades game page with bidding, trick play, and scoring. */
export function SpadesPage() {
  return (
    <TutorialWrapper gameName="spades" steps={SP_TUTORIAL_STEPS}>
      <SpadesPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Spades page, wrapped by TutorialProvider. */
function SpadesPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('spades');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    exec,
    retry,
    spadesConfig,
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
  } = useSpadesGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('spades', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidValue, setBidValue] = useState(1);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('spades');
  const cliConfig: CliGameConfig<SpadesResponse, Parameters<typeof spadesApi.exec>> = useMemo(
    () => ({
      gameName: 'spades',
      parseCommand: parseSpadesCommand,
      formatResponse: formatSpadesState,
      helpText: SPADES_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === SpadesPhase.PLAY;
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

  const phaseNames = usePhaseNames('spades', SPADES_PHASE_KEYS);

  if (!state) return <SpadesSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === SpadesPhase.BID;
  const isPlayPhase = state.phase === SpadesPhase.PLAY;
  const isTrickEnd = state.phase === SpadesPhase.TRICK_END;
  const isRoundEnd = state.phase === SpadesPhase.ROUND_END;
  const isGameEnd = state.phase === SpadesPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.spades')}
      gameThemeBg={gameTheme.spades.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/spades"
      gameEndFlag={!!state?.gameEndFlag}
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
                    value: spadesConfig.cpuDifficulty,
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
                    value: spadesConfig.pointLimit,
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
            <div className="text-white text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span>{state.spadesBroken ? t('spadesBroken') : t('spadesNotBroken')}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Bid phase instruction */}
                {isHumanBidTurn && (
                  <div className="text-yellow-300 text-center mb-2" data-tutorial="sp-bid-controls">
                    {t('bidPhase')}
                  </div>
                )}

                {/* Current trick */}
                {state.currentTrick.length > 0 && (
                  <div className="my-3 p-3 rounded bg-black/40" data-tutorial="sp-trick-display">
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
                        {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} | {t('bags', { count: p.bags })}
                      </div>
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="sp-score-table">
                  <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
                  <div className="overflow-x-auto">
                    <table className="w-full text-sm text-white/70 min-w-[360px]">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresBid')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                          <th scope="col">{t('scoresBags')}</th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => (
                          <tr key={p.id} className={p.isHuman ? 'text-yellow-300' : ''}>
                            <td>{playerName(p.id, p.isHuman)}</td>
                            <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                            <td className="text-center">{p.trickCount}</td>
                            <td className="text-center">{p.bags}</td>
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
            <div data-tutorial="sp-bags-info">
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
          <GameFooter className={`${gameTheme.spades.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="sp"
              />
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-yellow-300 text-sm mb-2">
                {hint.bid != null
                  ? `${t('hintBid')}: ${hint.bid} (${t(`hintReason.${hint.reason}`)})`
                  : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center" data-tutorial="sp-play-button">
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
                    max={13}
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
                data-tutorial="sp-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    return exec('reset', undefined, undefined, {
                      cpuDifficulty: spadesConfig.cpuDifficulty,
                      pointLimit: spadesConfig.pointLimit,
                      nilBonus: spadesConfig.nilBonus,
                      bagPenaltyThreshold: spadesConfig.bagPenaltyThreshold,
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
    </GamePageShell>
  );
}
