import { useEffect, useMemo, useState } from 'react';
import type { quadrilleApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useQuadrilleGame } from '../hooks/useQuadrilleGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { QuadrilleResponse } from '../types/card';
import { QuadrillePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseQuadrilleCommand, QUADRILLE_HELP } from '../utils/cli/commands/quadrilleCommands';
import { formatQuadrilleState } from '../utils/cli/formatters/quadrilleFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { MATADOR_NAME_KEY, matadorRank } from '../utils/quadrilleMatadors';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Quadrille tutorial step definitions. */
const QUADRILLE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="quadrille-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="quadrille-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="quadrille-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="quadrille-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="quadrille-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const QUADRILLE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [QuadrillePhase.BID]: 'bid',
  [QuadrillePhase.PLAY]: 'play',
  [QuadrillePhase.TRICK_END]: 'trickEnd',
  [QuadrillePhase.ROUND_END]: 'roundEnd',
  [QuadrillePhase.GAME_END]: 'gameEnd',
};

/** Bid labels indexed by bid value (0=pass/none, 1=entrar, 2=solo). */
const BID_KEYS = ['bidNone', 'bidEntrar', 'bidSolo'] as const;

/** Trump-suit i18n keys indexed by suit code (1=♠ 2=♣ 3=♥ 4=♦); index 0 = none. */
const SUIT_KEYS = ['suitNone', 'suitSpade', 'suitClub', 'suitHeart', 'suitDiamond'] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Sacar, 2=Puesta, 3=Codille). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeSacar', 'outcomePuesta', 'outcomeCodille'] as const;

/** Selectable trump suits with their playing-card symbols (1=♠ 2=♣ 3=♥ 4=♦). */
const TRUMP_CHOICES = [
  { code: 1, symbol: '♠' },
  { code: 2, symbol: '♣' },
  { code: 3, symbol: '♥' },
  { code: 4, symbol: '♦' },
] as const;

/**
 * Renders the Quadrille game page: a 4-player 40-card Spanish-deck trick-taker
 * with a bid + trump auction, followed by a king call that hands the winning
 * bidder a hidden partner.
 */
export const QuadrillePage = withTutorial(QuadrillePageContent, 'quadrille', QUADRILLE_TUTORIAL_STEPS);

