import { useCallback } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useBauernschnapsenGame } from '../hooks/useBauernschnapsenGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { BauernschnapsenPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Bauernschnapsen tutorial step definitions. */
const BAUERNSCHNAPSEN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="gg-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gg-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BAUERNSCHNAPSEN_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BauernschnapsenPhase.CONTRACT]: 'contract',
  [BauernschnapsenPhase.PLAY]: 'play',
  [BauernschnapsenPhase.TRICK_END]: 'trickEnd',
  [BauernschnapsenPhase.ROUND_END]: 'roundEnd',
  [BauernschnapsenPhase.GAME_END]: 'gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Suits a trump-taking contract can name, in the order the buttons appear. */
const TRUMP_SUITS = [1, 2, 3, 4] as const;

/** Contract labels, keyed by the domain's BauernschnapsenContract values. */
const CONTRACT_LABEL_KEYS: Readonly<Record<number, string>> = {
  0: 'contractPass',
  1: 'contractRufer',
  2: 'contractFarbenzwang',
  3: 'contractBettel',
};

/** Contract numbers that require a trump suit alongside the declaration. */
const TRUMP_CONTRACTS = [1, 2] as const;

/**
 * Renders the Bauernschnapsen game page (4-player / 2-team Schnapsen-family
 * point-trick game, 20-card deck, bid contract, must-follow from trick one).
 */
export const BauernschnapsenPage = withTutorial(
  BauernschnapsenPageContent,
  'bauernschnapsen',
  BAUERNSCHNAPSEN_TUTORIAL_STEPS,
);

function BauernschnapsenPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bauernschnapsen');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
    bauernschnapsenConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handlePlay,
    handleMarriage,
    handleContract,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useBauernschnapsenGame();

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bauernschnapsen', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const phaseNames = usePhaseNames('bauernschnapsen', BAUERNSCHNAPSEN_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void (
      dispatch as unknown as (
        command: string,
        a1?: number,
        ci?: number,
        cfg?: typeof bauernschnapsenConfig,
      ) => Promise<void>
    )('reset', undefined, undefined, bauernschnapsenConfig);
  }, [dispatch, hideActionLog, bauernschnapsenConfig]);

  if (!state)
    return (
      <GameSkeleton gameKey="bauernschnapsen" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />
    );

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isContractPhase = state.phase === BauernschnapsenPhase.CONTRACT;
  const isPlayPhase = state.phase === BauernschnapsenPhase.PLAY;
  const isTrickEnd = state.phase === BauernschnapsenPhase.TRICK_END;
  const isRoundEnd = state.phase === BauernschnapsenPhase.ROUND_END;
  const isGameEnd = state.phase === BauernschnapsenPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanContractTurn = isContractPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  const selectedIdx = selectedCardIndices.length === 1 ? selectedCardIndices[0] : -1;
  const canDeclareMarriage = isHumanTurn && selectedIdx >= 0 && state.marriageIndices.includes(selectedIdx);

  // Surface a 💍 badge on hand cards that can start a marriage (King + Queen of
  // the same suit, both currently held) so the 20/40-point opportunity is
  // visible without probing each card. `marriageIndices` is scoped to the
  // current player, so restrict it to the human's own lead turn.
  const marriageIndices = isHumanTurn ? state.marriageIndices : [];
  const marriageBadgeFor = (idx: number): { glyph: string; title: string } | null =>
    marriageIndices.includes(idx) ? { glyph: '💍', title: t('marriageBadge') } : null;

  return (
    <GamePageShell
      title={tc('nav.bauernschnapsen')}
      gameThemeBg={gameTheme.bauernschnapsen.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/bauernschnapsen"
      gameEndFlag={!!state?.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: bauernschnapsenConfig.cpuDifficulty,
                options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(`settings.${o.label.toLowerCase()}`),
                })),
                onSelect: (v) => handleConfigChange('cpuDifficulty', v),
              },
              {
                type: 'select',
                id: 'targetScore',
                label: t('settings.targetScore'),
                value: bauernschnapsenConfig.targetScore,
                options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('targetScore', v),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* Round / trick / trump / contract. There is no stock and no turn-up
            card in this game: all 20 cards are dealt, and the trump comes from
            the winning declaration rather than from a face-up card. */}
        <div className="text-ds-text-primary text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span>
            {state.trumpSuit > 0 ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) }) : t('noTrump')}
          </span>
        </div>

        <div className="text-ds-text-muted text-center text-sm mb-3" data-testid="bauernschnapsen-contract">
          {isContractPhase
            ? t('contractPending')
            : t('contractLine', {
                contract: t(CONTRACT_LABEL_KEYS[state.contract] ?? 'contractRufer'),
                name: playerName(state.declarerIdx, state.players[state.declarerIdx]?.isHuman === true),
              })}
        </div>

        {/* CPU players */}
        <div className="mb-3">
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="mb-1 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                {playerName(p.id, p.isHuman)}: {t('team', { n: p.team })} | {t('cards', { count: p.cardCount })} |{' '}
                {t('trickCount', { count: p.trickCount })}
              </div>
            ))}
        </div>

        {/* Current trick */}
        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players}
          cardWidth={cardWidth}
          label={t('currentTrick')}
          dataTutorial="gg-trick-display"
        />

        {/* Team scores */}
        <div className="my-3 p-2 rounded bg-black/30" data-tutorial="gg-score-table">
          <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
          <table className="w-full text-sm text-ds-text-muted">
            <thead>
              <tr>
                <th scope="col" className="text-left">
                  {t('team', { n: 0 })}
                </th>
                <th scope="col" className="text-center">
                  {t('team', { n: 1 })}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="text-ds-accent">{state.teamScores[0]}</td>
                <td className="text-center">{state.teamScores[1]}</td>
              </tr>
              <tr>
                <td className="text-xs">{t('roundPoints', { points: state.roundPoints[0] })}</td>
                <td className="text-center text-xs">{t('roundPoints', { points: state.roundPoints[1] })}</td>
              </tr>
              {(state.roundMarriage[0] > 0 || state.roundMarriage[1] > 0) && (
                <tr>
                  <td className="text-xs text-ds-warning">{t('marriagePoints', { points: state.roundMarriage[0] })}</td>
                  <td className="text-center text-xs text-ds-warning">
                    {t('marriagePoints', { points: state.roundMarriage[1] })}
                  </td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        <RoundScoreAnnouncement
          active={isRoundEnd || isGameEnd}
          entries={[
            {
              name: t('team', { n: 0 }),
              roundScore: state.roundPoints[0] + state.roundMarriage[0],
              cumulativeScore: state.teamScores[0],
            },
            {
              name: t('team', { n: 1 }),
              roundScore: state.roundPoints[1] + state.roundMarriage[1],
              cumulativeScore: state.teamScores[1],
            },
          ]}
        />

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.bauernschnapsen.footer} px-4 py-2.5`}>
        {humanPlayer && (
          <PlayerHandSection
            humanPlayer={humanPlayer}
            selectedCardIndices={selectedCardIndices}
            toggleCard={toggleCard}
            cardWidth={cardWidth}
            isMobile={isMobile}
            dataTutorialPrefix="gg"
            cardBadgeFor={marriageBadgeFor}
            validIndices={isHumanTurn ? state.validPlayIndices : undefined}
            restrictedTooltip={t('mustFollow')}
          />
        )}

        <ErrorAlert message={error ?? hintError} onRetry={retry} />

        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {`${t('hintPlay')}: [${hint.cardIndex ?? '-'}] (${t(`hintReason.${hint.reason}`, { defaultValue: hint.reason })})`}
          </div>
        )}

        {isHumanContractTurn && (
          <div className="mb-2" data-testid="bauernschnapsen-contract-controls">
            <div className="text-ds-text-muted text-sm mb-1">{t('declareContract')}</div>
            <div className="flex gap-2 items-center flex-wrap">
              <button
                type="button"
                className={btnPrimary}
                onClick={() => handleContract(0, 1)}
                disabled={loading}
                title={t('contractPassHelp')}
              >
                {t('contractPass')}
              </button>
              {TRUMP_CONTRACTS.map((c) =>
                TRUMP_SUITS.map((suit) => (
                  <button
                    key={`${c}-${suit}`}
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleContract(c, suit)}
                    disabled={loading}
                    title={t(`${CONTRACT_LABEL_KEYS[c]}Help`)}
                  >
                    {`${t(CONTRACT_LABEL_KEYS[c])} ${t(SUIT_LABEL_KEYS[suit])}`}
                  </button>
                )),
              )}
              <button
                type="button"
                className={btnPrimary}
                onClick={() => handleContract(3, 1)}
                disabled={loading}
                title={t('contractBettelHelp')}
              >
                {t('contractBettel')}
              </button>
            </div>
          </div>
        )}

        <div className="flex gap-2 items-center flex-wrap" data-tutorial="gg-play-button">
          {isHumanTurn && (
            <>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                {tc('button.hint')}
              </button>
              {canDeclareMarriage && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => handleMarriage(selectedIdx)}
                  disabled={loading}
                  title={t('declareMarriage')}
                >
                  {t('marriageButton')}
                </button>
              )}
              <button
                type="button"
                className={btnPrimary}
                onClick={handlePlay}
                disabled={loading || selectedCardIndices.length !== 1}
              >
                {t('playButton')}
              </button>
            </>
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
            isGameEnd={!!isGameEnd}
            onReset={handleManualReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="gg-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
