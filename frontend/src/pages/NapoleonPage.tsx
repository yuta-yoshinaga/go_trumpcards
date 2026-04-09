import { useCallback, useMemo, useState } from 'react';
import type { napoleonApi } from '../api/gameApi';
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
import { NapoleonSkeleton } from '../components/skeleton/NapoleonSkeleton';
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
  MIN_BID_OPTIONS,
  POINT_LIMIT_OPTIONS,
  useNapoleonGame,
} from '../hooks/useNapoleonGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnOutline, btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NapoleonResponse } from '../types/card';
import { NapoleonPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { valueName } from '../utils/cardUtils';
import { NAPOLEON_HELP, parseNapoleonCommand } from '../utils/cli/commands/napoleonCommands';
import { formatNapoleonState } from '../utils/cli/formatters/napoleonFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Napoleon tutorial step definitions. */
const NP_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="np-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-trump-declaration"]',
    messageKey: 'tutorial.trumpDeclaration',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-adjutant-info"]',
    messageKey: 'tutorial.adjutantInfo',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-kitty-cards"]',
    messageKey: 'tutorial.kittyCards',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="np-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const NAPOLEON_PHASE_KEYS: Readonly<Record<number, string>> = {
  [NapoleonPhase.BID]: 'bid',
  [NapoleonPhase.TRUMP_DECLARATION]: 'trumpDeclaration',
  [NapoleonPhase.KITTY_EXCHANGE]: 'kittyExchange',
  [NapoleonPhase.PLAY]: 'play',
  [NapoleonPhase.TRICK_END]: 'trickEnd',
  [NapoleonPhase.ROUND_END]: 'roundEnd',
  [NapoleonPhase.GAME_END]: 'gameEnd',
};

const SUIT_KEYS: Record<number, string> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };

/** Renders the Napoleon game page with bidding, trump declaration, kitty exchange, trick play, and scoring. */
export function NapoleonPage() {
  return (
    <TutorialWrapper gameName="napoleon" steps={NP_TUTORIAL_STEPS}>
      <NapoleonPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Napoleon page, wrapped by TutorialProvider. */
function NapoleonPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('napoleon');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    retry,
    apiExec,
    napoleonConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleBid,
    handlePass,
    handleTrumpDeclaration,
    handleExchange,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useNapoleonGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('napoleon', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const [bidValue, setBidValue] = useState(12);
  const [trumpSuitValue, setTrumpSuitValue] = useState(1);
  const [adjSuitValue, setAdjSuitValue] = useState(1);
  const [adjValueValue, setAdjValueValue] = useState(1);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('napoleon');
  const cliConfig: CliGameConfig<NapoleonResponse, Parameters<typeof napoleonApi.exec>> = useMemo(
    () => ({
      gameName: 'napoleon',
      parseCommand: parseNapoleonCommand,
      formatResponse: formatNapoleonState,
      helpText: NAPOLEON_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === NapoleonPhase.PLAY;
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

  const phaseNames = usePhaseNames('napoleon', NAPOLEON_PHASE_KEYS);

  if (!state) return <NapoleonSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === NapoleonPhase.BID;
  const isTrumpDeclaration = state.phase === NapoleonPhase.TRUMP_DECLARATION;
  const isKittyExchange = state.phase === NapoleonPhase.KITTY_EXCHANGE;
  const isPlayPhase = state.phase === NapoleonPhase.PLAY;
  const isTrickEnd = state.phase === NapoleonPhase.TRICK_END;
  const isRoundEnd = state.phase === NapoleonPhase.ROUND_END;
  const isGameEnd = state.phase === NapoleonPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanNapoleon = isTrumpDeclaration && state.players[state.napoleonIdx]?.isHuman === true;
  const isHumanExchange = isKittyExchange && state.players[state.napoleonIdx]?.isHuman === true;

  const roleBadge = (p: { isNapoleon: boolean; isAdjutant: boolean; adjutantRevealed: boolean }) => {
    if (p.isNapoleon) return ` [${t('role.napoleon')}]`;
    if (p.isAdjutant && (p.adjutantRevealed || state.adjutantRevealed)) return ` [${t('role.adjutant')}]`;
    return '';
  };

  const trumpLabel = state.trumpSuit > 0 ? t(`suitName.${SUIT_KEYS[state.trumpSuit]}`) : '';

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.napoleon.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.napoleon')} />
      {/* Phase indicator */}
      <PhaseIndicator
        phaseName={phaseNames[state.phase]}
        isHumanTurn={isHumanBidTurn || isHumanTurn || isHumanNapoleon || isHumanExchange}
      >
        <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        <TutorialButton />
        <ManualButton gamePath="/napoleon" />
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
                    value: napoleonConfig.cpuDifficulty,
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
                    value: napoleonConfig.pointLimit,
                    options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('pointLimit', v),
                  },
                  {
                    type: 'select',
                    id: 'minBid',
                    label: t('settings.minBid'),
                    value: napoleonConfig.minBid,
                    options: MIN_BID_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('minBid', v),
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
              {trumpLabel && (
                <span>
                  {t('trumpSuit')}: {trumpLabel}
                </span>
              )}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Adjutant card info */}
                {state.adjutantCard && (
                  <div className="text-white/70 text-center text-sm mb-2" data-tutorial="np-adjutant-info">
                    {t('adjutantCard')}:{' '}
                    <AnimatedCard
                      card={state.adjutantCard}
                      width={cardWidth * 0.6}
                      onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                    />
                  </div>
                )}

                {/* Highest bid info */}
                {state.highestBid > 0 && (
                  <div className="text-white/70 text-center text-sm mb-2">
                    {t('highestBid', { bid: state.highestBid })}
                  </div>
                )}

                {/* Bid phase instruction */}
                {isHumanBidTurn && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="np-bid-controls">
                    {t('bidPhase', { min: napoleonConfig.minBid })}
                  </div>
                )}

                {/* Trump declaration instruction */}
                {isHumanNapoleon && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="np-trump-declaration">
                    {t('trumpDeclarationPhase')}
                  </div>
                )}

                {/* Kitty exchange instruction */}
                {isHumanExchange && <div className="text-ds-warning text-center mb-2">{t('kittyExchangePhase')}</div>}

                {/* Kitty cards (during exchange phase) */}
                {isKittyExchange && state.kitty.length > 0 && (
                  <div className="my-2 p-2 rounded bg-black/40" data-tutorial="np-kitty-cards">
                    <div className="text-white/70 text-sm mb-1">{t('kittyLabel')}</div>
                    <div className="flex gap-2">
                      {state.kitty.map((card, idx) => (
                        <AnimatedCard
                          key={`kitty-${card.design}-${card.value}-${idx}`}
                          card={card}
                          width={cardWidth}
                          onDealComplete={() => playSound('cardDeal', { pitchVariation: 0.03 })}
                        />
                      ))}
                    </div>
                  </div>
                )}

                {/* Current trick */}
                {state.currentTrick.length > 0 && (
                  <div className="my-3 p-3 rounded bg-black/40" data-tutorial="np-trick-display">
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
                        {playerName(p.id, p.isHuman)}
                        {roleBadge(p)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('cumulativeScore', { score: p.cumulativeScore })} |{' '}
                        {t('roundScore', { score: p.roundScore })} |{' '}
                        {p.bid >= 0 ? t('bid', { n: p.bid }) : t('bidNone')} |{' '}
                        {t('pictureCards', { count: p.pictureCards })}
                      </div>
                    </div>
                  ))}

                {/* Score table */}
                <div className="my-3 p-2 rounded bg-black/30 relative" data-tutorial="np-score-table">
                  <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
                  <div className="overflow-x-auto -mx-2 px-2">
                    <table className="w-full text-sm text-white/70 min-w-[420px]">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('scoresPlayer')}
                          </th>
                          <th scope="col">{t('scoresRole')}</th>
                          <th scope="col">{t('scoresBid')}</th>
                          <th scope="col">{t('scoresPictureCards')}</th>
                          <th scope="col">{t('scoresTricks')}</th>
                          <th scope="col">{t('scoresRound')}</th>
                          <th scope="col">{t('scoresTotal')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.players.map((p) => (
                          <tr key={p.id} className={p.isHuman ? 'text-ds-accent' : ''}>
                            <td>{playerName(p.id, p.isHuman)}</td>
                            <td className="text-center">
                              {p.isNapoleon
                                ? t('role.napoleon')
                                : p.isAdjutant && (p.adjutantRevealed || state.adjutantRevealed)
                                  ? t('role.adjutant')
                                  : '-'}
                            </td>
                            <td className="text-center">{p.bid >= 0 ? p.bid : '-'}</td>
                            <td className="text-center">{p.pictureCards}</td>
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
          <GameFooter className={`${gameTheme.napoleon.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer &&
              (isMobile ? (
                <MobileHandGrid
                  cards={humanPlayer.cards}
                  selectedIndices={selectedCardIndices}
                  onToggle={toggleCard}
                  cardWidth={cardWidth}
                  dataTutorial="np-player-hand"
                />
              ) : (
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="np-player-hand">
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
                  : hint.trumpSuit != null
                    ? `${t('hintTrump')}: ${t(`suitName.${SUIT_KEYS[hint.trumpSuit]}`)} (${t(`hintReason.${hint.reason}`)})`
                    : hint.discardIndex != null
                      ? `${t('hintDiscard')}: [${hint.discardIndex}] (${t(`hintReason.${hint.reason}`)})`
                      : `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap">
              {(isHumanBidTurn || isHumanTurn || isHumanNapoleon || isHumanExchange) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Bid controls */}
              {isHumanBidTurn && (
                <>
                  <input
                    type="number"
                    min={napoleonConfig.minBid}
                    max={17}
                    value={bidValue}
                    onChange={(e) => setBidValue(Number(e.target.value))}
                    className="w-16 px-2 py-1 rounded bg-white/20 text-white text-center"
                    aria-label="bid-input"
                  />
                  <button type="button" className={btnPrimary} onClick={() => handleBid(bidValue)} disabled={loading}>
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {/* Trump declaration controls */}
              {isHumanNapoleon && (
                <>
                  <select
                    value={trumpSuitValue}
                    onChange={(e) => setTrumpSuitValue(Number(e.target.value))}
                    className="px-2 py-1 rounded bg-white/20 text-white"
                    aria-label="trump-suit"
                  >
                    {[1, 2, 3, 4].map((s) => (
                      <option key={s} value={s}>
                        {t(`suitName.${SUIT_KEYS[s]}`)}
                      </option>
                    ))}
                  </select>
                  <select
                    value={adjSuitValue}
                    onChange={(e) => setAdjSuitValue(Number(e.target.value))}
                    className="px-2 py-1 rounded bg-white/20 text-white"
                    aria-label="adjutant-suit"
                  >
                    <option value={0}>JOKER</option>
                    {[1, 2, 3, 4].map((s) => (
                      <option key={s} value={s}>
                        {t(`suitName.${SUIT_KEYS[s]}`)}
                      </option>
                    ))}
                  </select>
                  {adjSuitValue > 0 && (
                    <select
                      value={adjValueValue}
                      onChange={(e) => setAdjValueValue(Number(e.target.value))}
                      className="px-2 py-1 rounded bg-white/20 text-white"
                      aria-label="adjutant-value"
                    >
                      {Array.from({ length: 13 }, (_, i) => i + 1).map((v) => (
                        <option key={v} value={v}>
                          {valueName(v)}
                        </option>
                      ))}
                    </select>
                  )}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() =>
                      handleTrumpDeclaration(trumpSuitValue, adjSuitValue, adjSuitValue === 0 ? 0 : adjValueValue)
                    }
                    disabled={loading}
                  >
                    {t('declareButton')}
                  </button>
                </>
              )}

              {/* Kitty exchange controls */}
              {isHumanExchange && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => {
                    if (selectedCardIndices.length === 1) handleExchange(selectedCardIndices[0]);
                  }}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('exchangeButton')}
                </button>
              )}

              {/* Play controls */}
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
              <button
                type="button"
                className={btnOutline}
                data-tutorial="np-reset-button"
                onClick={() =>
                  requestConfirm(() => {
                    hideActionLog();
                    return apiExec('reset', undefined, undefined, undefined, undefined, undefined, undefined, {
                      cpuDifficulty: napoleonConfig.cpuDifficulty,
                      pointLimit: napoleonConfig.pointLimit,
                      minBid: napoleonConfig.minBid,
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