/** Inner content of the Quadrille page, wrapped by TutorialProvider. */
function QuadrillePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('quadrille');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    quadrilleConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handleCallKing,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useQuadrilleGame();

  // Two-stage bid flow. Stage 1 picks the bid type (entrar/solo/pass); choosing
  // entrar or solo sets `pendingBid` (1 or 2) and advances to stage 2, where a
  // trump suit is picked and the declaration is confirmed. `null` = stage 1.
  const [pendingBid, setPendingBid] = useState<number | null>(null);
  // Trump suit chosen for a pending entrar/solo declaration (null until picked).
  const [selectedTrump, setSelectedTrump] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('quadrille');
  const quadrilleCliConfig: CliGameConfig<QuadrilleResponse, Parameters<typeof quadrilleApi.exec>> = useMemo(
    () => ({
      gameName: 'quadrille',
      parseCommand: parseQuadrilleCommand,
      formatResponse: formatQuadrilleState,
      helpText: QUADRILLE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, quadrilleCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('quadrille', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('quadrille', QUADRILLE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="quadrille" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === QuadrillePhase.BID;
  const isPlayPhase = state.phase === QuadrillePhase.PLAY;
  const isTrickEnd = state.phase === QuadrillePhase.TRICK_END;
  const isRoundEnd = state.phase === QuadrillePhase.ROUND_END;
  const isGameEnd = state.phase === QuadrillePhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canCallKing = state.phase === QuadrillePhase.KING_CALL && state.isHumanKingCallTurn;
  const canPlay = isPlayPhase && isHumanTurn;

  const trumpLabel = state.trumpSuit >= 1 ? t(SUIT_KEYS[state.trumpSuit] ?? 'suitNone') : t('suitNone');

  // Badge the three matadors (Spadille ♠A / Manille = trump 7 / Basto ♣A) in
  // the human's hand once trump is decided. Ring only — never blocks clicks.
  const matadorBadgeFor = (idx: number): { glyph: string; title: string } | null => {
    const card = humanPlayer?.cards[idx];
    if (!card) return null;
    const rank = matadorRank(card, state.trumpSuit);
    if (rank === null) return null;
    return { glyph: String(rank), title: t(MATADOR_NAME_KEY[rank]) };
  };

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  // Stage 1 → pass immediately, or stage into trump selection for entrar/solo.
  const chooseBid = (bid: number) => {
    if (bid === 0) {
      handleBid(0);
      return;
    }
    setPendingBid(bid);
    setSelectedTrump(null);
  };

  // Stage 2 → back to bid-type selection, discarding the pending choice.
  const cancelBid = () => {
    setPendingBid(null);
    setSelectedTrump(null);
  };

  // Stage 2 → confirm the pending declaration with the chosen trump suit.
  const confirmBid = () => {
    if (pendingBid === null || selectedTrump === null) return;
    handleBid(pendingBid, selectedTrump);
    setPendingBid(null);
    setSelectedTrump(null);
  };

  return (
    <GamePageShell
      title={tc('nav.quadrille')}
      gameThemeBg={gameTheme.quadrille.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/quadrille"
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
                    value: quadrilleConfig.cpuDifficulty,
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
                    value: quadrilleConfig.targetRounds,
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
              <span className="mr-4">{t('winningBid', { bid: t(BID_KEYS[state.winningBid] ?? 'bidNone') })}</span>
              {/* **呼ばれた王は公開情報、持ち主は伏せる。** partnerIdx は
                  その王が場に出るまでサーバが -1 で返す。 */}
              <span className="mr-4" data-testid="quadrille-king-line">
                {state.roiSeul
                  ? t('roiSeul')
                  : state.calledKingSuit >= 1
                    ? state.partnerIdx >= 0
                      ? t('kingCalledRevealed', {
                          suit: t(SUIT_KEYS[state.calledKingSuit] ?? 'suitNone'),
                          name: playerName(state.partnerIdx, state.players[state.partnerIdx]?.isHuman === true),
                        })
                      : t('kingCalledHidden', { suit: t(SUIT_KEYS[state.calledKingSuit] ?? 'suitNone') })
                    : t('kingUncalled')}
              </span>
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
                  dataTutorial="quadrille-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="quadrille-info">
                {/* Per-player match scores with Quadrille badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isQuadrille ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isQuadrille && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>
                          {t('quadrilleBadge')}
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

                {/* Round result: the deal outcome (Sacar / Puesta / Codille) */}
                {(isRoundEnd || isGameEnd) && state.outcome > 0 && (
                  <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                    <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
                    <div>{t('roundResult.outcome', { outcome: t(OUTCOME_KEYS[state.outcome] ?? 'outcomeNone') })}</div>
                    {state.quadrilleIdx >= 0 && (
                      <div>
                        {t('roundResult.quadrille', {
                          name: playerName(state.quadrilleIdx, state.quadrilleIdx === humanIdx),
                        })}
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
          <GameFooter className={`${gameTheme.quadrille.footer} px-4 py-2.5`}>
            {canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="quadrille-bid-prompt">
                {t('bidPhase')}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="quadrille"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
                cardBadgeFor={matadorBadgeFor}
              />
            )}

            <ErrorAlert message={error} onRetry={retry} />

            {/*
              ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
              領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
              られないことがある (#5955)。
            */}
            <div data-testid="quadrille-hint-live" role="status" aria-live="polite">
              {state.hint && isRequestedHint(state) && (
                <div className="text-ds-warning text-sm mb-2">
                  {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
                  {state.hint.cardIndices &&
                    state.hint.cardIndices.length > 0 &&
                    ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
                </div>
              )}
            </div>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="quadrille-action-buttons">
              {canCallKing && (
                <div className="flex flex-wrap gap-2 items-center" data-testid="quadrille-king-call">
                  <span className="text-ds-text-muted text-sm">{t('chooseKing')}:</span>
                  {state.callableKingSuits.map((suit) => (
                    <button
                      key={suit}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleCallKing(suit)}
                      disabled={loading}
                      aria-label={t(SUIT_KEYS[suit] ?? 'suitNone')}
                    >
                      {t('callKingOf', { suit: t(SUIT_KEYS[suit] ?? 'suitNone') })}
                    </button>
                  ))}
                </div>
              )}
              {canBid && pendingBid === null && (
                <div className="flex flex-wrap gap-2 items-center" data-testid="quadrille-bid-stage1">
                  <span className="text-ds-text-muted text-sm">{t('chooseBidType')}:</span>
                  <button type="button" className={btnPrimary} onClick={() => chooseBid(1)} disabled={loading}>
                    {t('bidEntrar')}
                  </button>
                  <button type="button" className={btnPrimary} onClick={() => chooseBid(2)} disabled={loading}>
                    {t('bidSolo')}
                  </button>
                  <button type="button" className={btnSecondary} onClick={() => chooseBid(0)} disabled={loading}>
                    {t('bidPass')}
                  </button>
                </div>
              )}
              {canBid && pendingBid !== null && (
                <div className="flex flex-wrap gap-2 items-center" data-testid="quadrille-bid-stage2">
                  <span className="text-ds-text-muted text-sm">
                    {t('chooseTrumpFor', { bid: t(BID_KEYS[pendingBid] ?? 'bidNone') })}:
                  </span>
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
                    className={btnSuccess}
                    onClick={confirmBid}
                    disabled={loading || selectedTrump === null}
                    data-testid="quadrille-bid-confirm"
                  >
                    {t('confirmBid')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={cancelBid}
                    disabled={loading}
                    data-testid="quadrille-bid-back"
                  >
                    {t('bidBack')}
                  </button>
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
                dataTutorial="quadrille-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
