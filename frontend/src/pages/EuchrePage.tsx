import { useCallback, useMemo, useState } from 'react';
import type { euchreApi } from '../api/gameApi';
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
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useEuchreGame } from '../hooks/useEuchreGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { EuchrePlayerData, EuchreResponse } from '../types/card';
import { EuchrePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { EUCHRE_HELP, parseEuchreCommand } from '../utils/cli/commands/euchreCommands';
import { formatEuchreState } from '../utils/cli/formatters/euchreFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { bowerRole } from '../utils/euchreBower';
import { euchreLegalPlayIndices } from '../utils/euchreLegalPlay';
import { euchreSittingOutIdx } from '../utils/euchreSittingOut';
import { playerName } from '../utils/playerUtils';

const SUIT_NAMES: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Card design string → suit number used by the call-trump buttons. */
const DESIGN_TO_SUIT: Readonly<Record<string, number>> = { SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4, JOKER: 0 };

/** Euchre tutorial step definitions. */
const EU_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="eu-pickup-controls"]',
    messageKey: 'tutorial.pickUpControls',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-team-info"]',
    messageKey: 'tutorial.teamInfo',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="eu-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const EUCHRE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [EuchrePhase.PICK_UP]: 'pickUp',
  [EuchrePhase.CALL_TRUMP]: 'callTrump',
  [EuchrePhase.DISCARD]: 'discard',
  [EuchrePhase.PLAY]: 'play',
  [EuchrePhase.TRICK_END]: 'trickEnd',
  [EuchrePhase.ROUND_END]: 'roundEnd',
  [EuchrePhase.GAME_END]: 'gameEnd',
};

