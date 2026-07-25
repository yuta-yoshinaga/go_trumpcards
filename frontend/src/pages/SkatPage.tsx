import { useCallback, useMemo, useState } from 'react';
import type { skatApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useSkatGame } from '../hooks/useSkatGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SkatResponse } from '../types/card';
import { SkatGameType, SkatPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSkatCommand, SKAT_HELP } from '../utils/cli/commands/skatCommands';
import { formatSkatState } from '../utils/cli/formatters/skatFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';
import { skatBestBidEstimate } from '../utils/skatBidEstimate';

/** Suit identifiers matching internal/domain/Card.go (1=Spade, 2=Clover, 3=Heart, 4=Diamond). */
const SUIT_SPADE = 1;
const SUIT_CLOVER = 2;
const SUIT_HEART = 3;
const SUIT_DIAMOND = 4;

/** Skat tutorial step definitions. */
const SKAT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sk-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sk-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sk-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Phase translation key map for Skat. */
const SKAT_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SkatPhase.BID]: 'skat.phase.bid',
  [SkatPhase.SKAT_PICKUP]: 'skat.phase.skatPickup',
  [SkatPhase.DISCARD]: 'skat.phase.discard',
  [SkatPhase.GAME_DECLARATION]: 'skat.phase.gameDeclaration',
  [SkatPhase.PLAY]: 'phase.play',
  [SkatPhase.TRICK_END]: 'phase.trickEnd',
  [SkatPhase.ROUND_END]: 'phase.roundEnd',
  [SkatPhase.GAME_END]: 'phase.gameEnd',
};

