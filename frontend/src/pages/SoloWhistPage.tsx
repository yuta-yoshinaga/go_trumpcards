import { useEffect, useMemo, useRef, useState } from 'react';
import type { soloWhistApi } from '../api/gameApi';
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
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useSoloWhistGame } from '../hooks/useSoloWhistGame';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SoloWhistResponse } from '../types/card';
import { SoloWhistContract, SoloWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSoloWhistCommand, SOLO_WHIST_HELP } from '../utils/cli/commands/soloWhistCommands';
import { formatSoloWhistState } from '../utils/cli/formatters/soloWhistFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 = no trump). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Contract i18n key suffixes indexed by contract value (0=Pass…3=Abundance). */
const CONTRACT_KEYS = ['pass', 'solo', 'misere', 'abundance'] as const;

/** Bid button options (Pass/Solo/Misère/Abundance). */
const BIDS: { value: number; key: string }[] = [
  { value: SoloWhistContract.PASS, key: 'bid.pass' },
  { value: SoloWhistContract.SOLO, key: 'bid.solo' },
  { value: SoloWhistContract.MISERE, key: 'bid.misere' },
  { value: SoloWhistContract.ABUNDANCE, key: 'bid.abundance' },
];

/** Solo Whist tutorial step definitions. */
const SOLO_WHIST_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="solowhist-info"]', messageKey: 'tutorial.info', placement: 'bottom', advanceOn: 'next' },
  {
    target: '[data-tutorial="solowhist-trick-display"]',
    messageKey: 'tutorial.trick',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="solowhist-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="solowhist-action-buttons"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SOLO_WHIST_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SoloWhistPhase.BID]: 'bid',
  [SoloWhistPhase.PLAY]: 'play',
  [SoloWhistPhase.TRICK_END]: 'trickEnd',
  [SoloWhistPhase.ROUND_END]: 'roundEnd',
  [SoloWhistPhase.GAME_END]: 'gameEnd',
};

/** Renders the Solo Whist game page: a British 4-player trick-taker with a bidding phase. */
export const SoloWhistPage = withTutorial(SoloWhistPageContent, 'solowhist', SOLO_WHIST_TUTORIAL_STEPS);

/** Inner content of the Solo Whist page, wrapped by TutorialProvider. */
function SoloWhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('solowhist');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    soloWhistConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useSoloWhistGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('solowhist');
  const cliConfig: CliGameConfig<SoloWhistResponse, Parameters<typeof soloWhistApi.exec>> = useMemo(
    () => ({
      gameName: 'solowhist',
      parseCommand: parseSoloWhistCommand,
      formatResponse: formatSoloWhistState,
      helpText: SOLO_WHIST_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('solowhist', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('solowhist', SOLO_WHIST_PHASE_KEYS);

  // Transient pulse when the contract is decided (declarerIdx: -1 → a real seat) —
  // a plan-shaping event (esp. misère) that's otherwise easy to miss in the small text row.
  const [contractPulse, setContractPulse] = useState(false);
  const prevDeclarerRef = useRef<number | null>(null);
  useEffect(() => {
    const declarer = state?.declarerIdx ?? null;
    if (declarer == null) return;
    const prev = prevDeclarerRef.current;
    prevDeclarerRef.current = declarer;
    if (prev != null && prev < 0 && declarer >= 0) {
      setContractPulse(true);
      const id = setTimeout(() => setContractPulse(false), 2500);
      return () => clearTimeout(id);
    }
  }, [state?.declarerIdx]);

  if (!state)
    return <GameSkeleton gameKey="solowhist" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;
  const isHumanBidTurn = state.isHumanBidTurn;

  const isBidPhase = state.phase === SoloWhistPhase.BID;
  const isPlayPhase = state.phase === SoloWhistPhase.PLAY;
  const isTrickEnd = state.phase === SoloWhistPhase.TRICK_END;
  const isRoundEnd = state.phase === SoloWhistPhase.ROUND_END;
  const isGameEnd = state.phase === SoloWhistPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn;
  const trumpSymbol = state.trumpSuit === 0 ? t('noTrump') : (SUIT_SYMBOLS[state.trumpSuit] ?? '?');

  // The current highest (non-pass) bid; a new non-pass bid must beat it.
  const highestBid = Math.max(0, ...state.bids);
  // Resolve the holder via the players array (playerName expects a player id, not an index).
  const highestBidder = highestBid > 0 ? state.players[state.bids.indexOf(highestBid)] : undefined;
  const highestBidLabelKey = BIDS.find((b) => b.value === highestBid)?.key;
  const highestBidderName = highestBidder ? playerName(highestBidder.id, highestBidder.isHuman) : '';

  const contractName =
    state.declarerIdx >= 0 ? t(`contractName.${CONTRACT_KEYS[state.contract] ?? 'pass'}`) : t('contractUndecided');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.solowhist')}
      gameThemeBg={gameTheme.solowhist.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={(isHumanTurn || isHumanBidTurn) && !isGameEnd}
      gamePath="/solowhist"
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
                    value: soloWhistConfig.cpuDifficulty,
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
                    value: soloWhistConfig.targetPoints,
                    options: TARGET_POINTS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetPoints', v),
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
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="solowhist-info">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div
              className={`text-ds-text-muted text-center mb-2 text-sm${
                contractPulse ? ' motion-safe:animate-pulse text-ds-accent font-semibold' : ''
              }`}
              data-testid="solowhist-declarer"
              role="status"
              aria-live="polite"
            >
              {state.declarerIdx >= 0
                ? t('declarerLine', {
                    name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman ?? false),
                    contract: contractName,
                  })
                : t('contractUndecided')}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="solowhist-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div>
                {/* Per-player match scores with declarer badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isDeclarer ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isDeclarer && (
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
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
          <GameFooter className={`${gameTheme.solowhist.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="solowhist"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="solowhist-action-buttons">
              {isBidPhase && isHumanBidTurn && (
                <>
                  <span className="text-xs text-ds-text-muted self-center mr-1">{t('bidPrompt')}</span>
                  <span className="text-xs text-ds-text-muted self-center mr-1" data-testid="sw-highest-bid">
                    {highestBid > 0
                      ? t('bidHighest', {
                          bid: highestBidLabelKey ? t(highestBidLabelKey) : highestBid,
                          player: highestBidderName,
                        })
                      : t('bidNone')}
                  </span>
                  {BIDS.map((b) => {
                    // Pass (0) is always allowed; a non-pass bid must beat the current highest.
                    const tooLow = b.value !== SoloWhistContract.PASS && b.value <= highestBid;
                    const disabled = loading || tooLow;
                    const reason = tooLow ? t('bidTooLow') : undefined;
                    // The title lives on the wrapping span: browsers suppress native tooltips on
                    // disabled buttons, so hovering the span still surfaces the reason.
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
                dataTutorial="solowhist-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
