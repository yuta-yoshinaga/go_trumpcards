import { useEffect, useMemo } from 'react';
import type { koenigrufenApi } from '../api/gameApi';
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
import {
  CPU_DIFFICULTY_OPTIONS,
  KOENIGRUFEN_DISCARD_COUNT,
  TARGET_DEALS_OPTIONS,
  useKoenigrufenGame,
} from '../hooks/useKoenigrufenGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, KoenigrufenResponse } from '../types/card';
import { KoenigrufenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { KOENIGRUFEN_HELP, parseKoenigrufenCommand } from '../utils/cli/commands/koenigrufenCommands';
import { formatKoenigrufenState } from '../utils/cli/formatters/koenigrufenFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Königrufen tutorial step definitions. */
const KOENIGRUFEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="koenigrufen-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="koenigrufen-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="koenigrufen-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="koenigrufen-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="koenigrufen-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const KOENIGRUFEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [KoenigrufenPhase.BID]: 'bid',
  [KoenigrufenPhase.CALL]: 'call',
  [KoenigrufenPhase.TALON]: 'talon',
  [KoenigrufenPhase.PLAY]: 'play',
  [KoenigrufenPhase.TRICK_END]: 'trickEnd',
  [KoenigrufenPhase.ROUND_END]: 'roundEnd',
  [KoenigrufenPhase.GAME_END]: 'gameEnd',
};

/** Contract i18n keys indexed by contract/bid value (0=none, 1=Rufer). */
const CONTRACT_KEYS = ['contractNone', 'contractRufer'] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Win/made, 2=Loss/failed). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** The four callable King suits, mapped to the backend callSuit index (1=♠ 2=♣ 3=♥ 4=♦). */
const CALL_SUITS = [
  { suit: 1, design: 'SPADE', glyph: '♠', labelKey: 'suitSpade' },
  { suit: 2, design: 'CLOVER', glyph: '♣', labelKey: 'suitClover' },
  { suit: 3, design: 'HEART', glyph: '♥', labelKey: 'suitHeart' },
  { suit: 4, design: 'DIAMOND', glyph: '♦', labelKey: 'suitDiamond' },
] as const;

/** Suit i18n label keys indexed by the called-King suit value (0=none, 1=♠ 2=♣ 3=♥ 4=♦). */
const SUIT_KEYS = ['-', 'suitSpade', 'suitClover', 'suitHeart', 'suitDiamond'] as const;
const SUIT_GLYPHS = ['', '♠', '♣', '♥', '♦'] as const;

/** True when the card is a plain-suit King (value 8), which can never be buried. */
function isKing(card: Card): boolean {
  return (
    card.value === 8 &&
    (card.design === 'SPADE' || card.design === 'CLOVER' || card.design === 'HEART' || card.design === 'DIAMOND')
  );
}

/** True when the card is a Trull honour (Sküs, Pagat=trump 1, or XXI=trump 21), which can never be buried. */
function isTrull(card: Card): boolean {
  return card.color === 'gold' || (card.color === 'purple' && (card.value === 1 || card.value === 21));
}

/** True when the card is an ordinary trump (a tarock that is not a Trull honour). Trumps are buryable only when unavoidable. */
function isPlainTrump(card: Card): boolean {
  return card.color === 'purple' && card.value !== 1 && card.value !== 21;
}

/** True for an ordinary suit card (not a King, not a Trull honour, not a trump) — always buryable. */
function isPlainBuryable(card: Card): boolean {
  return !isKing(card) && !isTrull(card) && !isPlainTrump(card);
}

/**
 * Returns a predicate for cards that may NOT be buried in the talon exchange,
 * mirroring the backend rule: Kings and the Trull honours are never buryable,
 * and ordinary trumps are buryable only when there are fewer than 6 ordinary
 * suit cards to bury.
 */
function makeUndiscardable(hand: Card[]): (card: Card) => boolean {
  const allowTrump = hand.filter(isPlainBuryable).length < KOENIGRUFEN_DISCARD_COUNT;
  return (card: Card) => isKing(card) || isTrull(card) || (isPlainTrump(card) && !allowTrump);
}

