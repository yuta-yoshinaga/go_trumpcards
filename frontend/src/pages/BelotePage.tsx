import { useCallback, useEffect, useRef, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { HintTooltip } from '../components/hint/HintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useBeloteGame } from '../hooks/useBeloteGame';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { BelotePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { beloteLegalPlayIndices } from '../utils/beloteLegal';
import { playerName } from '../utils/playerUtils';

/** Belote tutorial step definitions. */
const BELOTE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="be-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BELOTE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BelotePhase.BID_PICK_UP]: 'phase.bidPickUp',
  [BelotePhase.BID_CALL_TRUMP]: 'phase.bidCallTrump',
  [BelotePhase.PLAY]: 'phase.play',
  [BelotePhase.TRICK_END]: 'phase.trickEnd',
  [BelotePhase.ROUND_END]: 'phase.roundEnd',
  [BelotePhase.GAME_END]: 'phase.gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Renders the Belote game page (4-player partnership trick-taking, 32-card deck). */
export const BelotePage = withTutorial(BelotePageContent, 'belote', BELOTE_TUTORIAL_STEPS);

function BelotePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('belote');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
    beloteConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handleToggle,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleOrderUp,
    handlePass,
    handleCallTrump,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useBeloteGame();

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('belote', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const phaseNames = usePhaseNames('belote', BELOTE_PHASE_KEYS);

  const { playSound } = useSound();
  const beloteTotal = state ? state.roundBeloteBonus[0] + state.roundBeloteBonus[1] : 0;
  const prevBeloteTotalRef = useRef<number | null>(null);
  const [beloteJustConfirmed, setBeloteJustConfirmed] = useState(false);
  useEffect(() => {
    if (!state) return;
    if (prevBeloteTotalRef.current === null) {
      prevBeloteTotalRef.current = beloteTotal;
      return;
    }
    if (beloteTotal > prevBeloteTotalRef.current) {
      setBeloteJustConfirmed(true);
      playSound('winFanfare');
    }
    prevBeloteTotalRef.current = beloteTotal;
  }, [beloteTotal, playSound, state]);

  // Clear timer keyed only on the flag so an unrelated state update mid-window can't cancel it.
  useEffect(() => {
    if (!beloteJustConfirmed) return;
    const id = setTimeout(() => setBeloteJustConfirmed(false), 2500);
    return () => clearTimeout(id);
  }, [beloteJustConfirmed]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void (
      dispatch as unknown as (command: string, ci?: number, s?: number, cfg?: typeof beloteConfig) => Promise<void>
    )('reset', undefined, undefined, beloteConfig);
  }, [dispatch, hideActionLog, beloteConfig]);

  if (!state)
    return <GameSkeleton gameKey="belote" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPickUp = state.phase === BelotePhase.BID_PICK_UP;
  const isBidCallTrump = state.phase === BelotePhase.BID_CALL_TRUMP;
  const isPlayPhase = state.phase === BelotePhase.PLAY;
  const isTrickEnd = state.phase === BelotePhase.TRICK_END;
  const isRoundEnd = state.phase === BelotePhase.ROUND_END;
  const isGameEnd = state.phase === BelotePhase.GAME_END || state.gameEndFlag;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanBidTurn = (isBidPickUp || isBidCallTrump) && state.bidPlayerIdx === humanIdx;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  // Legal-play highlight: compute the follow-suit / trump-obligation legal set
  // on the human's play turn only (mirrors internal/domain/Belote.go validatePlay).
  // Ring-only, additive — illegal cards stay clickable and the backend still validates.
  const legalPlayIndices =
    isHumanTurn && humanPlayer
      ? beloteLegalPlayIndices(humanPlayer.cards, state.currentTrick, state.trumpSuit, humanIdx)
      : undefined;
  const faceUpSuit = state.faceUpCard?.design;
  const allSuits = [1, 2, 3, 4];
  const faceUpSuitNum =
    faceUpSuit === 'SPADE'
      ? 1
      : faceUpSuit === 'CLOVER'
        ? 2
        : faceUpSuit === 'HEART'
          ? 3
          : faceUpSuit === 'DIAMOND'
            ? 4
            : 0;
  const callableSuits = allSuits.filter((s) => s !== faceUpSuitNum);

  return (
    <GamePageShell
      title={tc('nav.belote')}
      gameThemeBg={gameTheme.belote.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isHumanBidTurn}
      gamePath="/belote"
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
                value: beloteConfig.cpuDifficulty,
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
                value: beloteConfig.targetScore,
                options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('targetScore', v),
              },
              {
                type: 'checkbox',
                id: 'enableBeloteRebelote',
                label: t('settings.enableBeloteRebelote'),
                checked: beloteConfig.enableBeloteRebelote,
                onToggle: (v) => handleToggle('enableBeloteRebelote', v),
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
        {/* Round/Trick/Trump info */}
        <div className="text-ds-text-primary text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span>
            {state.trumpSuit > 0 ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) }) : t('noTrump')}
          </span>
        </div>
        {/* Bonus trackers — Dix de Der lights up on trick 8; Belote/Rebelote
            lights up as soon as ANY team finishes playing K + Q of trump
            (the bonus is awarded to whichever team holds those cards, not
            specifically the maker — see internal/domain/Belote.go #864). */}
        {state.trumpSuit > 0 &&
          (() => {
            const isLastTrick = state.trickNumber === 8;
            const hasBeloteBonus = state.roundBeloteBonus.some((b) => b > 0);
            return (
              <div
                className="text-center mb-2 flex justify-center gap-2 flex-wrap text-xs"
                data-testid="belote-bonus-trackers"
              >
                {state.config.dixDeDer > 0 && (
                  <span
                    data-testid="dix-de-der-badge"
                    data-active={isLastTrick ? 'true' : undefined}
                    className={
                      isLastTrick
                        ? 'px-2 py-0.5 rounded-full font-medium border bg-ds-accent text-ds-text-on-accent border-ds-accent animate-pulse'
                        : 'px-2 py-0.5 rounded-full font-medium border bg-ds-surface text-ds-text-muted border-ds-border'
                    }
                  >
                    👑 {t('tracker.dixDeDer')}
                  </span>
                )}
                {state.config.enableBeloteRebelote && (
                  <span
                    data-testid="belote-rebelote-badge"
                    data-active={hasBeloteBonus ? 'true' : undefined}
                    className={`${
                      hasBeloteBonus
                        ? 'px-2 py-0.5 rounded-full font-medium border bg-ds-success text-ds-text-on-accent border-ds-success'
                        : 'px-2 py-0.5 rounded-full font-medium border bg-ds-surface text-ds-text-muted border-ds-border'
                    }${beloteJustConfirmed ? ' ring-2 ring-ds-success motion-safe:animate-pulse' : ''}`}
                  >
                    {t('tracker.beloteKing')} · {t('tracker.beloteQueen')}
                    {hasBeloteBonus ? ` ${t('tracker.beloteBonus')}` : ''}
                  </span>
                )}
              </div>
            );
          })()}

        {beloteJustConfirmed && (
          <div
            role="status"
            aria-live="polite"
            data-testid="belote-bonus-confirmed"
            className="text-center mb-2 text-sm font-semibold text-ds-success motion-safe:animate-pulse"
          >
            {t('tracker.beloteConfirmed')}
          </div>
        )}

        {/* Face-up card (during bidding) */}
        {state.faceUpCard && (isBidPickUp || isBidCallTrump) && (
          <div className="mb-3 flex flex-col items-center gap-1 text-ds-text-muted">
            <span className="text-xs">{t('faceUpCard')}</span>
            <AnimatedCard card={state.faceUpCard} width={cardWidth * 0.8} />
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
          dataTutorial="be-trick-display"
        />

        {/* Team scores */}
        <div className="my-3 p-2 rounded bg-black/30" data-tutorial="be-score-table">
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
            </tbody>
          </table>
          {(state.roundBeloteBonus[0] > 0 || state.roundBeloteBonus[1] > 0) && (
            <div className="text-xs text-ds-warning mt-1">{t('beloteRebelote')}</div>
          )}
        </div>

        <RoundScoreAnnouncement
          active={isRoundEnd || isGameEnd}
          entries={[
            {
              name: t('team', { n: 0 }),
              roundScore: state.roundPoints[0] + state.roundBeloteBonus[0],
              cumulativeScore: state.teamScores[0],
            },
            {
              name: t('team', { n: 1 }),
              roundScore: state.roundPoints[1] + state.roundBeloteBonus[1],
              cumulativeScore: state.teamScores[1],
            },
          ]}
        />

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {frontendHintEnabled && frontendHint && (
          <HintTooltip reason={t(frontendHint.reason)} confidence={frontendHint.confidence} />
        )}

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.belote.footer} px-4 py-2.5`}>
        {humanPlayer && (
          <PlayerHandSection
            humanPlayer={humanPlayer}
            selectedCardIndices={selectedCardIndices}
            toggleCard={toggleCard}
            cardWidth={cardWidth}
            isMobile={isMobile}
            dataTutorialPrefix="be"
            legalIndices={legalPlayIndices}
          />
        )}

        <ErrorAlert message={error ?? hintError} onRetry={retry} />

        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {/* hint.reason is a raw backend identifier; translate via hintReason.*,
                falling back to a generic label. The hint shape depends on the phase:
                orderUp (take/pass) and suit (call trump) during bidding, cardIndex in play. */}
            {(() => {
              const reason = t(`hintReason.${hint.reason}`, { defaultValue: t('hintReason.fallback') });
              if (hint.orderUp !== undefined) {
                return `${hint.orderUp ? t('hintOrderUpTake') : t('hintOrderUpPass')} (${reason})`;
              }
              if (hint.suit !== undefined) {
                return `${t('hintCallSuit', { suit: t(SUIT_LABEL_KEYS[hint.suit]) })} (${reason})`;
              }
              return `${t('hintPlay')}: [${hint.cardIndex ?? '-'}] (${reason})`;
            })()}
          </div>
        )}

        <div className="flex gap-2 items-center flex-wrap" data-tutorial="be-play-button">
          {isHumanBidTurn && isBidPickUp && (
            <span data-tutorial="be-bid-controls" className="flex gap-2">
              <button type="button" className={btnPrimary} onClick={handleOrderUp} disabled={loading}>
                {t('orderUpButton')}
              </button>
              <button type="button" className={btnSuccess} onClick={handlePass} disabled={loading}>
                {t('passButton')}
              </button>
            </span>
          )}

          {isHumanBidTurn && isBidCallTrump && (
            <span data-tutorial="be-bid-controls" className="flex gap-2 flex-wrap">
              {callableSuits.map((s) => (
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
              <button type="button" className={btnSuccess} onClick={handlePass} disabled={loading}>
                {t('passButton')}
              </button>
            </span>
          )}

          {(isHumanTurn || isHumanBidTurn) && (
            <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
              {tc('button.hint')}
            </button>
          )}
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
            dataTutorial="be-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
