import { useEffect, useMemo } from 'react';
import type { frenchtarotApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
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
import {
  CPU_DIFFICULTY_OPTIONS,
  FRENCH_TAROT_ECART_COUNT,
  TARGET_DEALS_OPTIONS,
  useFrenchTarotGame,
} from '../hooks/useFrenchTarotGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, FrenchTarotResponse } from '../types/card';
import { FrenchTarotPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { FRENCH_TAROT_HELP, parseFrenchTarotCommand } from '../utils/cli/commands/frenchtarotCommands';
import { formatFrenchTarotState } from '../utils/cli/formatters/frenchtarotFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { frenchTarotTarget, heldBouts } from '../utils/frenchTarotBouts';
import { frenchTarotUnburiableReason } from '../utils/frenchtarotEcart';
import { playerName } from '../utils/playerUtils';

/** French Tarot tutorial step definitions. */
const FRENCH_TAROT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="frenchtarot-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="frenchtarot-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="frenchtarot-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="frenchtarot-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="frenchtarot-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const FRENCH_TAROT_PHASE_KEYS: Readonly<Record<number, string>> = {
  [FrenchTarotPhase.BID]: 'bid',
  [FrenchTarotPhase.CHIEN]: 'chien',
  [FrenchTarotPhase.PLAY]: 'play',
  [FrenchTarotPhase.TRICK_END]: 'trickEnd',
  [FrenchTarotPhase.ROUND_END]: 'roundEnd',
  [FrenchTarotPhase.GAME_END]: 'gameEnd',
};

/** Contract i18n keys indexed by contract/bid value (0=none, 1=Petite, 2=Garde, 3=Garde Sans, 4=Garde Contre). */
const CONTRACT_KEYS = [
  'contractNone',
  'contractPetite',
  'contractGarde',
  'contractGardeSans',
  'contractGardeContre',
] as const;

/** Outcome i18n keys indexed by outcome value (0=none, 1=Win/made, 2=Loss/failed). */
const OUTCOME_KEYS = ['outcomeNone', 'outcomeWin', 'outcomeLoss'] as const;

/** Selectable bid contracts with their numeric strength for the "must beat the highest bid" rule. */
const BID_CHOICES = [
  { contract: 'petite', value: 1, labelKey: 'bidPetite' },
  { contract: 'garde', value: 2, labelKey: 'bidGarde' },
  { contract: 'gardesans', value: 3, labelKey: 'bidGardeSans' },
  { contract: 'gardecontre', value: 4, labelKey: 'bidGardeContre' },
] as const;

/** True for cards that may NOT be buried in the écart: Kings (Roi, value 14) and the Excuse (gold face). */
function isUndiscardable(card: Card): boolean {
  return card.value === 14 || card.color === 'gold';
}

/** Renders the French Tarot (フレンチタロット) game page: a 4-player 78-card trick-taker with a Petite/Garde/Garde-Sans/Garde-Contre auction, a 6-card chien écart, and bouts-based deal scoring. */
export const FrenchTarotPage = withTutorial(FrenchTarotPageContent, 'frenchtarot', FRENCH_TAROT_TUTORIAL_STEPS);

/** Inner content of the French Tarot page, wrapped by TutorialProvider. */
function FrenchTarotPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('frenchtarot');
  const {
    state,
    loading,
    error,
    exec,
    retry,
    frenchtarotConfig,
    handleConfigChange,
    selectedCardIndices,
    toggleCard,
    reset,
    handleBid,
    handlePass,
    handleDiscard,
    handlePlay,
    handleNextTrick,
    handleNextRound,
  } = useFrenchTarotGame();

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    reset();
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('frenchtarot');
  const frenchtarotCliConfig: CliGameConfig<FrenchTarotResponse, Parameters<typeof frenchtarotApi.exec>> = useMemo(
    () => ({
      gameName: 'frenchtarot',
      parseCommand: parseFrenchTarotCommand,
      formatResponse: formatFrenchTarotState,
      helpText: FRENCH_TAROT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, frenchtarotCliConfig, state, { addInput, addOutput, addError, clearLog });

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('frenchtarot', state);
  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('frenchtarot', FRENCH_TAROT_PHASE_KEYS);

  if (!state)
    return (
      <GameSkeleton gameKey="frenchtarot" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 18 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.isHumanTurn;

  const isBidPhase = state.phase === FrenchTarotPhase.BID;
  const isChienPhase = state.phase === FrenchTarotPhase.CHIEN;
  const isPlayPhase = state.phase === FrenchTarotPhase.PLAY;
  const isTrickEnd = state.phase === FrenchTarotPhase.TRICK_END;
  const isRoundEnd = state.phase === FrenchTarotPhase.ROUND_END;
  const isGameEnd = state.phase === FrenchTarotPhase.GAME_END || state.gameEndFlag;

  const canBid = isBidPhase && state.isHumanBidTurn;
  const canDiscard = isChienPhase && state.isHumanDiscard;
  const canPlay = isPlayPhase && isHumanTurn;

  // Bouts (oudlers) the human currently holds in hand: the 21, the Petit (trump 1), and the Excuse.
  const heldBoutList = humanPlayer ? heldBouts(humanPlayer.cards) : [];

  const contractLabel = t(CONTRACT_KEYS[state.contract] ?? 'contractNone');
  const highestBidLabel = state.highestBid > 0 ? t(CONTRACT_KEYS[state.highestBid] ?? 'contractNone') : t('bidNone');

  // During the écart, restrict selection to buryable cards (no Kings / Excuse).
  const discardableIndices =
    canDiscard && humanPlayer
      ? humanPlayer.cards.map((c, i) => (isUndiscardable(c) ? -1 : i)).filter((i) => i >= 0)
      : undefined;

  const handValidIndices = canPlay ? state.playableIndices : canDiscard ? discardableIndices : undefined;

  // During the écart, explain per-card why an un-buriable card cannot go into the
  // chien (King / Excuse / bout / trump) via the card tooltip. Purely additive —
  // it never blocks selection; the backend still rejects illegal buries.
  const ecartTitleFor = (idx: number): string | undefined => {
    const card = humanPlayer?.cards[idx];
    if (!card) return undefined;
    const reason = frenchTarotUnburiableReason(card);
    return reason ? t(`chienUnburiable.${reason}`) : undefined;
  };

  const handleManualReset = () => {
    hideActionLog();
    reset();
  };

  return (
    <GamePageShell
      title={tc('nav.frenchtarot')}
      gameThemeBg={gameTheme.frenchtarot.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/frenchtarot"
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
                    value: frenchtarotConfig.cpuDifficulty,
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
                    value: frenchtarotConfig.targetDeals,
                    options: TARGET_DEALS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                    onSelect: (v) => handleConfigChange('targetDeals', v),
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
                  dataTutorial="frenchtarot-trick-display"
                />
                {/* Chien reveal for the human declarer during écart */}
                {isChienPhase && state.chienRevealed && state.chien.length > 0 && (
                  <div className="mt-2 p-2 rounded bg-black/30" data-testid="frenchtarot-chien">
                    <div className="text-ds-text-muted text-sm mb-1">{t('chienLabel')}</div>
                    <div className="flex flex-wrap gap-1">
                      {state.chien.map((c, i) => (
                        <CardImage
                          key={`chien-${c.design}-${c.value}-${i}`}
                          card={c}
                          width={Math.round(cardWidth * 0.7)}
                        />
                      ))}
                    </div>
                  </div>
                )}
              </div>

              {/* Right: info sidebar */}
              <div data-tutorial="frenchtarot-info">
                {/* Per-player scores with a declarer badge */}
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

                {/* Bouts (oudlers) held in the human's hand — drive the contract's point target. */}
                {humanPlayer && humanPlayer.cards.length > 0 && (
                  <div className="mb-2 p-2 rounded bg-black/30" data-testid="frenchtarot-bouts">
                    <div className="text-ds-text-muted text-sm mb-1">
                      {t('bouts.title', { count: heldBoutList.length })}
                    </div>
                    {heldBoutList.length > 0 ? (
                      <div className="flex flex-wrap gap-1">
                        {heldBoutList.map((b) => (
                          <span
                            key={b}
                            className="px-1.5 py-0.5 rounded bg-ds-warning/30 text-ds-warning text-xs"
                            data-testid={`frenchtarot-bout-${b}`}
                          >
                            {t(`bouts.${b}`)}
                          </span>
                        ))}
                      </div>
                    ) : (
                      <div className="text-ds-text-muted text-sm">{t('bouts.none')}</div>
                    )}
                    <div className="text-ds-text-muted text-xs mt-1">
                      {t('bouts.target', {
                        count: heldBoutList.length,
                        points: frenchTarotTarget(heldBoutList.length),
                      })}
                    </div>
                  </div>
                )}

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
                    data-testid="frenchtarot-result"
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
                        <div>{t('roundResult.contract', { contract: contractLabel })}</div>
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
          <GameFooter className={`${gameTheme.frenchtarot.footer} px-4 py-2.5`}>
            {canBid && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="frenchtarot-bid-prompt"
              >
                {t('bidPhase', { bid: highestBidLabel })}
              </div>
            )}
            {canDiscard && (
              <div
                className="mb-1 text-center text-sm text-ds-accent font-semibold"
                data-testid="frenchtarot-discard-prompt"
              >
                {t('chienPhase', { count: selectedCardIndices.length })}
              </div>
            )}
            {humanPlayer && (
              <PlayerHandSection
                humanPlayer={humanPlayer}
                selectedCardIndices={selectedCardIndices}
                toggleCard={toggleCard}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="frenchtarot"
                validIndices={handValidIndices}
                restrictedTooltip={canDiscard ? t('chienRestricted') : t('playButton')}
                cardTitleFor={canDiscard ? ecartTitleFor : undefined}
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

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="frenchtarot-action-buttons">
              {canBid && (
                <>
                  <button type="button" className={btnSecondary} onClick={handlePass} disabled={loading}>
                    {t('bidPass')}
                  </button>
                  {BID_CHOICES.map((c) => (
                    <button
                      key={c.contract}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBid(c.contract)}
                      disabled={loading || c.value <= state.highestBid}
                    >
                      {t(c.labelKey)}
                    </button>
                  ))}
                </>
              )}
              {canDiscard && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={handleDiscard}
                  disabled={loading || selectedCardIndices.length !== FRENCH_TAROT_ECART_COUNT}
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
                dataTutorial="frenchtarot-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
