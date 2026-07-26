import { useEffect, useMemo, useRef, useState } from 'react';
import type { ultiApi } from '../api/gameApi';
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
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useUltiGame } from '../hooks/useUltiGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { UltiResponse } from '../types/card';
import { UltiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseUltiCommand, ULTI_HELP } from '../utils/cli/commands/ultiCommands';
import { formatUltiState } from '../utils/cli/formatters/ultiFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ulti tutorial step definitions. */
const ULTI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ulti-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ulti-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ulti-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ulti-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ulti-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const ULTI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [UltiPhase.BID]: 'bid',
  [UltiPhase.DISCARD]: 'discard',
  [UltiPhase.PLAY]: 'play',
  [UltiPhase.TRICK_END]: 'trickEnd',
  [UltiPhase.ROUND_END]: 'roundEnd',
  [UltiPhase.GAME_END]: 'gameEnd',
};

/** Contract i18n keys indexed by contract value (0=none, 1=Party, 2=Betli, 3=Durchmarsch, 4=Ulti). */
const CONTRACT_KEYS = [
  'contractNone',
  'contractParty',
  'contractBetli',
  'contractDurchmarsch',
  'contractUlti',
] as const;

/** Trump-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Win/made, 2=Loss/failed). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Format a coin delta with an explicit sign so it reads without relying on color alone (e.g. "+2", "-1"). */
function signedCoins(delta: number): string {
  return delta > 0 ? `+${delta}` : `${delta}`;
}

/** Number of talon cards the declarer must discard in the Discard phase (matches `UltiDiscardSize` in `internal/domain/Ulti.go`). */
const DISCARD_COUNT = 2;

/** Selectable trump suits with their playing-card symbols (1=♠ 2=♣ 3=♥ 4=♦). */
const TRUMP_CHOICES = [
  { code: 1, symbol: '♠' },
  { code: 2, symbol: '♣' },
  { code: 3, symbol: '♥' },
  { code: 4, symbol: '♦' },
] as const;

/** Renders the Ulti (Ultimo) game page: a 3-player Hungarian contract 32-card trick-taker with a Party/Betli/Durchmarsch bid, a 2-card talon discard, and coin settlement. */
export const UltiPage = withTutorial(UltiPageContent, 'ulti', ULTI_TUTORIAL_STEPS);

