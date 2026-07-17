import { useEffect, useMemo } from 'react';
import type { calabresellaApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_POINTS_OPTIONS, useCalabresellaGame } from '../hooks/useCalabresellaGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { CalabresellaResponse } from '../types/card';
import { CalabresellaPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { CALABRESELLA_HELP, parseCalabresellaCommand } from '../utils/cli/commands/calabresellaCommands';
import { formatCalabresellaState } from '../utils/cli/formatters/calabresellaFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Calabresella tutorial step definitions. */
const CALABRESELLA_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="calabresella-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="calabresella-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="calabresella-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="calabresella-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="calabresella-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const CALABRESELLA_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CalabresellaPhase.BID]: 'bid',
  [CalabresellaPhase.DISCARD]: 'discard',
  [CalabresellaPhase.PLAY]: 'play',
  [CalabresellaPhase.TRICK_END]: 'trickEnd',
  [CalabresellaPhase.ROUND_END]: 'roundEnd',
  [CalabresellaPhase.GAME_END]: 'gameEnd',
};

/** Bid labels indexed by bid value (0=pass/none, 1=chiamo, 2=solo). */
const BID_KEYS = ['bidNone', 'bidChiamo', 'bidSolo'] as const;

/** Renders the Calabresella (Terziglio) game page: a Calabrian/Italian 3-player 40-card Tressette-family trick-taker with bidding and a monte exchange. */
export const CalabresellaPage = withTutorial(CalabresellaPageContent, 'calabresella', CALABRESELLA_TUTORIAL_STEPS);

/** Inner content of the Calabresella page, wrapped by TutorialProvider. */
function CalabresellaPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('calabresella');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    calabresellaConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useCalabresellaGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('calabresella');
  const calabresellaCliConfig: CliGameConfig<CalabresellaResponse, Parameters<typeof calabresellaApi.exec>> = useMemo(
    () => ({
      gameName: 'calabresella',
      parseCommand: parseCalabresellaCommand,
      formatResponse: formatCalabresellaState,
      helpText: CALABRESELLA_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, calabresellaCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('calabresella', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('calabresella', CALABRESELLA_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton gameKey="calabresella" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === CalabresellaPhase.BID;
  const isDiscardPhase = state.phase === CalabresellaPhase.DISCARD;
  const isPlayPhase = state.phase === CalabresellaPhase.PLAY;
  const isTrickEnd = state.phase === CalabresellaPhase.TRICK_END;
  const isRoundEnd = state.phase === CalabresellaPhase.ROUND_END;
  const isGameEnd = state.phase === CalabresellaPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.currentBidderIdx === humanIdx;
  const canDiscard = isDiscardPhase && state.soloistIdx === humanIdx;
  const canPlay = isPlayPhase && isHumanTurn;
  // The soloist takes the 4-card monte (16 cards) and discards down to the regulation
  // 12-card hand (CalabresellaHandSize in the backend).
  const REGULATION_HAND_SIZE = 12;
  const discardRemaining = humanPlayer ? Math.max(0, humanPlayer.cards.length - REGULATION_HAND_SIZE) : 0;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.calabresella')}
      gameThemeBg={gameTheme.calabresella.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/calabresella"
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
                    value: calabresellaConfig.cpuDifficulty,
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
                    value: calabresellaConfig.targetPoints,
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
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('winningBid', { bid: t(BID_KEYS[state.winningBid] ?? 'bidNone') })}</span>
              <span>{t('target', { points: state.config.targetPoints })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="calabresella-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="calabresella-info">
                {/* Per-player match scores with Soloist badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isSoloist ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isSoloist && (
                        <span className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs">
                          {t('soloistBadge')}
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

                {/* Round result: per-player thirds captured */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    {state.players.map((p) => (
                      <div key={p.id}>
                        {t('roundResult.thirds', {
                          name: playerName(p.id, p.isHuman),
                          thirds: state.roundThirds[p.id] ?? 0,
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
          <GameFooter className={`${gameTheme.calabresella.footer} px-4 py-2.5`}>
            {isBidPhase && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="calabresella-bid-prompt"
              >
                {t('bidPhase')}
              </div>
            )}
            {canDiscard && discardRemaining > 0 && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="calabresella-discard-prompt"
              >
                {t('discardPhaseRemaining', { count: discardRemaining })}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="calabresella"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="calabresella-action-buttons">
              {canBid && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => handleBid(1)} disabled={loading}>
                    {t('bidChiamo')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => handleBid(2)} disabled={loading}>
                    {t('bidSolo')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={() => handleBid(0)} disabled={loading}>
                    {t('bidPass')}
                  </button>
                </>
              )}
              {canDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 1}
                  data-testid="calabresella-discard-button"
                >
                  {discardRemaining > 0 ? `${t('discardCard')} (${discardRemaining})` : t('discardCard')}
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
                dataTutorial="calabresella-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
