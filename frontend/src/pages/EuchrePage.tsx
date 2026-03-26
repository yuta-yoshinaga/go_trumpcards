import { useCallback, useState } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { EuchreSkeleton } from '../components/skeleton/EuchreSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useEuchreGame } from '../hooks/useEuchreGame';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { gameTheme } from '../styles/gameTheme';
import { EuchrePhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

const SUIT_NAMES: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

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

/** Euchre tutorial configuration. */
const EU_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'euchre',
  steps: EU_TUTORIAL_STEPS,
};

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
export function EuchrePage() {
  const { t: tEuchre } = useTranslation('euchre');
  return (
    <TutorialProvider config={EU_TUTORIAL_CONFIG} translateMessage={tEuchre}>
      <EuchrePageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Euchre page, wrapped by TutorialProvider. */
function EuchrePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('euchre');
  const {
    state,
    loading,
    error,
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
  const { cardWidth, isMobile, solitaireMinColWidth } = useCardDimensions();
  const [goAlone, setGoAlone] = useState(false);

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

  if (!state) return <EuchreSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanTeam = humanPlayer?.team ?? 0;
  const isPickUpPhase = state.phase === EuchrePhase.PICK_UP;
  const isCallTrumpPhase = state.phase === EuchrePhase.CALL_TRUMP;
  const isDiscardPhase = state.phase === EuchrePhase.DISCARD;
  const isPlayPhase = state.phase === EuchrePhase.PLAY;
  const isTrickEnd = state.phase === EuchrePhase.TRICK_END;
  const isRoundEnd = state.phase === EuchrePhase.ROUND_END;
  const isGameEnd = state.phase === EuchrePhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  const isHumanBidTurn = (isPickUpPhase || isCallTrumpPhase) && state.players[state.bidPlayerIdx]?.isHuman === true;
  const isHumanDiscard = isDiscardPhase && state.players[state.dealerIdx]?.isHuman === true;

  const suitName = (suit: number) => (SUIT_NAMES[suit] ? t(SUIT_NAMES[suit]) : '');

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.euchre.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.euchre')} />
      {/* Phase indicator */}
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanBidTurn || isHumanTurn || isHumanDiscard}>
        <TutorialButton />
      </PhaseIndicator>

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
            ],
          },
        ]}
      />

      {/* Scrollable area */}
      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8 lg:max-w-5xl lg:mx-auto lg:w-full">
        {/* Round/Trick info */}
        <div className="text-white text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          {state.trumpSuit > 0 && <span>{t('trumpSuit', { suit: suitName(state.trumpSuit) })}</span>}
          {state.trumpSuit === 0 && <span>{t('noTrump')}</span>}
        </div>

        {/* Maker / Going alone info */}
        {state.makerTeam >= 0 && (
          <div className="text-yellow-300 text-center mb-2">
            <span className="mr-4">{t('maker', { team: state.makerTeam })}</span>
            {state.goingAlone && <span>{t('goingAlone')}</span>}
          </div>
        )}

        {/* Pick-up phase instruction */}
        {isHumanBidTurn && isPickUpPhase && (
          <div className="text-yellow-300 text-center mb-2" data-tutorial="eu-pickup-controls">
            {t('pickUpPhase')}
          </div>
        )}

        {/* Call trump phase instruction */}
        {isHumanBidTurn && isCallTrumpPhase && (
          <div className="text-yellow-300 text-center mb-2">{t('callTrumpPhase')}</div>
        )}

        {/* Discard phase instruction */}
        {isHumanDiscard && <div className="text-yellow-300 text-center mb-2">{t('discardPhase')}</div>}

        {/* Face-up card */}
        {state.faceUpCard && (isPickUpPhase || isCallTrumpPhase) && (
          <div className="my-2 text-center">
            <div className="text-white/70 text-sm mb-1">{t('faceUpCard')}</div>
            <div className="inline-block">
              <AnimatedCard card={state.faceUpCard} width={cardWidth} />
            </div>
          </div>
        )}

        {/* CPU players */}
        {state.players
          .filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-white/70 text-sm">
                {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} | {t('team', { n: p.team })} |{' '}
                {t('trickCount', { count: p.trickCount })}
                {state.dealerIdx === p.id ? ` | ${t('dealer')}` : ''}
              </div>
            </div>
          ))}

        {/* Current trick */}
        {state.currentTrick.length > 0 && (
          <div className="my-3 p-3 rounded bg-black/40" data-tutorial="eu-trick-display">
            <div className="text-white/70 text-sm mb-1">{t('currentTrick')}</div>
            <div className="flex gap-2">
              {state.currentTrick.map((trickCard) => (
                <div key={`trick-${trickCard.playerIdx}`} className="text-center">
                  <AnimatedCard card={trickCard.card} width={cardWidth} />
                  <div className="text-game-text-muted text-xs mt-1">
                    {playerName(
                      state.players[trickCard.playerIdx]?.id ?? trickCard.playerIdx,
                      state.players[trickCard.playerIdx]?.isHuman ?? false,
                    )}
                  </div>
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Team scores */}
        <div className="my-3 p-2 rounded bg-black/30" data-tutorial="eu-score-table">
          <div className="text-white/70 text-sm mb-1">{t('teamScores')}</div>
          <table className="w-full text-sm text-white/70">
            <thead>
              <tr>
                <th scope="col" className="text-left">
                  {t('team', { n: '' })}
                </th>
                <th scope="col">{tc('button.score', { defaultValue: 'Score' })}</th>
              </tr>
            </thead>
            <tbody>
              {state.teamScores.map((score, idx) => (
                <tr key={idx} className={idx === humanTeam ? 'text-yellow-300' : ''}>
                  <td>{idx === humanTeam ? t('teamYou', { n: idx }) : t('team', { n: idx })}</td>
                  <td className="text-center">{score}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        {/* Partnership info */}
        {humanPlayer && (
          <div className="text-white/70 text-sm text-center mb-2" data-tutorial="eu-team-info">
            {t('partnership', {
              partner: playerName(state.players.find((p) => !p.isHuman && p.team === humanTeam)?.id ?? -1, false),
            })}
            {state.dealerIdx === humanPlayer.id ? ` | ${t('dealer')}` : ''}
          </div>
        )}

        {/* Message */}
        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

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
            {humanPlayer.cards.map((card, idx) => (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => toggleCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={selectedCardIndices.includes(idx)}
                className={`transition-transform ${focusRingCard}`}
                style={{
                  background: 'none',
                  padding: 0,
                  borderRadius: 8,
                  ...selectedCardStyle(selectedCardIndices.includes(idx)),
                  boxSizing: 'border-box',
                  ...(isMobile ? { minWidth: solitaireMinColWidth, flexShrink: 0 } : {}),
                }}
              >
                <AnimatedCard card={card} width={cardWidth} />
              </button>
            ))}
          </div>
        )}

        <ErrorAlert message={error ?? hintError} />

        {hint && (
          <div className="text-yellow-300 text-sm mb-2">
            {hint.cardIndex != null
              ? `${t('hintPlay')}: [${hint.cardIndex}] (${t(`hintReason.${hint.reason}`)})`
              : `(${t(`hintReason.${hint.reason}`)})`}
          </div>
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
              {[1, 2, 3, 4]
                .filter(
                  (s) =>
                    state.faceUpCard == null ||
                    s !==
                      ({ SPADE: 1, CLOVER: 2, HEART: 3, DIAMOND: 4, JOKER: 0 } as Record<string, number>)[
                        state.faceUpCard.design
                      ],
                )
                .map((s) => (
                  <button
                    key={s}
                    type="button"
                    className={btnPrimary}
                    onClick={() => handleCallTrump(s, goAlone)}
                    disabled={loading}
                  >
                    {t(SUIT_NAMES[s])}
                  </button>
                ))}
              <label className="text-white text-sm flex items-center gap-1">
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
          <button
            type="button"
            className={btnWarning}
            data-tutorial="eu-reset-button"
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                return apiExec('reset', undefined, undefined, undefined, {
                  cpuDifficulty: euchreConfig.cpuDifficulty,
                  pointLimit: euchreConfig.pointLimit,
                });
              })
            }
            disabled={loading}
          >
            {tc('button.reset')}
          </button>
        </div>
      </GameFooter>
      <WinCelebration show={!!state?.gameEndFlag} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