/** Renders the Euchre game page with pick-up, trump calling, trick play, and team scoring. */
export const EuchrePage = withTutorial(EuchrePageContent, 'euchre', EU_TUTORIAL_STEPS);
/** Inner content of the Euchre page, wrapped by TutorialProvider. */
function EuchrePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('euchre');
  const { playSound } = useSound();
  const {
    state,
    loading,
    error,
    retry,
    apiExec,
    euchreConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleOrderUp,
    handlePass,
    handleCallTrump,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useEuchreGame();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('euchre', state);
  const { cardWidth, isMobile, solitaireMinColWidth } = useCardDimensions();
  const [goAlone, setGoAlone] = useState(false);
  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('euchre');
  const cliConfig: CliGameConfig<EuchreResponse, Parameters<typeof euchreApi.exec>> = useMemo(
    () => ({
      gameName: 'euchre',
      parseCommand: parseEuchreCommand,
      formatResponse: formatEuchreState,
      helpText: EUCHRE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(apiExec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const isPlayPhaseForKbd = state?.phase === EuchrePhase.PLAY;
  const isDiscardPhaseForKbd = state?.phase === EuchrePhase.DISCARD;
  const isHumanTurnForKbd =
    (isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true) ||
    (isDiscardPhaseForKbd && state?.players[state.dealerIdx]?.isHuman === true);
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    } else {
      handlePlay();
    }
  }, [handlePlay, handleDiscard, isDiscardPhaseForKbd]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('euchre', EUCHRE_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void apiExec('reset', undefined, undefined, undefined, {
      cpuDifficulty: euchreConfig.cpuDifficulty,
      pointLimit: euchreConfig.pointLimit,
    });
  }, [apiExec, hideActionLog, euchreConfig.cpuDifficulty, euchreConfig.pointLimit]);

  if (!state)
    return <GameSkeleton gameKey="euchre" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const sittingOutIdx = euchreSittingOutIdx(state);
  const isPickUpPhase = state.phase === EuchrePhase.PICK_UP;
  const isCallTrumpPhase = state.phase === EuchrePhase.CALL_TRUMP;
  const isDiscardPhase = state.phase === EuchrePhase.DISCARD;
  const isPlayPhase = state.phase === EuchrePhase.PLAY;
  const isTrickEnd = state.phase === EuchrePhase.TRICK_END;
  const isRoundEnd = state.phase === EuchrePhase.ROUND_END;
  const isGameEnd = state.phase === EuchrePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = (isPickUpPhase || isCallTrumpPhase) && state.players[state.bidPlayerIdx]?.isHuman === true;

  // On the human's play turn, mirror the server's follow-suit rule
  // (internal/domain/Euchre.go validatePlay) so legal cards get an additive
  // success ring. Highlight only — clicks are never blocked; the backend still
  // validates the actual play. The effective-suit check counts the left bower
  // as trump, matching the domain so the ring is intuitive.
  const legalPlayIndices =
    isHumanTurn && humanPlayer
      ? euchreLegalPlayIndices(humanPlayer.cards, state.currentTrick[0]?.card, state.trumpSuit)
      : undefined;
  const isHumanDiscard = isDiscardPhase && state.players[state.dealerIdx]?.isHuman === true;

  const suitName = (suit: number) => (SUIT_NAMES[suit] ? t(SUIT_NAMES[suit]) : '');

  return (
    <GamePageShell
      title={tc('nav.euchre')}
      gameThemeBg={gameTheme.euchre.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBidTurn || isHumanTurn || isHumanDiscard}
      gamePath="/euchre"
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
                    value: euchreConfig.cpuDifficulty,
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
                    value: euchreConfig.pointLimit,
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

          {/* Scrollable area */}
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            {/* Round/Trick info */}
            <div className="text-ds-text-primary text-center mb-2">
              <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
              <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
              {state.trumpSuit > 0 && <span>{t('trumpSuit', { suit: suitName(state.trumpSuit) })}</span>}
              {state.trumpSuit === 0 && <span>{t('noTrump')}</span>}
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: game play area */}
              <div>
                {/* Maker / Going alone info */}
                {state.makerTeam >= 0 && (
                  <div className="text-ds-warning text-center mb-2">
                    <span className="mr-4">{t('maker', { team: state.makerTeam })}</span>
                    {state.goingAlone && <span>{t('goingAlone')}</span>}
                  </div>
                )}

                {/* Pick-up phase instruction */}
                {isHumanBidTurn && isPickUpPhase && (
                  <div className="text-ds-warning text-center mb-2" data-tutorial="eu-pickup-controls">
                    {t('pickUpPhase')}
                  </div>
                )}

                {/* Call trump phase instruction */}
                {isHumanBidTurn && isCallTrumpPhase && (
                  <div className="text-ds-warning text-center mb-2">{t('callTrumpPhase')}</div>
                )}

                {/* Discard phase instruction */}
                {isHumanDiscard && <div className="text-ds-warning text-center mb-2">{t('discardPhase')}</div>}

                {/* Face-up card */}
                {state.faceUpCard && (isPickUpPhase || isCallTrumpPhase) && (
                  <div className="my-2 text-center">
                    <div className="text-ds-text-muted text-sm mb-1">{t('faceUpCard')}</div>
                    <div className="inline-block">
                      <AnimatedCard card={state.faceUpCard} width={cardWidth} />
                    </div>
                  </div>
                )}

                {/* Current trick */}
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="eu-trick-display"
                />

                {/* Partnership info */}
                {humanPlayer && (
                  <div className="text-ds-text-muted text-sm text-center mb-2" data-tutorial="eu-team-info">
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
                        .map((p) => {
                          const sittingOut = sittingOutIdx === p.id;
                          return (
                            <div
                              key={p.id}
                              data-testid={`eu-player-row-${p.id}`}
                              className={`text-ds-text-muted text-sm py-0.5 ${sittingOut ? 'opacity-40 grayscale' : ''}`}
                            >
                              <CpuPlayerLine player={p} dealerIdx={state.dealerIdx} sittingOut={sittingOut} t={t} />
                            </div>
                          );
                        })}
                    </div>
                  </details>
                ) : (
                  state.players
                    .filter((p) => !p.isHuman)
                    .map((p) => {
                      const sittingOut = sittingOutIdx === p.id;
                      return (
                        <div
                          key={p.id}
                          data-testid={`eu-player-row-${p.id}`}
                          className={`mb-2 p-2 rounded bg-black/30 ${sittingOut ? 'opacity-40 grayscale' : ''}`}
                        >
                          <div className="text-ds-text-muted text-sm">
                            <CpuPlayerLine player={p} dealerIdx={state.dealerIdx} sittingOut={sittingOut} t={t} />
                          </div>
                        </div>
                      );
                    })
                )}

                {/* Team scores */}
                {isMobile ? (
                  <details
                    className="my-3 p-2 rounded bg-black/30"
                    data-tutorial="eu-score-table"
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
                          <th scope="col">{tc('button.score')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, idx) => (
                          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                            <td className="text-center">{score}</td>
                          </tr>
                        ))}
                      </tbody>
                    </table>
                  </details>
                ) : (
                  <div className="my-3 p-2 rounded bg-black/30" data-tutorial="eu-score-table">
                    <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
                    <table className="w-full text-sm text-ds-text-muted">
                      <thead>
                        <tr>
                          <th scope="col" className="text-left">
                            {t('team', { n: '' })}
                          </th>
                          <th scope="col">{tc('button.score')}</th>
                        </tr>
                      </thead>
                      <tbody>
                        {state.teamScores.map((score, idx) => (
                          <tr key={idx} className={idx === humanTeam ? 'text-ds-accent' : ''}>
                            <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                            <td className="text-center">{score}</td>
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
          <GameFooter className={`${gameTheme.euchre.footer} px-4 py-2.5`}>
            {/* Human cards */}
            {humanPlayer && (
              <div
                className={isMobile ? 'flex gap-1 overflow-x-auto mb-2' : 'flex flex-wrap gap-1 mb-2'}
                data-tutorial="eu-player-hand"
              >
                {humanPlayer.cards.map((card, idx) => {
                  // Once trump is set, flag the two bowers so the player can read
                  // the left bower (a same-color Jack that plays as trump despite
                  // its printed suit). Badge only — never blocks card selection.
                  const role = bowerRole(card, state.trumpSuit);
                  // Additive success ring on legal-to-play cards (never blocks
                  // clicks). Uses `outline` so it stacks on top of the inline
                  // selection border/shadow instead of being clobbered by them.
                  const isLegal = legalPlayIndices?.includes(idx) ?? false;
                  return (
                    <button
                      type="button"
                      key={`${card.design}-${card.value}-${idx}`}
                      onClick={() => toggleCard(idx)}
                      aria-label={cardAlt(card)}
                      aria-pressed={selectedCardIndices.includes(idx)}
                      data-legal={isLegal ? 'true' : undefined}
                      className={`relative transition-transform ${focusRingCard}`}
                      style={{
                        background: 'none',
                        padding: 0,
                        borderRadius: 8,
                        ...selectedCardStyle(selectedCardIndices.includes(idx)),
                        ...(isLegal ? { outline: '2px solid var(--color-ds-success)', outlineOffset: '1px' } : {}),
                        boxSizing: 'border-box',
                        ...(isMobile ? { minWidth: solitaireMinColWidth, flexShrink: 0 } : {}),
                      }}
                    >
                      <AnimatedCard card={card} width={cardWidth} />
                      {role && (
                        <span
                          className="absolute top-0.5 right-0.5 px-1 rounded-full bg-ds-accent text-ds-text-on-accent text-[10px] font-bold shadow-sm pointer-events-none"
                          data-testid={`eu-bower-badge-${idx}`}
                          data-bower={role}
                          title={t(role === 'right' ? 'rightBowerTitle' : 'leftBowerTitle')}
                        >
                          {t(role === 'right' ? 'rightBowerBadge' : 'leftBowerBadge')}
                        </span>
                      )}
                    </button>
                  );
                })}
              </div>
            )}

            <ErrorAlert message={error ?? hintError} onRetry={retry} />

            {hint && (
              <div className="text-ds-warning text-sm mb-2">
                {hint.cardIndex != null
                  ? `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`
                  : `(${t(`hintReason.${hint.reason}`)})`}
              </div>
            )}
            {frontendHintEnabled && frontendHint && (
              <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
            )}

            <div className="flex gap-2 items-center flex-wrap" data-tutorial="eu-play-button">
              {(isHumanBidTurn || isHumanTurn) && (
                <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                  {tc('button.hint')}
                </button>
              )}

              {/* Pick-up phase controls */}
              {isHumanBidTurn && isPickUpPhase && (
                <>
                  <button type="button" className={btnPrimary} onClick={() => handleOrderUp(false)} disabled={loading}>
                    {t('orderUpButton')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => handleOrderUp(true)} disabled={loading}>
                    {t('orderUpAloneButton')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {/* Call trump phase controls */}
              {isHumanBidTurn && isCallTrumpPhase && (
                <>
                  {[1, 2, 3, 4].map((s) => {
                    // The turned-down card's suit cannot be called; keep its button
                    // visible but disabled so the rule is discoverable.
                    const isTurnedDownSuit = state.faceUpCard != null && s === DESIGN_TO_SUIT[state.faceUpCard.design];
                    return (
                      <button
                        key={s}
                        type="button"
                        // Red suits (♥♦) get a red border as a colour cue. A non-text ring is
                        // used instead of red label text, which on the accent button background
                        // would fail WCAG text contrast; a ring only needs the 3:1 non-text ratio.
                        className={`${btnPrimary}${s >= 3 ? ' ring-2 ring-inset ring-ds-error' : ''}`}
                        onClick={() => handleCallTrump(s, goAlone)}
                        disabled={loading || isTurnedDownSuit}
                        title={isTurnedDownSuit ? t('turnedDownSuit') : undefined}
                        data-suit={s}
                      >
                        {/* The localized name already carries the suit glyph (e.g. "♥ ハート"). */}
                        {t(SUIT_NAMES[s])}
                      </button>
                    );
                  })}
                  <label className="text-ds-text-primary text-sm flex items-center gap-1">
                    <input type="checkbox" checked={goAlone} onChange={(e) => setGoAlone(e.target.checked)} />
                    {t('goAloneCheck')}
                  </label>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {/* Discard phase */}
              {isHumanDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== 1}
                >
                  {t('discardButton')}
                </button>
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
                isGameEnd={!!isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="eu-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}

/** Translated row line + sitting-out badge for a single CPU player, shared
 * between the mobile collapsed-details and desktop card layouts. */
function CpuPlayerLine({
  player,
  dealerIdx,
  sittingOut,
  t,
}: {
  player: EuchrePlayerData;
  dealerIdx: number;
  sittingOut: boolean;
  t: (key: string, opts?: Record<string, unknown>) => string;
}) {
  return (
    <>
      {playerName(player.id, player.isHuman)}: {t('cards', { count: player.cardCount })} |{' '}
      {t('team', { n: player.team })} | {t('trickCount', { count: player.trickCount })}
      {dealerIdx === player.id ? ` | ${t('dealer')}` : ''}
      {sittingOut && (
        <span data-testid={`eu-sitting-out-${player.id}`} className="ml-2 text-ds-warning">
          💤 {t('sittingOut')}
        </span>
      )}
    </>
  );
}
