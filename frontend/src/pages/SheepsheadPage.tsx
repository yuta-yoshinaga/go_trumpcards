import { useCallback, useEffect, useMemo } from 'react';
import type { sheepsheadApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardNavShortcutsPanel } from '../components/CardNavShortcutsPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  CPU_DIFFICULTY_OPTIONS,
  SHEEPSHEAD_BURY_COUNT,
  TARGET_CHIPS_OPTIONS,
  useSheepsheadGame,
} from '../hooks/useSheepsheadGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { SheepsheadResponse } from '../types/card';
import { SheepsheadPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSheepsheadCommand, SHEEPSHEAD_HELP } from '../utils/cli/commands/sheepsheadCommands';
import { formatSheepsheadState } from '../utils/cli/formatters/sheepsheadFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Sheepshead tutorial step definitions. */
const SH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sh-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sh-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SHEEPSHEAD_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SheepsheadPhase.PICK]: 'pick',
  [SheepsheadPhase.BURY]: 'bury',
  [SheepsheadPhase.CALL]: 'call',
  [SheepsheadPhase.PLAY]: 'play',
  [SheepsheadPhase.TRICK_END]: 'trickEnd',
  [SheepsheadPhase.ROUND_END]: 'roundEnd',
  [SheepsheadPhase.GAME_END]: 'gameEnd',
};

/** Maps a called-suit id (1=♠ 2=♣ 3=♥) to its i18n suit key. */
const SUIT_KEYS: Readonly<Record<number, string>> = { 1: 'spade', 2: 'club', 3: 'heart' };

/** Renders the Sheepshead game page: 5-player partnership trick-taking with pick/bury/call/play phases. */
export const SheepsheadPage = withTutorial(SheepsheadPageContent, 'sheepshead', SH_TUTORIAL_STEPS);

