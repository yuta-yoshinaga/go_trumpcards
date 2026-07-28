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
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useJassGame } from '../hooks/useJassGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { JassPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Jass (Schieber) tutorial step definitions. */
const JASS_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ja-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ja-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ja-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ja-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ja-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ja-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const JASS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [JassPhase.BID_TRUMP]: 'bidTrump',
  [JassPhase.BID_PARTNER]: 'bidPartner',
  [JassPhase.PLAY]: 'play',
  [JassPhase.TRICK_END]: 'trickEnd',
  [JassPhase.ROUND_END]: 'roundEnd',
  [JassPhase.GAME_END]: 'gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

const ALL_SUITS = [1, 2, 3, 4];

/** Renders the Jass (Schieber) game page (4-player partnership trick-taking, 36-card deck). */
export const JassPage = withTutorial(JassPageContent, 'jass', JASS_TUTORIAL_STEPS);

function JassPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('jass');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
    jassConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handleToggle,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleCallTrump,
    handleSchieben,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useJassGame();

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('jass', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const phaseNames = usePhaseNames('jass', JASS_PHASE_KEYS);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void (dispatch as unknown as (command: string, s?: number, ci?: number, cfg?: typeof jassConfig) => Promise<void>)(
      'reset',
      undefined,
      undefined,
      jassConfig,
    );
  }, [dispatch, hideActionLog, jassConfig]);

  if (!state)
    return <GameSkeleton gameKey="jass" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 9 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidTrump = state.phase === JassPhase.BID_TRUMP;
  const isBidPartner = state.phase === JassPhase.BID_PARTNER;
  const isPlayPhase = state.phase === JassPhase.PLAY;
  const isTrickEnd = state.phase === JassPhase.TRICK_END;
  const isRoundEnd = state.phase === JassPhase.ROUND_END;
  const isGameEnd = state.phase === JassPhase.GAME_END || state.gameEndFlag;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanBidTurn = (isBidTrump || isBidPartner) && state.bidPlayerIdx === humanIdx;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <GamePageShell
      title={tc('nav.jass')}
      gameThemeBg={gameTheme.jass.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isHumanBidTurn}
      gamePath="/jass"
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
                value: jassConfig.cpuDifficulty,
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
                value: jassConfig.targetScore,
                options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('targetScore', v),
              },
              {
                type: 'checkbox',
                id: 'enableWeis',
                label: t('settings.enableWeis'),
                checked: jassConfig.enableWeis,
                onToggle: (v) => handleToggle('enableWeis', v),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* Round/Trick/Trump info */}
        <div className="text-ds-text-primary text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span>
            {state.trumpSuit > 0 ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) }) : t('noTrump')}
          </span>
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
          dataTutorial="ja-trick-display"
        />

        {/* Previous trick reviewer: lets the player recount the just-completed trick + winner */}
        <details className="mb-2 p-2 rounded bg-black/30" data-testid="ja-previous-trick">
          <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('previousTrick')}</summary>
          <div className="mt-1">
            {state.lastTrick.length > 0 ? (
              <TrickDisplay
                currentTrick={state.lastTrick}
                players={state.players}
                cardWidth={Math.round(cardWidth * 0.7)}
                label={
                  state.lastTrickWinner >= 0
                    ? t('previousTrickWinner', {
                        name: playerName(state.lastTrickWinner, state.players[state.lastTrickWinner]?.isHuman === true),
                      })
                    : t('previousTrick')
                }
                winnerIdx={state.lastTrickWinner >= 0 ? state.lastTrickWinner : undefined}
              />
            ) : (
              <div className="text-ds-text-muted text-sm">{t('previousTrickEmpty')}</div>
            )}
          </div>
        </details>

        {/* Team scores */}
        <div className="my-3 p-2 rounded bg-black/30" data-tutorial="ja-score-table">
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
              {state.config.enableWeis && (state.roundWeisPoints[0] > 0 || state.roundWeisPoints[1] > 0) && (
                <tr>
                  <td className="text-xs text-ds-warning">{t('weisPoints', { points: state.roundWeisPoints[0] })}</td>
                  <td className="text-center text-xs text-ds-warning">
                    {t('weisPoints', { points: state.roundWeisPoints[1] })}
                  </td>
                </tr>
              )}
              {(state.roundStockPoints[0] > 0 || state.roundStockPoints[1] > 0) && (
                <tr>
                  <td className="text-xs">{t('stockPoints', { points: state.roundStockPoints[0] })}</td>
                  <td className="text-center text-xs">{t('stockPoints', { points: state.roundStockPoints[1] })}</td>
                </tr>
              )}
            </tbody>
          </table>
        </div>

        {/* Weis (meld) declaration panel — surfaces where the Weis bonus came from.
            Only per-team totals are exposed by the API, so we visualize those faithfully:
            each team's declared Weis total, a "counted" marker for the scoring team, and a
            note explaining the meld mechanic. */}
        {state.config.enableWeis && (state.roundWeisPoints[0] > 0 || state.roundWeisPoints[1] > 0) && (
          <section
            className="my-3 p-3 rounded bg-black/30 border border-ds-warning/40"
            aria-label={t('weisPanel.title')}
            data-testid="jass-weis-panel"
          >
            <h3 className="text-ds-warning text-sm font-semibold mb-2">{t('weisPanel.title')}</h3>
            <ul className="flex flex-col gap-1">
              {[0, 1].map((team) => (
                <li key={team} className="flex items-center gap-2 text-sm text-ds-text-muted">
                  <span>
                    {t('team', { n: team })}
                    {humanPlayer?.team === team ? t('weisPanel.you') : ''}
                  </span>
                  <span className="text-ds-warning font-medium">
                    {t('weisPanel.teamPoints', { points: state.roundWeisPoints[team] })}
                  </span>
                  {state.roundWeisPoints[team] > 0 && (
                    <span className={`text-xs px-1.5 py-0.5 rounded ${badgeWarningColors}`}>
                      {t('weisPanel.scored')}
                    </span>
                  )}
                </li>
              ))}
            </ul>
            <p className="mt-2 text-xs text-ds-text-muted">{t('weisPanel.note')}</p>
          </section>
        )}

        <RoundScoreAnnouncement
          active={isRoundEnd || isGameEnd}
          entries={[
            {
              name: t('team', { n: 0 }),
              roundScore: state.roundPoints[0] + state.roundWeisPoints[0] + state.roundStockPoints[0],
              cumulativeScore: state.teamScores[0],
            },
            {
              name: t('team', { n: 1 }),
              roundScore: state.roundPoints[1] + state.roundWeisPoints[1] + state.roundStockPoints[1],
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

      <GameFooter className={`${gameTheme.jass.footer} px-4 py-2.5`}>
        {humanPlayer && (
          <PlayerHandSection
            humanPlayer={humanPlayer}
            selectedCardIndices={selectedCardIndices}
            toggleCard={toggleCard}
            cardWidth={cardWidth}
            isMobile={isMobile}
            dataTutorialPrefix="ja"
          />
        )}

        <ErrorAlert message={error ?? hintError} onRetry={retry} />

        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {/* hint.reason is a raw backend identifier; translate via hintReason.*,
                falling back to a generic label for any unknown reason. */}
            {`${t('hintPlay')}: [${hint.cardIndex ?? '-'}] (${t(`hintReason.${hint.reason}`, {
              defaultValue: t('hintReason.fallback'),
            })})`}
          </div>
        )}

        <div className="flex gap-2 items-center flex-wrap" data-tutorial="ja-play-button">
          {isHumanBidTurn && (
            <span data-tutorial="ja-bid-controls" className="flex gap-2 flex-wrap">
              {ALL_SUITS.map((s) => (
                <button
                  key={s}
                  type="button"
                  className={btnPrimary}
                  onClick={() => handleCallTrump(s)}
                  disabled={loading}
                >
                  {t(SUIT_LABEL_KEYS[s])}
                </button>
              ))}
              {isBidTrump && (
                <button type="button" className={btnSuccess} onClick={handleSchieben} disabled={loading}>
                  {t('schiebenButton')}
                </button>
              )}
            </span>
          )}

          {isHumanTurn && (
            <>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
                {tc('button.hint')}
              </button>
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
            dataTutorial="ja-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
