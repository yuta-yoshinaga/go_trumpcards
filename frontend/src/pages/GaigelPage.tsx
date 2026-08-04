import { useCallback } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useGaigelGame } from '../hooks/useGaigelGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { GaigelPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Gaigel tutorial step definitions. */
const GAIGEL_TUTORIAL_STEPS: TutorialStep[] = [
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

const GAIGEL_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GaigelPhase.PLAY]: 'play',
  [GaigelPhase.TRICK_END]: 'trickEnd',
  [GaigelPhase.ROUND_END]: 'roundEnd',
  [GaigelPhase.GAME_END]: 'gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Renders the Gaigel game page (4-player / 2-team Schnapsen-family point-trick game, 48-card deck). */
export const GaigelPage = withTutorial(GaigelPageContent, 'gaigel', GAIGEL_TUTORIAL_STEPS);

function GaigelPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('gaigel');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
    gaigelConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handlePlay,
    handleMarriage,
    handleNextTrick,
    handleNextRound,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useGaigelGame();

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('gaigel', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const phaseNames = usePhaseNames('gaigel', GAIGEL_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void (
      dispatch as unknown as (command: string, a1?: number, ci?: number, cfg?: typeof gaigelConfig) => Promise<void>
    )('reset', undefined, undefined, gaigelConfig);
  }, [dispatch, hideActionLog, gaigelConfig]);

  if (!state)
    return <GameSkeleton gameKey="gaigel" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 5 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === GaigelPhase.PLAY;
  const isTrickEnd = state.phase === GaigelPhase.TRICK_END;
  const isRoundEnd = state.phase === GaigelPhase.ROUND_END;
  const isGameEnd = state.phase === GaigelPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

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
      title={tc('nav.gaigel')}
      gameThemeBg={gameTheme.gaigel.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn}
      gamePath="/gaigel"
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
                value: gaigelConfig.cpuDifficulty,
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
                value: gaigelConfig.targetScore,
                options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('targetScore', v),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* Round/Trick/Trump/Stock info */}
        <div className="text-ds-text-primary text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span className="mr-4">
            {state.trumpSuit > 0 ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) }) : t('noTrump')}
          </span>
          <span>{t('stock', { count: state.stockRemaining })}</span>
        </div>

        {/* Face-up turn-up card that fixes the trump suit. It sits under the
            stock and is drawn last, so it disappears once the stock is empty. */}
        {state.trumpCard && (
          <div className="flex items-center justify-center gap-2 mb-3" data-testid="gaigel-trump-card">
            <span className="text-ds-text-muted text-sm">{t('turnUpCard')}</span>
            <CardImage card={state.trumpCard} width={Math.round(cardWidth * 0.7)} />
          </div>
        )}

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

      <GameFooter className={`${gameTheme.gaigel.footer} px-4 py-2.5`}>
        {humanPlayer && (
          <PlayerHandSection
            humanPlayer={humanPlayer}
            selectedCardIndices={selectedCardIndices}
            toggleCard={toggleCard}
            cardWidth={cardWidth}
            isMobile={isMobile}
            dataTutorialPrefix="gg"
            cardBadgeFor={marriageBadgeFor}
          />
        )}

        <ErrorAlert message={error ?? hintError} onRetry={retry} />

        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {`${t('hintPlay')}: [${hint.cardIndex ?? '-'}] (${t(`hint.${hint.reason}`, { defaultValue: hint.reason })})`}
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
