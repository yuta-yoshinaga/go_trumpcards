import { useCallback, useMemo, useState } from 'react';
import type { bridgeApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { MobileHandGrid } from '../components/MobileHandGrid';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, useBridgeGame } from '../hooks/useBridgeGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BridgeResponse } from '../types/card';
import { BridgePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { buildBridgeAuctionGrid, finalContractBid } from '../utils/bridgeAuction';
import { canBid, canDouble, canRedouble } from '../utils/bridgeBidRules';
import { cardAlt } from '../utils/cardAlt';
import { BRIDGE_HELP, parseBridgeCommand } from '../utils/cli/commands/bridgeCommands';
import { formatBridgeState } from '../utils/cli/formatters/bridgeFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Denomination names used in bid controls. */
const DENOMINATIONS: readonly { suit: number; labelKey: string }[] = [
  { suit: 1, labelKey: 'denominationClub' },
  { suit: 2, labelKey: 'denominationDiamond' },
  { suit: 3, labelKey: 'denominationHeart' },
  { suit: 4, labelKey: 'denominationSpade' },
  { suit: 5, labelKey: 'denominationNoTrump' },
] as const;

/** Suit display name map for trump/contract display. */
const SUIT_DISPLAY: Readonly<Record<number, string>> = {
  1: 'suitClover',
  2: 'suitDiamond',
  3: 'suitHeart',
  4: 'suitSpade',
  5: 'suitNoTrump',
};

/** Bridge tutorial step definitions. */
const BR_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="br-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-bid-history"]',
    messageKey: 'tutorial.bidHistory',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-dummy-hand"]',
    messageKey: 'tutorial.dummyHand',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-team-scores"]',
    messageKey: 'tutorial.teamScores',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="br-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BRIDGE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BridgePhase.BID]: 'bid',
  [BridgePhase.PLAY]: 'play',
  [BridgePhase.TRICK_END]: 'trickEnd',
  [BridgePhase.ROUND_END]: 'roundEnd',
  [BridgePhase.GAME_END]: 'gameEnd',
};

