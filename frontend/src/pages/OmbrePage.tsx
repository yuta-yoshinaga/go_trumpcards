import { useEffect, useMemo, useState } from 'react';
import type { ombreApi } from '../api/gameApi';
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
import { CPU_DIFFICULTY_OPTIONS, TARGET_ROUNDS_OPTIONS, useOmbreGame } from '../hooks/useOmbreGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { OmbreResponse } from '../types/card';
import { OmbrePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { OMBRE_HELP, parseOmbreCommand } from '../utils/cli/commands/ombreCommands';
import { formatOmbreState } from '../utils/cli/formatters/ombreFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { MATADOR_NAME_KEY, matadorRank } from '../utils/ombreMatadors';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Ombre tutorial step definitions. */
const OMBRE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ombre-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ombre-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ombre-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ombre-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ombre-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const OMBRE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [OmbrePhase.BID]: 'bid',
  [OmbrePhase.PLAY]: 'play',
  [OmbrePhase.TRICK_END]: 'trickEnd',
  [OmbrePhase.ROUND_END]: 'roundEnd',
  [OmbrePhase.GAME_END]: 'gameEnd',
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

/** Renders the Ombre (Hombre) game page: a 3-player soloist-vs-coalition 40-card Spanish-deck trick-taker with a bid + trump auction. */
export const OmbrePage = withTutorial(OmbrePageContent, 'ombre', OMBRE_TUTORIAL_STEPS);

/** Inner content of the Ombre page, wrapped by TutorialProvider. */
function OmbrePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ombre');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    ombreConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useOmbreGame();

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
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('ombre');
  const ombreCliConfig: CliGameConfig<OmbreResponse, Parameters<typeof ombreApi.exec>> = useMemo(
    () => ({
      gameName: 'ombre',
      parseCommand: parseOmbreCommand,
      formatResponse: formatOmbreState,
      helpText: OMBRE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, ombreCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('ombre', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('ombre', OMBRE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="ombre" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === OmbrePhase.BID;
  const isPlayPhase = state.phase === OmbrePhase.PLAY;
  const isTrickEnd = state.phase === OmbrePhase.TRICK_END;
  const isRoundEnd = state.phase === OmbrePhase.ROUND_END;
  const isGameEnd = state.phase === OmbrePhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.isHumanBidTurn;
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
      title={tc('nav.ombre')}
      gameThemeBg={gameTheme.ombre.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/ombre"
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
                    value: ombreConfig.cpuDifficulty,
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
                    value: ombreConfig.targetRounds,
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
                  dataTutorial="ombre-trick-display"
                />
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="ombre-info">
                {/* Per-player match scores with Ombre badge */}
                <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                  {state.players.map((p) => (
                    <div key={p.id} className="py-0.5 flex items-center gap-2">
                      <span className={p.isOmbre ? 'text-ds-warning font-semibold' : ''}>
                        {playerName(p.id, p.isHuman)}: {t('score', { score: p.score })}
                      </span>
                      {p.isOmbre && (
                        <span className={`px-1.5 py-0.5 rounded text-xs ${badgeWarningColors}`}>{t('ombreBadge')}</span>
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
                    {state.ombreIdx >= 0 && (
                      <div>
                        {t('roundResult.ombre', { name: playerName(state.ombreIdx, state.ombreIdx === humanIdx) })}
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
          <GameFooter className={`${gameTheme.ombre.footer} px-4 py-2.5`}>
            {canBid && (
              <div className="mb-1 text-center text-sm text-ds-accent font-semibold" data-testid="ombre-bid-prompt">
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
                dataTutorialPrefix="ombre"
                validIndices={canPlay ? state.playableIndices : undefined}
                restrictedTooltip={t('playButton')}
                cardBadgeFor={matadorBadgeFor}
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="ombre-action-buttons">
              {canBid && pendingBid === null && (
                <div className="flex flex-wrap gap-2 items-center" data-testid="ombre-bid-stage1">
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
                <div className="flex flex-wrap gap-2 items-center" data-testid="ombre-bid-stage2">
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
                    data-testid="ombre-bid-confirm"
                  >
                    {t('confirmBid')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={cancelBid}
                    disabled={loading}
                    data-testid="ombre-bid-back"
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
                dataTutorial="ombre-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