/** Renders the Skat (German trick-taking) game page. */
export const SkatPage = withTutorial(SkatPageContent, 'skat', SKAT_TUTORIAL_STEPS);
/** Inner content of the Skat page. */
function SkatPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('skat');
  const skatGame = useSkatGame();
  const {
    state,
    loading,
    error,
    skatConfig,
    handleConfigChange,
    reset,
    selectedCardIndices,
    toggleCard,
    handleBid,
    handlePickSkat,
    handleDiscard,
    handleDeclareGame,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    retry,
  } = skatGame;
  const { cardWidth, isMobile: _isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('skat', SKAT_PHASE_KEYS);
  const [trumpSuit, setTrumpSuit] = useState<number>(SUIT_SPADE);
  const { hint, hintEnabled, setHintEnabled } = useGameHint('skat', state);
  const cliMode = useCliMode('skat');
  const cliConfig: CliGameConfig<SkatResponse, Parameters<typeof skatApi.exec>> = useMemo(
    () => ({
      gameName: 'skat',
      parseCommand: parseSkatCommand,
      formatResponse: formatSkatState,
      helpText: SKAT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(skatGame.dispatch, cliConfig, state, {
    addInput: cliMode.addInput,
    addOutput: cliMode.addOutput,
    addError: cliMode.addError,
    clearLog: cliMode.clearLog,
  });

  const handleManualReset = useCallback(() => {
    hideActionLog();
    reset();
  }, [hideActionLog, reset]);

  if (!state) {
    return (
      <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.skat.bg}`}>
        <div className="flex-1 flex items-center justify-center text-ds-text-primary">
          <p>{tc('skeleton.loading')}</p>
        </div>
      </div>
    );
  }

  const isBid = state.phase === SkatPhase.BID;
  const isPickup = state.phase === SkatPhase.SKAT_PICKUP;
  const isDiscard = state.phase === SkatPhase.DISCARD;
  const isDeclareGame = state.phase === SkatPhase.GAME_DECLARATION;
  const isPlay = state.phase === SkatPhase.PLAY;
  const isTrickEnd = state.phase === SkatPhase.TRICK_END;
  const isRoundEnd = state.phase === SkatPhase.ROUND_END;
  const isGameEnd = state.phase === SkatPhase.GAME_END || state.gameEndFlag;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = isPlay && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn =
    isBid && state.activeBidActorIdx >= 0 && state.players[state.activeBidActorIdx]?.isHuman === true;
  const isHumanDeclarer = state.declarerIdx >= 0 && state.players[state.declarerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.skat')}
      gameThemeBg={gameTheme.skat.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isHumanBidTurn || (isHumanDeclarer && (isPickup || isDiscard || isDeclareGame))}
      gamePath="/skat"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliMode.cliEnabled} onToggle={cliMode.toggleCli} />}
    >
      {cliMode.cliEnabled ? (
        <CliTerminal logEntries={cliMode.logEntries} onCommand={handleCommand} disabled={loading} />
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
                    value: skatConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetScore',
                    label: t('settings.targetScore'),
                    value: skatConfig.targetScore,
                    options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetScore', v),
                  },
                  hintCheckboxItem(tc, hintEnabled, setHintEnabled),
                ],
              },
            ]}
          />
          <div className="flex-1 overflow-y-auto px-4 py-2 space-y-4">
            {error && <ErrorAlert message={error} onRetry={retry} />}

            {/* Round / declarer / game info */}
            <div className="bg-black/30 text-ds-text-primary p-3 rounded space-y-1 text-sm">
              <div>
                {t('round')}: {state.roundNumber} | {t('dealer')}: CPU {state.dealerIdx} | {t('currentBid')}:{' '}
                {state.currentBid}
              </div>
              {isBid &&
                humanPlayer &&
                (() => {
                  const est = skatBestBidEstimate(humanPlayer.cards ?? []);
                  const exceeds = state.currentBid > est.value;
                  return (
                    <div
                      data-testid="bid-estimate"
                      data-exceeds={exceeds ? 'true' : undefined}
                      className={`text-xs ${exceeds ? 'text-ds-error' : 'text-ds-text-muted'}`}
                    >
                      {t('bidEstimate', { value: est.value, type: t(`gameTypeLabel.${est.gameType.toLowerCase()}`) })}
                      {exceeds && <span className="ml-2">⚠️ {t('bidExceedsHand')}</span>}
                    </div>
                  );
                })()}
              {state.declarerIdx >= 0 && (
                <div data-tutorial="sk-declarer-info">
                  {t('declarer')}: {state.players[state.declarerIdx]?.isHuman ? t('you') : `CPU ${state.declarerIdx}`}
                  {state.gameType !== SkatGameType.NONE && (
                    <span className="ml-2">
                      | {t('gameType')}:{' '}
                      {state.gameType === SkatGameType.SUIT
                        ? `${t('suitGame')} (${suitLabel(state.trumpSuit)})`
                        : state.gameType === SkatGameType.GRAND
                          ? t('grandGame')
                          : t('nullGame')}
                    </span>
                  )}
                </div>
              )}
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* Trick display */}
            <TrickDisplay
              currentTrick={state.currentTrick.map((tc) => ({ playerIdx: tc.playerIdx, card: tc.card }))}
              players={state.players.map((p) => ({ id: p.id, isHuman: p.isHuman }))}
              cardWidth={cardWidth}
              label={t('currentTrick')}
              dataTutorial="sk-trick-display"
            />

            {/* Skat (face-up at round end) */}
            {state.originalSkat && state.originalSkat.length > 0 && (
              <div className="bg-black/30 text-ds-text-primary p-3 rounded">
                <div className="text-sm mb-1">{t('skatLabel')}:</div>
                <div className="flex gap-2" data-testid="skat-reveal">
                  {state.originalSkat.map((c, i) => (
                    <AnimatedCard key={`skat-${i}`} card={c} width={cardWidth} dealDelay={i * 0.15} />
                  ))}
                </div>
              </div>
            )}

            {/* Player hand */}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={_isMobile}
                dataTutorialPrefix="sk"
              />
            )}

            {/* Player scores */}
            <div className="bg-black/30 text-ds-text-primary p-3 rounded text-sm">
              <table className="w-full">
                <thead>
                  <tr>
                    <th className="text-left">{t('player')}</th>
                    <th className="text-right">{t('tricks')}</th>
                    <th className="text-right">{t('cardPoints')}</th>
                    <th className="text-right">{t('total')}</th>
                  </tr>
                </thead>
                <tbody>
                  {state.players.map((p) => (
                    <tr key={p.id}>
                      <td>
                        {p.isHuman ? t('you') : `CPU ${p.id}`}
                        {p.isDeclarer && ` (${t('declarer')})`}
                      </td>
                      <td className="text-right">{p.trickCount}</td>
                      <td className="text-right">{p.cardPoints}</td>
                      <td className="text-right">{p.cumulativeScore}</td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>

            <ActionLogSection
              isEndPhase={isRoundEnd || isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.skat.footer} px-4 py-2.5`}>
            <div className="flex flex-wrap gap-2 items-center">
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="sk-reset-button"
              />

              {/* Bid phase actions */}
              {isHumanBidTurn && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => handleBid(true)} disabled={loading}>
                    {state.currentBid === 0 ? t('callBid') : `${t('acceptAt')} ${state.currentBid}`}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => handleBid(false)} disabled={loading}>
                    {t('pass')}
                  </button>
                </>
              )}

              {/* Skat pickup */}
              {isPickup && isHumanDeclarer && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => handlePickSkat(true)} disabled={loading}>
                    {t('pickUp')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => handlePickSkat(false)} disabled={loading}>
                    {t('handGame')}
                  </button>
                </>
              )}

              {/* Discard */}
              {isDiscard && isHumanDeclarer && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 2}
                >
                  {t('discardSelected')}
                </button>
              )}

              {/* Game declaration */}
              {isDeclareGame && isHumanDeclarer && (
                <>
                  <select
                    aria-label={t('trumpSuitLabel')}
                    value={trumpSuit}
                    onChange={(e) => setTrumpSuit(Number(e.target.value))}
                    className="px-2 py-1 rounded text-black"
                    disabled={loading}
                  >
                    <option value={SUIT_SPADE}>{t('spades')} ♠</option>
                    <option value={SUIT_CLOVER}>{t('clubs')} ♣</option>
                    <option value={SUIT_HEART}>{t('hearts')} ♥</option>
                    <option value={SUIT_DIAMOND}>{t('diamonds')} ♦</option>
                  </select>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleDeclareGame(SkatGameType.SUIT, trumpSuit)}
                    disabled={loading}
                  >
                    {t('declareSuit')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleDeclareGame(SkatGameType.GRAND)}
                    disabled={loading}
                  >
                    {t('declareGrand')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleDeclareGame(SkatGameType.NULL)}
                    disabled={loading}
                  >
                    {t('declareNull')}
                  </button>
                </>
              )}

              {/* Play phase */}
              {isPlay && isHumanTurn && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handlePlay}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('play')}
                </button>
              )}

              {isTrickEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
                  {t('nextTrick')}
                </button>
              )}

              {isRoundEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
                  {t('nextRound')}
                </button>
              )}
            </div>
          </GameFooter>
          {hintEnabled && hint && <HintTooltip reason={t(hint.reason)} confidence={hint.confidence} />}
        </>
      )}
    </GamePageShell>
  );
}

/** Returns a translated suit symbol. */
function suitLabel(suit: number): string {
  switch (suit) {
    case SUIT_SPADE:
      return '♠';
    case SUIT_CLOVER:
      return '♣';
    case SUIT_HEART:
      return '♥';
    case SUIT_DIAMOND:
      return '♦';
    default:
      return '?';
  }
}
