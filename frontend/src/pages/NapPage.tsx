import { useEffect, useMemo } from 'react';
import type { napApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useNapGame } from '../hooks/useNapGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { NapResponse } from '../types/card';
import { NapContract, NapPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { NAP_HELP, parseNapCommand } from '../utils/cli/commands/napCommands';
import { formatNapState } from '../utils/cli/formatters/napFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { napPayout } from '../utils/napPayout';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = no trump). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Total tricks in a Nap round (5 cards dealt to each player). */
const NAP_TOTAL_TRICKS = 5;

/** Maps a Nap contract/bid value (0/2/3/4/5) to its i18n key suffix. */
const CONTRACT_KEYS: Readonly<Record<number, string>> = {
  [NapContract.PASS]: 'pass',
  [NapContract.TWO]: 'two',
  [NapContract.THREE]: 'three',
  [NapContract.FOUR]: 'four',
  [NapContract.NAP]: 'nap',
};

/** Bid button options (Pass/Two/Three/Four/Nap). */
const BIDS: { value: number; key: string }[] = [
  { value: NapContract.PASS, key: 'bid.pass' },
  { value: NapContract.TWO, key: 'bid.two' },
  { value: NapContract.THREE, key: 'bid.three' },
  { value: NapContract.FOUR, key: 'bid.four' },
  { value: NapContract.NAP, key: 'bid.nap' },
];

/** Nap tutorial step definitions. */
const NAP_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="nap-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="nap-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nap-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="nap-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const NAP_PHASE_KEYS: Readonly<Record<number, string>> = {
  [NapPhase.BID]: 'bid',
  [NapPhase.PLAY]: 'play',
  [NapPhase.TRICK_END]: 'trickEnd',
  [NapPhase.ROUND_END]: 'roundEnd',
  [NapPhase.GAME_END]: 'gameEnd',
};

/** Renders the Nap (Napoleon) game page: a British 4-player 5-trick gambling trick-taker with a bidding phase. */
export const NapPage = withTutorial(NapPageContent, 'nap', NAP_TUTORIAL_STEPS);