/** Renders the Königrufen (ケーニッヒルーフェン) game page: a 4-player 54-card tarock trick-taker with a Rufer auction, a King-calling phase that names a secret partner, a 6-card talon exchange, and captured-points deal scoring. */
export const KoenigrufenPage = withTutorial(KoenigrufenPageContent, 'koenigrufen', KOENIGRUFEN_TUTORIAL_STEPS);

/** Inner content of the Königrufen page, wrapped by TutorialProvider. */
function KoenigrufenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('koenigrufen');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    koenigrufenConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePass,
    handleCallKing,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useKoenigrufenGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('koenigrufen');
  const koenigrufenCliConfig: CliGameConfig<KoenigrufenResponse, Parameters<typeof koenigrufenApi.exec>> = useMemo(
    () => ({
      gameName: 'koenigrufen',
      parseCommand: parseKoenigrufenCommand,
      formatResponse: formatKoenigrufenState,
      helpText: KOENIGRUFEN_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, koenigrufenCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('koenigrufen', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('koenigrufen', KOENIGRUFEN_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton gameKey="koenigrufen" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 12 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === KoenigrufenPhase.BID;
  const isCallPhase = state.phase === KoenigrufenPhase.CALL;
  const isTalonPhase = state.phase === KoenigrufenPhase.TALON;
  const isPlayPhase = state.phase === KoenigrufenPhase.PLAY;
  const isTrickEnd = state.phase === KoenigrufenPhase.TRICK_END;
  const isRoundEnd = state.phase === KoenigrufenPhase.ROUND_END;
  const isGameEnd = state.phase === KoenigrufenPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canCall = isCallPhase && state.isHumanCall;
  const canDiscard = isTalonPhase && state.isHumanDiscard;
  const canPlay = isPlayPhase && isHumanTurn;

  const contractLabel = t(CONTRACT_KEYS[state.contract] ?? 'contractNone');

  // Called-King clue derivation. Uses only information the human legitimately holds:
  // the public called suit and the human's own hand — never the hidden partnerIdx.
  const calledSuitLabel =
    state.calledKing >= 1
      ? `${SUIT_GLYPHS[state.calledKing] ?? ''} ${t(SUIT_KEYS[state.calledKing] ?? '-')}`.trim()
      : '';
  // The human knows they are the secret partner when they hold the called King themselves
  // (the frontend never lets the human declarer call a King they hold, so this is unambiguous).
  const humanHoldsCalledKing =
    state.calledKing >= 1 &&
    !humanPlayer?.isDeclarer &&
    !!humanPlayer?.cards.some(
      (c) => isKing(c) && CALL_SUITS.find((s) => s.design === c.design)?.suit === state.calledKing,
    );

  // King suits the human declarer already holds cannot be called.
  const heldKingSuits = new Set<number>(
    humanPlayer
      ? humanPlayer.cards.filter(isKing).map((c) => CALL_SUITS.find((s) => s.design === c.design)?.suit ?? 0)
      : [],
  );

  // During the talon exchange, restrict selection to buryable cards (no Kings / Trull honours;
  // ordinary trumps only when there are too few plain suit cards to bury).
  const discardableIndices =
    canDiscard && humanPlayer
      ? (() => {
          const undiscardable = makeUndiscardable(humanPlayer.cards);
          return humanPlayer.cards.map((c, i) => (undiscardable(c) ? -1 : i)).filter((i) => i >= 0);
        })()
      : undefined;

  const handValidIndices = canPlay ? state.playableIndices : canDiscard ? discardableIndices : undefined;

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.koenigrufen')}
      gameThemeBg={gameTheme.koenigrufen.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/koenigrufen"
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
                    value: koenigrufenConfig.cpuDifficulty,
                    options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                      value: o.value,
                      label: t(`settings.${o.label.toLowerCase()}`),
                    })),
                    onSelect: (v) => handleConfigChange('cpuDifficulty', v),
                  },
                  {
                    type: 'select',
                    id: 'targetDeals',
                    label: t('settings.targetDeals'),
                    value: koenigrufenConfig.targetDeals,
                    options: TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetDeals', v),
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
              <span>{t('contract', { contract: contractLabel })}</span>
            </div>

            <div className={lgTwoColGrid}>
              {/* Left: play area */}
              <div>
                <TrickDisplay
                  currentTrick={state.currentTrick}
                  players={state.players}
                  cardWidth={cardWidth}
                  label={t('currentTrick')}
                  dataTutorial="koenigrufen-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="koenigrufen-info">
                {/* Called King (public once named) + secret-partner clues */}
                {state.calledKing >= 1 && (
                  <div
                    className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
                    data-testid="koenigrufen-called-king"
                  >
                    {t('calledKing', { suit: calledSuitLabel })}
                    {state.partnerRevealed && state.partnerIdx >= 0 ? (
                      <span className="ml-2">
                        {t('partnerRevealed', {
                          name: playerName(state.partnerIdx, state.partnerIdx === humanIdx),
                        })}
                      </span>
                    ) : (
                      // Partner not yet revealed: organize the clues the human legitimately has.
                      <div className="mt-1" data-testid="koenigrufen-partner-clue">
                        <div>{t('partnerClue.unknown', { suit: calledSuitLabel })}</div>
                        {humanHoldsCalledKing && (
                          <div className="text-ds-warning font-semibold" data-testid="koenigrufen-partner-clue-you">
                            {t('partnerClue.youArePartner')}
                          </div>
                        )}
                      </div>
                    )}
                  </div>
                )}

                {/* Per-player scores with declarer / partner badges */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => {
                    const isTeam = p.isDeclarer || (state.partnerRevealed && p.isPartner);
                    return (
                      <div key={p.id} className="py-0.5 flex items-center gap-2">
                        <span className={isTeam ? 'text-ds-warning font-semibold' : ''}>
                          {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                        </span>
                        {p.isDeclarer && (
                          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                            {t('declarerBadge')}
                          </span>
                        )}
                        {state.partnerRevealed && p.isPartner && !p.isDeclarer && (
                          <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                            {t('partnerBadge')}
                          </span>
                        )}
                      </div>
                    );
                  })}
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
                    data-testid="koenigrufen-result"
                  >
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
                    {state.declarerIdx >= 0 && (
                      <>
                        <div>
                          {t('roundResult.declarer', {
                            name: playerName(state.declarerIdx, state.declarerIdx === humanIdx),
                          })}
                        </div>
                        <div>
                          {t('roundResult.captured', {
                            points: state.players[state.declarerIdx]?.cardPoints ?? 0,
                          })}
                        </div>
                      </>
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
          <GameFooter className={`${gameTheme.koenigrufen.footer} px-4 py-2.5`}>
            {canBid && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="koenigrufen-bid-prompt"
              >
                {t('bidPhase')}
              </div>
            )}
            {canCall && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="koenigrufen-call-prompt"
              >
                {t('callPhase')}
              </div>
            )}
            {canDiscard && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="koenigrufen-discard-prompt"
              >
                {t('talonPhase', { count: selectedCardIndices.length })}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="koenigrufen"
                validIndices={handValidIndices}
                restrictedTooltip={canDiscard ? t('talonRestricted') : t('playButton')}
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="koenigrufen-action-buttons">
              {canBid && (
                <>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('bidPass')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={handleBid} disabled={loading}>
                    {t('bidRufer')}
                  </button>
                </>
              )}
              {canCall &&
                CALL_SUITS.map((s) => {
                  const held = heldKingSuits.has(s.suit);
                  const reason = held ? t('callKingHeldReason') : undefined;
                  // Wrap in a span so the explanatory tooltip still shows on the disabled button.
                  return (
                    <span key={s.suit} title={reason} className="inline-flex">
                      <button
                        type="button"
                        className={`${btnPrimary} ${held ? 'line-through' : ''}`}
                        onClick={() => handleCallKing(s.suit)}
                        disabled={loading || held}
                        aria-label={reason ? `${t(s.labelKey)} — ${reason}` : t(s.labelKey)}
                        data-testid={`call-king-${s.suit}`}
                      >
                        {s.glyph} {t(s.labelKey)}
                      </button>
                    </span>
                  );
                })}
              {canDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== KOENIGRUFEN_DISCARD_COUNT}
                >
                  {t('discardButton', { count: selectedCardIndices.length })}
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
                dataTutorial="koenigrufen-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
