import { useEffect, useMemo } from 'react';
import type { cinchApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCinchGame } from '../hooks/useCinchGame';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CinchResponse } from '../types/card';
import { CinchPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { estimateCinchBidStrength } from '../utils/cinchBidStrength';
import { CINCH_HELP, parseCinchCommand } from '../utils/cli/commands/cinchCommands';
import { formatCinchState } from '../utils/cli/formatters/cinchFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; 0 = unset). */
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'] as const;

/** i18n suit-name key by suit number (1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_KEYS: Readonly<Record<number, string>> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };

/** Hearts and diamonds render red; spades and clubs stay the default text color. */
const isRedSuit = (suit: number): boolean => suit === 3 || suit === 4;

/** Format a signed match-point delta for display (e.g. 6 -> "+6", -8 -> "-8", 0 -> "0"). */
const signedDelta = (n: number): string => (n > 0 ? `+${n}` : String(n));

/** Selectable trump suits named by the bid winner. */
const TRUMP_SUITS = [1, 2, 3, 4] as const;

/** Maximum Cinch bid (all 14 points). */
const MAX_BID = 14;

/** Cinch tutorial step definitions. */
const CINCH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="cinch-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cinch-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="cinch-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="cinch-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="cinch-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CINCH_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CinchPhase.BID]: 'bid',
  [CinchPhase.NAME_TRUMP]: 'nameTrump',
  [CinchPhase.PLAY]: 'play',
  [CinchPhase.TRICK_END]: 'trickEnd',
  [CinchPhase.ROUND_END]: 'roundEnd',
  [CinchPhase.GAME_END]: 'gameEnd',
};

/** Renders the Cinch (Double Pedro) game page: a 4-player 52-card All-Fours/Pitch-family bidding trick-taker. */
export const CinchPage = withTutorial(CinchPageContent, 'cinch', CINCH_TUTORIAL_STEPS);