/** Renders the Bridge game page with auction, trick play, and team scoring. */
export const BridgePage = withTutorial(BridgePageContent, 'bridge', BR_TUTORIAL_STEPS);
/** Inner content of the Bridge page, wrapped by TutorialProvider. */
function BridgePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bridge');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    retry,
    apiExec,
    bridgeConfig,
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
  } = useBridgeGame();
  const { cardWidth, isMobile } = useCardDimensions();
  const [bidLevel, setBidLevel] = useState(1);
  const [bidSuit, setBidSuit] = useState(5);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bridge');
  const cliConfig: CliGameConfig<BridgeResponse, Parameters<typeof bridgeApi.exec>> = useMemo(
    () => ({
      gameName: 'bridge',
      parseCommand: parseBridgeCommand,
      formatResponse: formatBridgeState,
      helpText: BRIDGE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === BridgePhase.PLAY;
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

  const phaseNames = usePhaseNames('bridge', BRIDGE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiExec('reset', undefined, undefined, undefined, undefined, {
      cpuDifficulty: bridgeConfig.cpuDifficulty,
    });
  }, [apiExec, hideActionLog, bridgeConfig.cpuDifficulty]);

  if (!state)
    return <GameSkeleton gameKey="bridge" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isBidPhase = state.phase === BridgePhase.BID;
  const isPlayPhase = state.phase === BridgePhase.PLAY;
  const isTrickEnd = state.phase === BridgePhase.TRICK_END;
  const isRoundEnd = state.phase === BridgePhase.ROUND_END;
  const isGameEnd = state.phase === BridgePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = isBidPhase && state.players[state.bidPlayerIdx]?.isHuman === true;

  // Auction legality: mirror internal/domain/Bridge.go so illegal controls are
  // disabled before a request is sent.
  const doubleLegal = canDouble(state);
  const redoubleLegal = canRedouble(state);
  const bidLegal = canBid(state.contractLevel, state.contractSuit, bidLevel, bidSuit);

  const suitLabel = (suit: number) => (SUIT_DISPLAY[suit] ? t(SUIT_DISPLAY[suit]) : '');

  // Render one auction cell: pass/double/redouble labels, or level+strain.
  const bidCellLabel = (entry: { bidType: number; level: number; suit: number } | null) => {
    if (!entry) return '';
    if (entry.bidType === 0) return t('passButton');
    if (entry.bidType === 2) return t('doubleButton');
    if (entry.bidType === 3) return t('redoubleButton');
    return `${entry.level}${suitLabel(entry.suit)}`;
  };

  const auctionGrid = buildBridgeAuctionGrid(state.bidHistory);
  const winningBid = state.contractLevel > 0 ? finalContractBid(state.bidHistory) : null;

  const contractDisplay = () => {
    if (state.contractLevel <= 0) return null;
    const suit = suitLabel(state.contractSuit);
    if (state.doubled === 2) return t('contractRedoubled', { level: state.contractLevel, suit });
    if (state.doubled === 1) return t('contractDoubled', { level: state.contractLevel, suit });
    return t('contract', { level: state.contractLevel, suit });
  };

  return (
    <GamePageShell
      title={tc('nav.bridge')}
      gameThemeBg={gameTheme.bridge.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn}
      gamePath="/bridge"
      gameEndFlag={isGameEnd}
      onCelebrate={() => playSound('winFanfare')}
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
                    value: bridgeConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
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
              {isPlayPhase && <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>}
              {contractDisplay() && <span className="mr-4">{contractDisplay()}</span>}
              {state.trumpSuit !== 0 && state.contractLevel > 0 && (
                <span>
                  {state.trumpSuit === 5 ? t('noTrump') : t('trumpSuit', { suit: suitLabel(state.trumpSuit) })}
                </span>
              )}
            </div>

            {/* Vulnerability */}
            <div className="text-ds-text-muted text-center text-sm mb-2">
              <span className="mr-4">
                {t('team', { n: 0 })}: {state.vulnerability[0] ? t('vulnerable') : t('notVulnerable')}
              </span>
              <span>
                {t('team', { n: 1 })}: {state.vulnerability[1] ? t('vulnerable') : t('notVulnerable')}
              </span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Declarer/Dummy info */}
                {state.declarerIdx >= 0 && (
                  <div className="text-ds-warning text-center mb-2">
                    <span className="mr-4">
                      {t('declarer')}:{' '}
                      {playerName(
                        state.players[state.declarerIdx]?.id ?? -1,
                        state.players[state.declarerIdx]?.isHuman ?? false,
                      )}
                    </span>
                    {state.dummyIdx >= 0 && (
                      <span>
                        {t('dummy')}:{' '}
                        {playerName(
                          state.players[state.dummyIdx]?.id ?? -1,
                          state.players[state.dummyIdx]?.isHuman ?? false,
                        )}
                      </span>
                    )}
                  </div>
                )}

                {/* Bid phase instruction */}
                {isHumanBidTurn && <div className="text-ds-warning text-center mb-2">{t('bidPhase')}</div>}

                {/* Bid History — traditional 4-column auction grid */}
                {auctionGrid && (
                  <div className="my-2 p-2 rounded bg-black/30" data-tutorial="br-bid-history">
                    <div className="text-ds-text-muted text-sm mb-1">{t('bidHistory')}</div>
                    <div className="overflow-x-auto">
                      <table className="w-full text-xs text-ds-text-primary" data-testid="bridge-auction-table">
                        <caption className="sr-only">{t('auctionTableCaption')}</caption>
                        <thead>
                          <tr>
                            {auctionGrid.columns.map((seat) => {
                              const seatPlayer = state.players[seat];
                              const isHumanCol = seatPlayer?.isHuman === true;
                              return (
                                <th
                                  key={seat}
                                  scope="col"
                                  className={`px-2 py-0.5 text-center font-semibold ${
                                    isHumanCol ? 'text-ds-accent' : 'text-ds-text-muted'
                                  }`}
                                >
                                  {playerName(seatPlayer?.id ?? seat, seatPlayer?.isHuman ?? false)}
                                </th>
                              );
                            })}
                          </tr>
                        </thead>
                        <tbody>
                          {auctionGrid.rows.map((row, rIdx) => (
                            <tr key={rIdx}>
                              {row.map((cell, cIdx) => {
                                const isWinning = cell !== null && cell === winningBid;
                                return (
                                  <td
                                    key={cIdx}
                                    className={`px-2 py-0.5 text-center ${
                                      isWinning ? 'rounded bg-ds-accent/30 font-bold text-ds-accent' : ''
                                    }`}
                                    data-winning-contract={isWinning ? 'true' : undefined}
                                  >
                                    {bidCellLabel(cell)}
                                  </td>
                                );
                              })}
                            </tr>
                          ))}
                        </tbody>
                      </table>
                    </div>
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="br-trick-display"
                />

                {/* Dummy hand */}
                {state.openingLeadDone && state.dummyHand && state.dummyHand.length > 0 && (
                  <div
                    className={`relative my-3 overflow-hidden rounded p-3 bg-black/40 transition-shadow ${
                      humanPlayer?.id === state.declarerIdx && state.currentPlayerIdx === state.dummyIdx
                        ? 'ring-2 ring-ds-info shadow-[0_0_18px_rgba(91,143,185,0.55)] motion-safe:animate-pulse'
                        : ''
                    }`}
                    data-tutorial="br-dummy-hand"
                    data-testid="bridge-dummy-area"
                    data-dummy-active={
                      humanPlayer?.id === state.declarerIdx && state.currentPlayerIdx === state.dummyIdx
                        ? 'true'
                        : undefined
                    }
                  >
                    <span
                      aria-hidden="true"
                      className="pointer-events-none absolute inset-0 flex items-center justify-center select-none text-5xl font-extrabold tracking-widest text-ds-text-muted opacity-10"
                    >
                      DUMMY
                    </span>
                    <div className="text-ds-text-muted text-sm mb-1">{t('dummyHand')}</div>
                    <div className="flex gap-1 flex-wrap">
                      {state.dummyHand.map((card, idx) => (
                        <AnimatedCard key={`dummy-${card.design}-${card.value}-${idx}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                    {humanPlayer?.id === state.declarerIdx && state.currentPlayerIdx === state.dummyIdx && (
                      <div className="relative mt-1 text-center text-xs font-semibold text-ds-info">
                        {t('yourTurnDummy')}
                      </div>
                    )}
                  </div>
                )}

                {/* Partnership info */}
                {humanPlayer && (
                  <div className="text-ds-text-muted text-sm text-center mb-2">
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
                            {state.declarerIdx === p.id ? ` | ${t('declarer')}` : ''}
                            {state.dummyIdx === p.id ? ` | ${t('dummy')}` : ''}
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
                          {state.declarerIdx === p.id ? ` | ${t('declarer')}` : ''}
                          {state.dummyIdx === p.id ? ` | ${t('dummy')}` : ''}
                        </div>
                      </div>
                    ))
                )}

                {/* Team scores */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30"
                    data-tutorial="br-team-scores"
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
                          <th scope="col">{t('gamesWon')}</th>
                          <th scope="col">{t('belowLine')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, idx) => (
                          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                            <td className="text-center">{score}</td>
                            <td className="text-center">{state.gamesWon[idx]}</td>
                            <td className="text-center">{state.belowLine[idx]}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="br-team-scores">
                    <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('team', { n: '' })}
                          </th>
                          <th scope="col">{tc('button.score', { defaultValue: 'Score' })}</th>
                          <th scope="col">{t('gamesWon')}</th>
                          <th scope="col">{t('belowLine')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, idx) => (
                          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                            <td className="text-center">{score}</td>
                            <td className="text-center">{state.gamesWon[idx]}</td>
                            <td className="text-center">{state.belowLine[idx]}</td>
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
          <GameFooter className={`${gameTheme.bridge.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer &&
              (isMobile ? (
                <MobileHandGrid
                  cards={humanPlayer.cards}
                  selectedIndices={selectedCardIndices}
                  onToggle={toggleCard}
                  cardWidth={cardWidth}
                  dataTutorial="br-player-hand"
                />
              ) : (
                <div className="flex flex-wrap gap-1 mb-2" data-tutorial="br-player-hand">
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
                      <AnimatedCard card={card} width={cardWidth} />
                    </button>
                  ))}
                </div>
              ))}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.cardIndex != null
                  ? `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`
                  : `(${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}

            <div className="flex gap-1 items-center flex-wrap" data-tutorial="br-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Bid phase controls */}
              {isHumanBidTurn && (
                <span data-tutorial="br-bid-controls" className="flex gap-1 items-center flex-wrap">
                  <select
                    className="text-xs rounded bg-black/50 text-ds-text-primary px-1.5 py-0.5"
                    value={bidLevel}
                    onChange={(e) => setBidLevel(Number(e.target.value))}
                    aria-label={t('bidLevel')}
                  >
                    {[1, 2, 3, 4, 5, 6, 7].map((lv) => (
                      <option key={lv} value={lv}>
                        {lv}
                      </option>
                    ))}
                  </select>
                  <select
                    className="text-xs rounded bg-black/50 text-ds-text-primary px-1.5 py-0.5"
                    value={bidSuit}
                    onChange={(e) => setBidSuit(Number(e.target.value))}
                    aria-label={t('bidSuit')}
                  >
                    {DENOMINATIONS.map((d) => (
                      <option key={d.suit} value={d.suit}>
                        {t(d.labelKey)}
                      </option>
                    ))}
                  </select>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleBid(1, bidLevel, bidSuit)}
                    disabled={loading || !bidLegal}
                    title={!bidLegal ? t('bidTooLow') : undefined}
                    data-testid="br-bid-submit"
                  >
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={() => handleBid(0)} disabled={loading}>
                    {t('passButton')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleBid(2)}
                    disabled={loading || !doubleLegal}
                    title={!doubleLegal ? t('doubleIllegal') : undefined}
                    data-testid="br-double"
                  >
                    {t('doubleButton')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleBid(3)}
                    disabled={loading || !redoubleLegal}
                    title={!redoubleLegal ? t('redoubleIllegal') : undefined}
                    data-testid="br-redouble"
                  >
                    {t('redoubleButton')}
                  </button>
                </span>
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
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="br-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