/** Inner content of the Ulti page, wrapped by TutorialProvider. */
function UltiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ulti');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    ultiConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useUltiGame();

  // Trump suit chosen for a pending Party declaration (null until picked).
  const [selectedTrump, setSelectedTrump] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // Per-player coin change at settlement. We remember the pre-settlement balances
  // (the last snapshot taken outside ROUND_END) so that when the round settles we
  // can show each player's signed delta — the settlement is otherwise invisible.
  const prevCoinsRef = useRef<number[] | null>(null);
  const [coinDeltas, setCoinDeltas] = useState<number[] | null>(null);
  useEffect(() => {
    if (!state) return;
    const coins = state.players.map((p) => p.coins);
    // The backend skips ROUND_END on the match-deciding round (it settles then
    // jumps straight to GAME_END), so treat GAME_END as a settlement too or the
    // final round's deltas would never appear.
    const isSettlement = state.phase === UltiPhase.ROUND_END || state.phase === UltiPhase.GAME_END || state.gameEndFlag;
    if (isSettlement) {
      if (prevCoinsRef.current) {
        const prev = prevCoinsRef.current;
        setCoinDeltas(coins.map((c, i) => c - (prev[i] ?? c)));
      }
    } else {
      prevCoinsRef.current = coins;
      setCoinDeltas(null);
    }
  }, [state]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ulti');
  const ultiCliConfig: CliGameConfig<UltiResponse, Parameters<typeof ultiApi.exec>> = useMemo(
    () => ({
      gameName: 'ulti',
      parseCommand: parseUltiCommand,
      formatResponse: formatUltiState,
      helpText: ULTI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, ultiCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ulti', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ulti', ULTI_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="ulti" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 10 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === UltiPhase.BID;
  const isDiscardPhase = state.phase === UltiPhase.DISCARD;
  const isPlayPhase = state.phase === UltiPhase.PLAY;
  const isTrickEnd = state.phase === UltiPhase.TRICK_END;
  const isRoundEnd = state.phase === UltiPhase.ROUND_END;
  const isGameEnd = state.phase === UltiPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canDiscard = isDiscardPhase && isHumanTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  const contractLabel = t(CONTRACT_KEYS[state.contract] ?? 'contractNone');
  const trumpLabel = state.trumpSuit >= 1 ? t(SUIT_KEYS[state.trumpSuit] ?? 'suitNone') : t('suitNone');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  const declareParty = () => {
    if (selectedTrump === null) return;
    handleBid('party', selectedTrump);
    setSelectedTrump(null);
  };

  const declareUlti = () => {
    if (selectedTrump === null) return;
    handleBid('ulti', selectedTrump);
    setSelectedTrump(null);
  };

  return (
    <GamePageShell
      title={tc('nav.ulti')}
      gameThemeBg={gameTheme.ulti.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/ulti"
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
                    value: ultiConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetRounds',
                    label: t('settings.targetRounds'),
                    value: ultiConfig.targetRounds,
                    options: TARGET_ROUNDS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetRounds', v),
                  },
                  hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
                ],
              },
            ]}
          />

          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('contract', { contract: contractLabel })}</span>
              <span>{t('trump', { suit: trumpLabel })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="ulti-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="ulti-info">
                {/* Per-player coin balances with a declarer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p, i) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('coins', { coins: p.coins })}
                      </span>
                      {(isRoundEnd || isGameEnd) && coinDeltas && coinDeltas[i] !== 0 && (
                        <span
                          className={`text-xs font-semibold ${coinDeltas[i] > 0 ? 'text-ds-success' : 'text-ds-error'}`}
                          data-testid={`ulti-coin-delta-${p.id}`}
                        >
                          {t('coinDelta', { delta: signedCoins(coinDeltas[i]) })}
                        </span>
                      )}
                      {p.isDeclarer && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                          {t('declarerBadge')}
                        </span>
                      )}
                    </div>
                  ))}
                </div>

                {/* Players: cards / tricks / captured points */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                          {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                        {t('tricks', { count: p.trickCount })} | {t('points', { points: p.cardPoints })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result: the deal outcome (contract made / failed) */}
                {(isRoundEnd || isGameEnd) && state.outcome > 0 && (
                  <div
                    className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    role="status"
                    aria-live="polite"
                    data-testid="ulti-round-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
                    {state.declarerIdx >= 0 && (
                      <div>
                        {t('roundResult.declarer', {
                          name: playerName(state.declarerIdx, state.declarerIdx === humanIdx),
                        })}
                      </div>
                    )}
                    {(isRoundEnd || isGameEnd) && coinDeltas && humanIdx >= 0 && (
                      <div
                        className={
                          coinDeltas[humanIdx] > 0 ? 'text-ds-success' : coinDeltas[humanIdx] < 0 ? 'text-ds-error' : ''
                        }
                      >
                        {t('roundResult.yourCoins', { delta: signedCoins(coinDeltas[humanIdx] ?? 0) })}
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
          <GameFooter className={`${gameTheme.ulti.footer} px-4 py-2.5`}>
            {canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="ulti-bid-prompt">
                {t('bidPhase')}
              </div>
            )}
            {canDiscard && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="ulti-discard-prompt">
                <div>{t('discardPhase')}</div>
                <div className="text-ds-text-muted" data-testid="ulti-discard-progress">
                  {t('discardProgress', { selected: selectedCardIndices.length, required: DISCARD_COUNT })}
                </div>
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="ulti"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="ulti-action-buttons">
              {canBid && (
                <>
                  <span className="text-ds-text-muted text-sm">{t('chooseTrump')}:</span>
                  {TRUMP_CHOICES.map((c) => (
                    <button
                      key={c.code}
                      type="button"
                      className={selectedTrump === c.code ? btnPrimary : btnSecondary}
                      onClick={() => setSelectedTrump(c.code)}
                      disabled={loading}
                      aria-label={t(SUIT_KEYS[c.code])}
                      aria-pressed={selectedTrump === c.code}
                    >
                      {c.symbol}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={declareParty}
                    disabled={loading || selectedTrump === null}
                  >
                    {t('bidParty')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={declareUlti}
                    disabled={loading || selectedTrump === null}
                  >
                    {t('bidUlti')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => handleBid('betli')} disabled={loading}>
                    {t('bidBetli')}
                  </button>
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleBid('durchmarsch')}
                    disabled={loading}
                  >
                    {t('bidDurchmarsch')}
                  </button>
                </>
              )}
              {canDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 2}
                >
                  {t('discardButton')}
                </button>
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
                dataTutorial="ulti-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
