import { useCallback, useMemo } from 'react';
import type { prsiApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, usePrsiGame } from '../hooks/usePrsiGame';
import { useSound } from '../providers/SoundProvider';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, PrsiResponse } from '../types/card';
import { PrsiPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { PRSI_HELP, parsePrsiCommand } from '../utils/cli/commands/prsiCommands';
import { formatPrsiState } from '../utils/cli/formatters/prsiFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { isPrsiLegalPlay } from '../utils/prsiLegal';
import { hintCheckboxItem } from '../utils/settingsItems';

const PRSI_PHASE_KEYS: Readonly<Record<number, string>> = {
  [PrsiPhase.PLAY]: 'play',
  [PrsiPhase.GAME_END]: 'gameEnd',
};

/** Prší tutorial step definitions. */
const PRSI_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="prsi-discard-pile"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="prsi-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="prsi-play-draw"]', messageKey: 'tutorial.playDraw', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="prsi-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Prší game page with suit/rank matching card play. */
export const PrsiPage = withTutorial(PrsiPageContent, 'prsi', PRSI_TUTORIAL_STEPS);
/** Inner content of the Prší page, wrapped by TutorialProvider. */
function PrsiPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('prsi');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    retry,
    prsiConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
  } = usePrsiGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('prsi', state);
  const { cardWidth } = useCardDimensions();
  const { playSound } = useSound();

  // Drawing from the stock keeps its own shuffle sound: the central tap maps
  // only the `reset` command to shuffle, so this action would otherwise get
  // the generic card sound.
  const handleDrawWithSound = useCallback(() => {
    playSound('shuffle');
    handleDraw();
  }, [handleDraw, playSound]);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('prsi');
  const cliConfig: CliGameConfig<PrsiResponse, Parameters<typeof prsiApi.exec>> = useMemo(
    () => ({
      gameName: 'prsi',
      parseCommand: parsePrsiCommand,
      formatResponse: formatPrsiState,
      helpText: PRSI_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(gameExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === PrsiPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void gameExec('reset', undefined, { cpuDifficulty: prsiConfig.cpuDifficulty });
  }, [gameExec, hideActionLog, prsiConfig.cpuDifficulty]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('prsi', PRSI_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton
        gameKey="prsi"
        layout={{ kind: 'trick-taking', centerCard: true, trickArea: true, footerHandSize: 5 }}
      />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === PrsiPhase.PLAY;
  const isGameEnd = state.phase === PrsiPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const hasPenalty = state.penaltyDrawCount > 0;
  // **エース/ジャックのスキップも重ねられる。**7 の累積ペナルティは「+N」バッジと
  // 警告バナーで目立たせているのに、pendingSkips は一度も読まれていなかった
  // (#4772)。2以上になると複数人が連続で飛ばされる、同じくらい大きな状態変化。
  const hasSkips = state.pendingSkips > 0;

  return (
    <GamePageShell
      title={tc('nav.prsi')}
      gameThemeBg={gameTheme.prsi.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/prsi"
      gameEndFlag={!!state.gameEndFlag}
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
                    value: prsiConfig.cpuDifficulty,
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
              <span>{t('drawPile', { count: state.drawPileCount })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Discard pile top + clickable stock pile */}
                <div className="my-3 flex items-start gap-4 flex-wrap">
                  {state.discardTop && (
                    <div className="p-3 rounded bg-black/40 flex items-center gap-3" data-tutorial="prsi-discard-pile">
                      <AnimatedCard card={state.discardTop} width={cardWidth} />
                      <div className="text-ds-text-muted text-sm">
                        <div>{t('discardTop')}</div>
                      </div>
                    </div>
                  )}

                  {/* Stock pile: a face-down stack the human can click to draw. */}
                  <div className="p-3 rounded bg-black/40 flex flex-col items-center gap-1">
                    <button
                      type="button"
                      data-testid="prsi-stock"
                      onClick={handleDrawWithSound}
                      disabled={!isHumanTurn || loading || state.drawPileCount === 0}
                      aria-label={t('stockAria', { count: state.drawPileCount })}
                      className={`relative ${focusRingCard} ${
                        isHumanTurn && state.drawPileCount > 0 ? 'cursor-pointer' : 'cursor-default opacity-70'
                      }`}
                      style={{ background: 'none', padding: 0, border: 'none', lineHeight: 0 }}
                    >
                      <AnimatedCardBack width={cardWidth} silent />
                      <span
                        className="absolute -top-1 -right-1 min-w-[1.25rem] px-1 rounded-full bg-ds-surface text-ds-text-primary text-xs font-bold text-center leading-5 shadow"
                        aria-hidden="true"
                      >
                        {state.drawPileCount}
                      </span>
                      {hasPenalty && (
                        <span
                          className="absolute -bottom-1 -left-1 px-1.5 rounded-full bg-ds-warning text-black text-xs font-bold leading-5 shadow"
                          aria-hidden="true"
                          data-testid="prsi-stock-penalty"
                        >
                          +{state.penaltyDrawCount}
                        </span>
                      )}
                    </button>
                    <div className="text-ds-text-muted text-sm">{t('stock')}</div>
                  </div>
                </div>

                {hasPenalty && (
                  <div
                    className={`my-2 p-2 rounded text-sm font-semibold ${badgeWarningColors}`}
                    data-testid="penalty-indicator"
                  >
                    {t('penalty', { count: state.penaltyDrawCount })}
                  </div>
                )}

                {hasSkips && (
                  <div
                    className={`my-2 p-2 rounded text-sm font-semibold ${badgeWarningColors}`}
                    role="status"
                    data-testid="skip-indicator"
                  >
                    {t('pendingSkips', { count: state.pendingSkips })}
                  </div>
                )}

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

              {/* Right: info sidebar */}
              <div>
                {state.players
                  .filter((p) => !p.isHuman)
                  .map((p) => (
                    <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
                      <div className="text-ds-text-muted text-sm">
                        {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })}
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          </div>

          <GameFooter className={`${gameTheme.prsi.footer} px-4 py-2.5`}>
            {humanPlayer && (
              <div className="flex flex-wrap gap-1 mb-2" data-tutorial="prsi-player-hand">
                {humanPlayer.cards.map((card, idx) => {
                  // On the human's turn, highlight legal cards (matching suit/rank, or a 7 under
                  // penalty) and dim the rest with a reason tooltip, so the rule is visible at a glance.
                  const legal = !isHumanTurn || isPrsiLegalPlay(card, state.discardTop, state.penaltyDrawCount);
                  // **合法/非合法が title と見た目にしか無かった** (#5684)。title は
                  // スクリーンリーダーで安定して読まれる保証がなく、data-legal は
                  // アクセシビリティツリーに出ない。CUI は本文で出している。
                  const prsiCardAria = (c: Card, isLegal: boolean): string => {
                    if (!isHumanTurn) return cardAlt(c);
                    if (isLegal) return t('cardAriaPlayable', { card: cardAlt(c) });
                    // ペナルティ中は理由が変わる: スート違いではなく「7しか出せない」。
                    return state.penaltyDrawCount > 0
                      ? t('cardAriaPenalty', { card: cardAlt(c) })
                      : t('cardAriaIllegal', { card: cardAlt(c) });
                  };
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={prsiCardAria(card, legal)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      title={isHumanTurn && !legal ? t('illegalHint') : undefined}
                      data-legal={isHumanTurn ? legal : undefined}
                      className={`transition-transform ${focusRingCard} ${
                        isHumanTurn && legal ? 'rounded-lg ring-2 ring-ds-success' : ''
                      } ${isHumanTurn && !legal ? 'opacity-50' : ''}`}
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
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error} onRetry={retry} />

            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex gap-2 items-center flex-wrap">
              {isHumanTurn && (
                <div className="flex gap-2" data-tutorial="prsi-play-draw">
                  <button
                    type="button"
                    className={btnPrimary}
                    onClick={handlePlay}
                    disabled={loading || selectedCardIndices.length !== 1}
                  >
                    {t('playButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleDrawWithSound} disabled={loading}>
                    {t('drawButton')}
                  </button>
                </div>
              )}
              <GameResetButton
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="prsi-reset-button"
              />
            </div>
            <CardNavShortcutsPanel data-testid="prsi-kbd-shortcuts" />
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
