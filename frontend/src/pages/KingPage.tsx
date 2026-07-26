import { useEffect, useMemo, useState } from 'react';
import type { kingApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, useKingGame } from '../hooks/useKingGame';
import { badgeErrorColors, badgeSuccessColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KingResponse } from '../types/card';
import type { TutorialStep } from '../types/tutorial';
import { KING_HELP, parseKingCommand } from '../utils/cli/commands/kingCommands';
import { formatKingState } from '../utils/cli/formatters/kingFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; 0/-1 = unset). */
const SUIT_SYMBOLS = ['-', '♠', '♣', '♥', '♦'] as const;

/** Selectable trump suits for the King (Trump) contract. */
const TRUMP_SUITS = [1, 2, 3, 4] as const;

/** The "King (Trump)" contract index, which requires a trump-suit choice. */
const KING_TRUMP_CONTRACT = 6;

/** King tutorial step definitions. */
const KING_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="king-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="king-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="king-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="king-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="king-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the King game page: a 4-player 52-card compendium trick-avoidance game. */
export const KingPage = withTutorial(KingPageContent, 'king', KING_TUTORIAL_STEPS);

/** Inner content of the King page, wrapped by TutorialProvider. */
function KingPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('king');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    kingConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    selectContract,
    handlePlay,
    handleNextDeal,
  } = useKingGame();

  // Contract to play when the dealer picks "King (Trump)" and must still pick a suit.
  const [pendingTrumpContract, setPendingTrumpContract] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('king');
  const kingCliConfig: CliGameConfig<KingResponse, Parameters<typeof kingApi.exec>> = useMemo(
    () => ({
      gameName: 'king',
      parseCommand: parseKingCommand,
      formatResponse: formatKingState,
      helpText: KING_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, kingCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('king', state);
  const { cardWidth, isMobile } = useCardDimensions();

  if (!state)
    return <GameSkeleton gameKey="king" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isSelectPhase = state.phase === 'selectContract';
  const isPlayPhase = state.phase === 'play';
  const isDealEnd = state.phase === 'dealEnd';
  const isGameEnd = state.phase === 'gameEnd' || state.gameEndFlag;

  const humanWon = isGameEnd && state.roundWinners.includes(humanIdx);
  const canSelect = isSelectPhase && state.dealerIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;
  const contractName = state.currentContract >= 0 ? t(`contracts.${state.currentContract}`) : '-';
  const trumpSymbol = state.trumpSuit >= 1 ? (SUIT_SYMBOLS[state.trumpSuit] ?? '-') : '-';
  const phaseName = t(`phase.${state.phase}`);

  const handleManualReset = () => {
    hideActionLog();
    setPendingTrumpContract(null);
    reset();
  };

  // Dispatch a contract choice. For "King (Trump)" we first surface a trump
  // picker; every other contract is sent immediately with trumpSuit = -1.
  const handleContractClick = (contract: number) => {
    if (contract === KING_TRUMP_CONTRACT) {
      setPendingTrumpContract(contract);
      return;
    }
    setPendingTrumpContract(null);
    selectContract(contract, -1);
  };

  const handleTrumpClick = (suit: number) => {
    if (pendingTrumpContract === null) return;
    selectContract(pendingTrumpContract, suit);
    setPendingTrumpContract(null);
  };

  return (
    <GamePageShell
      title={tc('nav.king')}
      gameThemeBg={gameTheme.king.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/king"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
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
                    value: kingConfig.cpuDifficulty,
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

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('deal', { n: state.dealNumber + 1, total: state.totalDeals })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('contract', { name: contractName })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('dealer', { name: playerName(state.dealerIdx, state.dealerIdx === humanIdx) })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="king-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="king-info">
                {/* Per-player match scores with a Dealer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.id === state.dealerIdx ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.totalScore })}
                      </span>
                      {p.id === state.dealerIdx && (
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                          {t('dealerBadge')}
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

                {/* Deal result: contract played, its loss basis, and per-player gained points */}
                {(isDealEnd || isGameEnd) && state.lastDealDetail && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="king-deal-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('dealResult.title')}</div>
                    {/* Contract breakdown: which contract was played and why points moved */}
                    <div className="mb-1.5 pb-1.5 border-b border-white/10" data-testid="king-deal-breakdown">
                      <div className="text-ds-text-primary font-semibold">
                        {state.lastDealDetail.contract === KING_TRUMP_CONTRACT && state.lastDealDetail.trumpSuit >= 1
                          ? t('dealResult.contractLineTrump', {
                              name: t(`contracts.${state.lastDealDetail.contract}`),
                              suit: SUIT_SYMBOLS[state.lastDealDetail.trumpSuit] ?? '-',
                            })
                          : t('dealResult.contractLine', {
                              name: t(`contracts.${state.lastDealDetail.contract}`),
                            })}
                      </div>
                      <div className="text-xs opacity-80">
                        {t('dealResult.basis', { desc: t(`contractDesc.${state.lastDealDetail.contract}`) })}
                      </div>
                    </div>
                    {state.players.map((p) => (
                      <div key={p.id} data-testid={`king-deal-breakdown-row-${p.id}`}>
                        {t('dealResult.gained', {
                          name: playerName(p.id, p.isHuman),
                          points: state.lastDealDetail?.gained[p.id] ?? 0,
                        })}
                      </div>
                    ))}
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
          <GameFooter className={`${gameTheme.king.footer} px-4 py-2.5`}>
            {isSelectPhase && !canSelect && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="king-select-cpu">
                {t('selectContractCpu', { id: state.dealerIdx })}
              </div>
            )}
            {canSelect && pendingTrumpContract === null && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="king-select-prompt">
                {t('selectContractPrompt')}
              </div>
            )}
            {canSelect && pendingTrumpContract !== null && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="king-trump-prompt">
                {t('selectTrumpPrompt')}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="king"
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
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="king-action-buttons">
              {canSelect && pendingTrumpContract === null && (
                <div className="flex flex-wrap gap-2" data-testid="king-contract-buttons">
                  {state.usedContracts.map((used, contract) => {
                    // Contract 6 (King/Trump) rewards taking tricks; every other
                    // contract penalises capturing its target cards. See
                    // internal/domain/King.go for the authoritative classification.
                    const isAchieve = contract === KING_TRUMP_CONTRACT;
                    return (
                      <button
                        key={contract}
                        type="button"
                        className={`${btnSecondary} flex flex-col items-center gap-1`}
                        onClick={() => handleContractClick(contract)}
                        disabled={loading || used}
                        title={t(`contractDesc.${contract}`)}
                        data-testid={`king-contract-${contract}`}
                      >
                        <span className={used ? 'line-through opacity-60' : ''}>{t(`contracts.${contract}`)}</span>
                        <span
                          className="flex items-center gap-1 text-xs"
                          aria-hidden="true"
                          data-testid={`king-contract-badge-${contract}`}
                        >
                          <span className={`rounded px-1 py-0.5 ${isAchieve ? badgeSuccessColors : badgeErrorColors}`}>
                            {t(`contractType.${isAchieve ? 'achieve' : 'avoid'}`)}
                          </span>
                          <span>{t(`contractIcon.${contract}`)}</span>
                        </span>
                      </button>
                    );
                  })}
                </div>
              )}
              {canSelect && pendingTrumpContract !== null && (
                <div className="flex flex-wrap gap-2" data-testid="king-trump-buttons">
                  {TRUMP_SUITS.map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleTrumpClick(suit)}
                      disabled={loading}
                    >
                      {SUIT_SYMBOLS[suit]}
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
              {isDealEnd && (
                <button type="button" className={btnSuccess} onClick={handleNextDeal} disabled={loading}>
                  {t('nextDeal')}
                </button>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="king-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