/** Inner content of the Sheepshead page, wrapped by TutorialProvider. */
function SheepsheadPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('sheepshead');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    sheepsheadConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    reset,
    handlePick,
    handlePass,
    handleBury,
    handleCall,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useSheepsheadGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('sheepshead');
  const sheepsheadCliConfig: CliGameConfig<SheepsheadResponse, Parameters<typeof sheepsheadApi.exec>> = useMemo(
    () => ({
      gameName: 'sheepshead',
      parseCommand: parseSheepsheadCommand,
      formatResponse: formatSheepsheadState,
      helpText: SHEEPSHEAD_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, sheepsheadCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('sheepshead', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('sheepshead', SHEEPSHEAD_PHASE_KEYS);

  // Keyboard hand navigation: number keys toggle a card, Enter confirms,
  // Escape clears. BURY needs two cards and PLAY needs one; both handlers
  // self-guard on the selected count, so confirmAction just routes to the
  // active phase's handler.
  const humanIdxForKbd = state?.players.findIndex((p) => p.isHuman) ?? -1;
  const canBuryForKbd = state?.phase === SheepsheadPhase.BURY && state.pickerIdx === humanIdxForKbd;
  const canPlayForKbd = state?.phase === SheepsheadPhase.PLAY && state.currentPlayerIdx === humanIdxForKbd;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;
  const confirmAction = useCallback(() => {
    if (canBuryForKbd) handleBury();
    else if (canPlayForKbd) handlePlay();
  }, [canBuryForKbd, canPlayForKbd, handleBury, handlePlay]);
  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: (canBuryForKbd || canPlayForKbd) && !loading,
  });

  if (!state)
    return <GameSkeleton gameKey="sheepshead" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 6 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;

  const isPickPhase = state.phase === SheepsheadPhase.PICK;
  const isBuryPhase = state.phase === SheepsheadPhase.BURY;
  const isCallPhase = state.phase === SheepsheadPhase.CALL;
  const isPlayPhase = state.phase === SheepsheadPhase.PLAY;
  const isTrickEnd = state.phase === SheepsheadPhase.TRICK_END;
  const isRoundEnd = state.phase === SheepsheadPhase.ROUND_END;
  const isGameEnd = state.phase === SheepsheadPhase.GAME_END || state.gameEndFlag;

  // Phase-specific "it's the human's actionable turn" flags.
  const canPick = isPickPhase && isHumanTurn;
  const canBury = isBuryPhase && state.pickerIdx === humanIdx;
  const canCall = isCallPhase && state.pickerIdx === humanIdx && state.callableSuits.length > 0;
  const canPlay = isPlayPhase && isHumanTurn;

  // Partner is only shown once revealed (or at round/game end).
  const showPartner = state.partnerRevealed || isRoundEnd || isGameEnd;
  const pickerLabel = state.pickerIdx >= 0 ? playerName(state.pickerIdx, state.pickerIdx === humanIdx) : '-';
  const partnerLabel =
    showPartner && state.partnerIdx >= 0
      ? playerName(state.partnerIdx, state.partnerIdx === humanIdx)
      : t('partnerHidden');

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.sheepshead')}
      gameThemeBg={gameTheme.sheepshead.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/sheepshead"
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
                    value: sheepsheadConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetChips',
                    label: t('settings.targetChips'),
                    value: sheepsheadConfig.targetChips,
                    options: TARGET_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetChips', v),
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
              <span>{t('blindCount', { count: state.blindCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                {/* Blind: face-down cards, shown only during the Pick phase before a picker takes them. */}
                {isPickPhase && state.blindCount > 0 && (
                  <div className="mb-3" data-testid="sh-blind-display">
                    <div className="text-ds-text-muted text-sm mb-1 text-center">
                      {t('blind')} ({state.blindCount})
                    </div>
                    <div className="flex justify-center gap-2">
                      {Array.from({ length: state.blindCount }).map((_, i) => (
                        <AnimatedCardBack key={i} width={cardWidth} dealDelay={i * 0.08} />
                      ))}
                    </div>
                  </div>
                )}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="sh-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="sh-info">
                {/* Picker / partner / called suit */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  <div>
                    {t('picker')}: {pickerLabel}
                  </div>
                  <div>
                    {t('partner')}: {partnerLabel}
                  </div>
                  <div>
                    {t('calledSuit')}:{' '}
                    {state.calledSuit > 0 && SUIT_KEYS[state.calledSuit]
                      ? t(`suit.${SUIT_KEYS[state.calledSuit]}`)
                      : t('none')}
                  </div>
                  {isPickPhase && <div>{t('passCount', { count: state.passCount })}</div>}
                </div>

                {/* Players: chips / cards / tricks */}
                {isMobile ? (
                  <details className="mb-2 p-2 rounded bg-black/30">
                    <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
                    <div className="mt-1">
                      {state.players.map((p) => (
                        <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                          {playerName(p.id, p.isHuman)}: {t('chips', { count: p.chips })} |{' '}
                          {t('tricks', { count: p.trickCount })}
                        </div>
                      ))}
                    </div>
                  </details>
                ) : (
                  <div className="mb-2 p-2 rounded bg-black/30">
                    {state.players.map((p) => (
                      <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                        {playerName(p.id, p.isHuman)}: {t('chips', { count: p.chips })} |{' '}
                        {t('tricks', { count: p.trickCount })}
                      </div>
                    ))}
                  </div>
                )}

                {/* Round result */}
                {(isRoundEnd || isGameEnd) && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.pickerPoints', { points: state.roundPickerPoints })}</div>
                    <div>{t('roundResult.multiplier', { multiplier: state.roundMultiplier })}</div>
                    <div>{state.roundPickerWon ? t('roundResult.pickerWon') : t('roundResult.pickerLost')}</div>
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
          <GameFooter className={`${gameTheme.sheepshead.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="sh"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {state.hint && isRequestedHint(state) && (
              <div className="text-ds-warning text-sm mb-2">
                {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                {state.hint.cardIndices.length > 0 && ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
              </div>
            )}
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="sh-action-buttons">
              {canPick && (
                <>
                  <button type="button" className={btnPrimary} onClick={handlePick} disabled={loading}>
                    {t('pickButton')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}
              {canBury && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleBury}
                  disabled={loading || selectedCardIndices.length !== SHEEPSHEAD_BURY_COUNT}
                >
                  {t('buryButton')} ({t('burySelected', { count: selectedCardIndices.length })})
                </button>
              )}
              {canCall &&
                state.callableSuits.map((suit) => {
                  // callableSuits are always 1-3 (♠♣♥), all present in SUIT_KEYS.
                  const suitName = t(`suit.${SUIT_KEYS[suit]}`);
                  return (
                    <button
                      key={`call-${suit}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleCall(suit)}
                      disabled={loading}
                      aria-label={t('callButtonAriaLabel', { suit: suitName })}
                    >
                      {t('callButton', { suit: suitName })}
                    </button>
                  );
                })}
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
                dataTutorial="sh-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="sheepshead-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
