import { useEffect, useMemo } from 'react';
import type { knockoutWhistApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, useKnockoutWhistGame } from '../hooks/useKnockoutWhistGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeInfoColors, badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { KnockoutWhistResponse } from '../types/card';
import { KnockoutWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KNOCKOUT_WHIST_HELP, parseKnockoutWhistCommand } from '../utils/cli/commands/knockoutWhistCommands';
import { formatKnockoutWhistState } from '../utils/cli/formatters/knockoutWhistFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';

/** Suit symbols indexed by suit number (1=♠ 2=♣ 3=♥ 4=♦; index 0 unused). */
const SUIT_SYMBOLS = ['', '♠', '♣', '♥', '♦'] as const;

/** Knockout Whist tutorial step definitions. */
const KNOCKOUT_WHIST_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="knockoutwhist-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="knockoutwhist-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="knockoutwhist-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="knockoutwhist-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="knockoutwhist-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KNOCKOUT_WHIST_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KnockoutWhistPhase.PLAY]: 'play',
  [KnockoutWhistPhase.TRICK_END]: 'trickEnd',
  [KnockoutWhistPhase.ROUND_END]: 'roundEnd',
  [KnockoutWhistPhase.GAME_END]: 'gameEnd',
  [KnockoutWhistPhase.TRUMP_SELECT]: 'trumpSelect',
};

/** Renders the Knockout Whist game page: a British play-only survival trick-taker with elimination and Dogbone tokens. */
export const KnockoutWhistPage = withTutorial(KnockoutWhistPageContent, 'knockoutwhist', KNOCKOUT_WHIST_TUTORIAL_STEPS);

/** Inner content of the Knockout Whist page, wrapped by TutorialProvider. */
function KnockoutWhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('knockoutwhist');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    knockoutWhistConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleSelectTrump,
  } = useKnockoutWhistGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('knockoutwhist');
  const knockoutWhistCliConfig: CliGameConfig<
    KnockoutWhistResponse,
    Parameters<typeof knockoutWhistApi.exec>
  > = useMemo(
    () => ({
      gameName: 'knockoutwhist',
      parseCommand: parseKnockoutWhistCommand,
      formatResponse: formatKnockoutWhistState,
      helpText: KNOCKOUT_WHIST_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, knockoutWhistCliConfig, state, {
    addInput,
    addOutput,
    addError,
    clearLog,
  });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('knockoutwhist', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('knockoutwhist', KNOCKOUT_WHIST_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton gameKey="knockoutwhist" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 7 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isPlayPhase = state.phase === KnockoutWhistPhase.PLAY;
  const isTrickEnd = state.phase === KnockoutWhistPhase.TRICK_END;
  const isRoundEnd = state.phase === KnockoutWhistPhase.ROUND_END;
  const isTrumpSelect = state.phase === KnockoutWhistPhase.TRUMP_SELECT;
  const isGameEnd = state.phase === KnockoutWhistPhase.GAME_END || state.gameEndFlag;

  const canPlay = isPlayPhase && isHumanTurn && !(state.players[humanIdx]?.eliminated ?? false);
  const trumpSymbol = SUIT_SYMBOLS[state.trumpSuit] ?? '?';

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  /** Renders one player panel showing Dogbones, round tricks, and elimination / leader / round-winner state. */
  const renderPlayerPanel = (p: KnockoutWhistResponse['players'][number]) => {
    const isLeader = !isGameEnd && state.leadPlayerIdx === p.id;
    const isRoundWinner = (isRoundEnd || isGameEnd) && state.roundWinnerIdx === p.id;
    return (
      <div
        key={p.id}
        className={`py-0.5 flex items-center gap-2 ${p.eliminated ? 'opacity-70 line-through' : ''}`}
        data-eliminated={p.eliminated ? 'true' : 'false'}
      >
        <span className={p.eliminated ? '' : 'text-ds-text-muted'}>
          {playerName(p.id, p.isHuman)}
          {p.eliminated
            ? ` — ${t('eliminated')}`
            : ` — ${t('roundTricks', { count: p.roundTricks })} · ${t('dogbones', { count: p.dogbones })}`}
        </span>
        {isLeader && (
          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeInfoColors}`} aria-label={t('leader')}>
            {t('leader')}
          </span>
        )}
        {isRoundWinner && (
          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`} aria-label={t('roundWinner')}>
            {t('roundWinner')}
          </span>
        )}
      </div>
    );
  };

  return (
    <GamePageShell
      title={tc('nav.knockoutwhist')}
      gameThemeBg={gameTheme.knockoutwhist.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/knockoutwhist"
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
                    value: knockoutWhistConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
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
              <span className="mr-4">{t('handSize', { n: state.handSize })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              <span className="mr-4">{t('trump', { suit: trumpSymbol })}</span>
              <span>{t('activeCount', { n: state.activeCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="knockoutwhist-trick-display"
                />
              </div>

              {/* Right: info sidebar — 4 player panels */}
              <div data-tutorial="knockoutwhist-info">
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1 text-sm">{state.players.map(renderPlayerPanel)}</div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30 text-sm">{state.players.map(renderPlayerPanel)}</div>
                )}

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && state.roundWinnerIdx >= 0 && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>
                      {t('roundResult.winner', {
                        name: playerName(state.roundWinnerIdx, state.players[state.roundWinnerIdx]?.isHuman ?? false),
                      })}
                    </div>
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
          <GameFooter className={`${gameTheme.knockoutwhist.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="knockoutwhist"
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="knockoutwhist-action-buttons">
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
              {isTrumpSelect && (
                <div className="flex flex-wrap items-center gap-2" data-testid="knockoutwhist-trump-select">
                  <span className="text-ds-text-primary text-sm">{t('trumpSelectPrompt')}</span>
                  {[1, 2, 3, 4].map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleSelectTrump(suit)}
                      disabled={loading}
                      data-testid={`knockoutwhist-trump-${suit}`}
                      aria-label={t('trumpSelectSuit', { suit: SUIT_SYMBOLS[suit] })}
                    >
                      {SUIT_SYMBOLS[suit]}
                    </button>
                  ))}
                </div>
              )}
              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="knockoutwhist-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