/** Inner content of the Nap page, wrapped by TutorialProvider. */
function NapPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('nap');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    napConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useNapGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('nap');
  const cliConfig: CliGameConfig<NapResponse, Parameters<typeof napApi.exec>> = useMemo(
    () => ({
      gameName: 'nap',
      parseCommand: parseNapCommand,
      formatResponse: formatNapState,
      helpText: NAP_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('nap', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('nap', NAP_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="nap" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;

  const isBidPhase = state.phase === NapPhase.BID;
  const isPlayPhase = state.phase === NapPhase.PLAY;
  const isTrickEnd = state.phase === NapPhase.TRICK_END;
  const isRoundEnd = state.phase === NapPhase.ROUND_END;
  // ラウンドで実際に動いたチップ。達成なら宣言者が make を得て、失敗なら相手が
  // それぞれ fail を得る (減るのは宣言者ではない)。
  const roundPayout = (() => {
    const payout = napPayout(state.contract);
    if (!payout || state.declarerIdx < 0) return null;
    const made = (state.roundTricks[state.declarerIdx] ?? 0) >= state.contract;
    return { made, chips: made ? payout.make : payout.fail };
  })();
  const isGameEnd = state.phase === NapPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);
  const highestBidder = highestBid > 0 ? state.players[state.bids.indexOf(highestBid)] : undefined;
  const highestBidLabelKey = BIDS.find((b) => b.value === highestBid)?.key;
  const highestBidderName = highestBidder ? playerName(highestBidder.id, highestBidder.isHuman) : '';

  const contractName =
    state.declarerIdx >= 0 ? t(`contractName.${CONTRACT_KEYS[state.contract] ?? 'pass'}`) : t('contractUndecided');

  // Declarer progress toward the contract. In Nap a round is 5 tricks and the
  // contract value (2/3/4/5) is exactly the number of tricks the declarer must
  // win, so we can show won/needed and whether it's still reachable.
  const declarer = state.declarerIdx >= 0 ? state.players[state.declarerIdx] : undefined;
  const tricksPlayed = state.players.reduce((sum, p) => sum + p.trickCount, 0);
  const tricksRemaining = NAP_TOTAL_TRICKS - tricksPlayed;
  const declarerWon = declarer?.trickCount ?? 0;
  const tricksStillNeeded = state.contract - declarerWon;
  const contractUnreachable = tricksStillNeeded > tricksRemaining;
  const showDeclarerProgress = (isPlayPhase || isTrickEnd) && state.declarerIdx >= 0;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.nap')}
      gameThemeBg={gameTheme.nap.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanBidTurn) && !isGameEnd}
      gamePath="/nap"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerPlayer === humanIdx}
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
                    value: napConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetPoints',
                    label: t('settings.targetPoints'),
                    value: napConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="nap-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div className="text-ds-text-muted text-center mb-2 text-sm">
              {state.declarerIdx >= 0
                ? t('declarerLine', {
                    name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
                    contract: contractName,
                  })
                : t('contractUndecided')}
            </div>

            {showDeclarerProgress && (
              <div
                className={`text-center mb-2 text-sm ${contractUnreachable ? 'text-ds-error font-semibold' : 'text-ds-text-muted'}`}
                data-testid="nap-declarer-progress"
                role="status"
                aria-live="polite"
              >
                {t('declarerProgress', {
                  won: declarerWon,
                  needed: state.contract,
                  remaining: tricksRemaining,
                })}
                {contractUnreachable && <span className="ml-2">{t('contractUnreachable')}</span>}
              </div>
            )}

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="nap-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Per-player chip scores with declarer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isDeclarer && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                          {t('declarerBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* Players: cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result: tricks won per player */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.tricks', {
                          name: playerName(p.id, p.isHuman),
                          count: state.roundTricks[p.id] ?? 0,
                        })}
                      </div>
                    ))}
                    {/* トリック数だけでは、そのラウンドでチップがいくら動いたのか
                        分からない。達成なら宣言者が、失敗なら相手それぞれが得る (#5651)。 */}
                    {roundPayout && (
                      <div className="mt-1 text-ds-text-primary" data-testid="nap-round-payout">
                        {t(roundPayout.made ? 'roundResult.payoutMade' : 'roundResult.payoutFailed', {
                          name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
                          chips: roundPayout.chips,
                        })}
                      </div>
                    )}
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

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.nap.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="nap"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="nap-action-buttons">
              {isBidPhase && isHumanBidTurn && (
                <>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
                  <span className="text-xs text-ds-text-muted self-center mr-1" data-testid="nap-highest-bid">
                    {highestBid > 0 && highestBidLabelKey
                      ? t('bidHighest', { bid: t(highestBidLabelKey), player: highestBidderName })
                      : t('bidNone')}
                  </span>
                  {BIDS.map((b) => {
                    // Pass (0) is always allowed; a non-pass bid must beat the current highest.
                    const tooLow = b.value !== NapContract.PASS && b.value <= highestBid;
                    const disabled = loading || tooLow;
                    const reason = tooLow
                      ? t('bidTooLow', { bid: highestBidLabelKey ? t(highestBidLabelKey) : '' })
                      : undefined;
                    // The title lives on the wrapping span: browsers suppress native tooltips and
                    // hover events on disabled buttons, so hovering the span still surfaces the reason.
                    const payout = napPayout(b.value);
                    return (
                      <span key={b.value} title={reason} data-testid={`bid-wrap-${b.value}`}>
                        <button
                          type="button"
                          className="px-3 py-2 rounded-lg bg-ds-info text-white text-sm disabled:opacity-40"
                          onClick={() => handleBid(b.value)}
                          disabled={disabled}
                          aria-disabled={disabled}
                          aria-label={reason ? `${t(b.key)} — ${reason}` : undefined}
                          data-testid={`bid-${b.value}`}
                        >
                          {t(b.key)}
                          {/* 賭け金は契約ごとに違い、ナップだけ非対称 (#5651)。
                              失敗時に動くのは「相手それぞれが得る」数。 */}
                          {payout && (
                            <span className="ml-1 text-xs opacity-80">
                              {t('bidStake', { make: payout.make, fail: payout.fail })}
                            </span>
                          )}
                        </button>
                      </span>
                    );
                  })}
                </>
              )}
              {canPlay && (
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
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="nap-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