/** Inner content of the Cinch page, wrapped by TutorialProvider. */
function CinchPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('cinch');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    cinchConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handleNameTrump,
    handlePlay,
    handleNextDeal,
  } = useCinchGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('cinch');
  const cinchCliConfig: CliGameConfig<CinchResponse, Parameters<typeof cinchApi.exec>> = useMemo(
    () => ({
      gameName: 'cinch',
      parseCommand: parseCinchCommand,
      formatResponse: formatCinchState,
      helpText: CINCH_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cinchCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('cinch', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('cinch', CINCH_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="cinch" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === CinchPhase.BID;
  const isNameTrumpPhase = state.phase === CinchPhase.NAME_TRUMP;
  const isPlayPhase = state.phase === CinchPhase.PLAY;
  const isRoundEnd = state.phase === CinchPhase.ROUND_END;
  const isGameEnd = state.phase === CinchPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.bidPlayerIdx === humanIdx && isHumanTurn;
  const canNameTrump = isNameTrumpPhase && state.bidWinnerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;

  // Rough, hand-only strength guide shown alongside the bid buttons so a player
  // can gauge how high to bid. Purely advisory — not a capture guarantee.
  const bidStrength = canBid && humanPlayer ? estimateCinchBidStrength(humanPlayer.cards) : null;

  // Legal bids: pass (0) plus any raise strictly above the current highest bid, up to 14.
  const minRaise = Math.max(1, state.currentBid + 1);
  const bidChoices: number[] = [];
  for (let b = minRaise; b <= MAX_BID; b++) bidChoices.push(b);

  /** Localized suit name for a suit number (e.g. 3 -> "ハート"). */
  const suitLabel = (suit: number): string => (SUIT_KEYS[suit] ? t(`suit.${SUIT_KEYS[suit]}`) : '');
  /** Colored suit symbol: hearts/diamonds red, spades/clubs default. */
  const renderSuitSymbol = (suit: number) => (
    <span className={isRedSuit(suit) ? 'text-ds-error' : undefined}>{SUIT_SYMBOLS[suit]}</span>
  );

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.cinch')}
      gameThemeBg={gameTheme.cinch.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/cinch"
      gameEndFlag={isGameEnd}
      winShow={isGameEnd && state.winnerIdx === humanIdx}
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
                    value: cinchConfig.cpuDifficulty,
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
                    value: cinchConfig.pointLimit,
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
              <span className="mr-4">{t('deal', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber, total: state.totalTricks })}</span>
              <span className="mr-4">{t('highBid', { bid: state.currentBid })}</span>
              <span className="mr-4" data-testid="cinch-trump-header">
                {t('trumpLabel')}:{' '}
                {state.trumpSuit >= 1 ? (
                  <>
                    {renderSuitSymbol(state.trumpSuit)} {suitLabel(state.trumpSuit)}
                  </>
                ) : (
                  '-'
                )}
              </span>
              <span>{t('target', { points: state.config.pointLimit })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="cinch-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="cinch-info">
                {/* Per-player match scores with a Bidder badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.id === state.bidWinnerIdx ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.totalScore })}
                      </span>
                      {p.id === state.bidWinnerIdx && (
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                          {t('bidderBadge')}
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

                {/* Deal result: per-player gained points */}
                {(isRoundEnd || isGameEnd) && state.lastDealDetail && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="cinch-deal-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
                    {state.lastDealDetail.bidderIdx >= 0 && (
                      <div
                        className={`mb-1 ${state.lastDealDetail.setBack ? 'text-ds-error font-semibold' : 'text-ds-text-primary'}`}
                        data-testid="cinch-bidder-detail"
                      >
                        {t(state.lastDealDetail.setBack ? 'dealResult.bidderSet' : 'dealResult.bidderMade', {
                          name: playerName(state.lastDealDetail.bidderIdx, state.lastDealDetail.bidderIdx === humanIdx),
                          bid: state.lastDealDetail.bid,
                          captured: state.lastDealDetail.points[state.lastDealDetail.bidderIdx] ?? 0,
                          delta: signedDelta(state.lastDealDetail.gained[state.lastDealDetail.bidderIdx] ?? 0),
                        })}
                      </div>
                    )}
                    {state.players.map((p) => {
                      const isSetBackRow = p.id === state.lastDealDetail?.bidderIdx && state.lastDealDetail?.setBack;
                      return (
                        <div
                          key={p.id}
                          className={isSetBackRow ? 'text-ds-error font-semibold' : undefined}
                          data-testid={isSetBackRow ? 'cinch-setback-row' : undefined}
                        >
                          {t('dealResult.gained', {
                            name: playerName(p.id, p.isHuman),
                            points: state.lastDealDetail?.gained[p.id] ?? 0,
                          })}
                        </div>
                      );
                    })}
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
          <GameFooter className={`${gameTheme.cinch.footer} px-4 py-2.5`}>
            {isBidPhase && !canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-bid-cpu">
                {t('bidCpu', { id: state.bidPlayerIdx })}
              </div>
            )}
            {canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-bid-prompt">
                {t('bidPrompt')}
              </div>
            )}
            {canBid && bidStrength && (
              <div
                className="mb-2 mx-auto max-w-xl p-2 rounded bg-black/30 text-ds-text-muted text-xs text-center"
                data-testid="cinch-bid-strength"
              >
                <div className="text-ds-text-primary" data-testid="cinch-bid-strength-range">
                  {t('bidStrength.range', { min: bidStrength.minPoints, max: bidStrength.maxPoints })}
                </div>
                <div data-testid="cinch-bid-strength-best">
                  {t('bidStrength.best', {
                    symbol: SUIT_SYMBOLS[bidStrength.bestSuit],
                    suit: suitLabel(bidStrength.bestSuit),
                    points: bidStrength.pointsBySuit[bidStrength.bestSuit],
                  })}
                </div>
                <div className="mt-1">
                  <span className="text-ds-text-primary">{t('bidStrength.legendTitle')}:</span>{' '}
                  {t('bidStrength.legend')}
                </div>
                <div className="mt-1 italic">{t('bidStrength.note')}</div>
              </div>
            )}
            {canNameTrump && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="cinch-trump-prompt">
                {t('trumpPrompt')}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="cinch"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices &&
                  state.hint.cardIndices.length > 0 &&
                  ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="cinch-action-buttons">
              {canBid && (
                <div className="flex flex-wrap gap-2" data-testid="cinch-bid-buttons">
                  <button type="button" className={btnSecondary} onClick={() => handleBid(0)} disabled={loading}>
                    {t('bidPass')}
                  </button>
                  {bidChoices.map((b) => (
                    <button
                      key={b}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBid(b)}
                      disabled={loading}
                    >
                      {b}
                    </button>
                  ))}
                </div>
              )}
              {canNameTrump && (
                <div className="flex flex-wrap gap-2" data-testid="cinch-trump-buttons">
                  {TRUMP_SUITS.map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleNameTrump(suit)}
                      disabled={loading}
                      aria-label={suitLabel(suit)}
                    >
                      {renderSuitSymbol(suit)}
                    </button>
                  ))}
                </div>
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
              {isRoundEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextDeal} disabled={loading}>
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="cinch-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
