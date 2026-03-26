import { useCallback } from 'react';
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
import { GinRummySkeleton } from '../components/skeleton/GinRummySkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useGinRummyGame } from '../hooks/useGinRummyGame';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { focusRingCard, selectedCardStyle } from '../styles/cardStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { GinRummyPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

const GINRUMMY_PHASE_KEYS: Readonly<Record<number, string>> = {
  [GinRummyPhase.DRAW]: 'draw',
  [GinRummyPhase.DISCARD]: 'discard',
  [GinRummyPhase.LAYOFF]: 'layoff',
  [GinRummyPhase.ROUND_END]: 'roundEnd',
  [GinRummyPhase.GAME_END]: 'gameEnd',
};

/** Gin Rummy tutorial step definitions. */
const GR_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="gr-draw-area"]', messageKey: 'tutorial.drawArea', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="gr-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-discard-button"]',
    messageKey: 'tutorial.discardButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-knock-button"]',
    messageKey: 'tutorial.knockButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="gr-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Gin Rummy tutorial configuration. */
const GR_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'ginrummy',
  steps: GR_TUTORIAL_STEPS,
};

/** Renders the Gin Rummy game page with draw, discard, knock, and layoff phases. */
export function GinRummyPage() {
  const { t: tGr } = useTranslation('ginrummy');
  return (
    <TutorialProvider config={GR_TUTORIAL_CONFIG} translateMessage={tGr}>
      <GinRummyPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Gin Rummy page, wrapped by TutorialProvider. */
function GinRummyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('ginrummy');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    ginRummyConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handleDrawStock,
    handleDrawDiscard,
    handleDiscard,
    handleKnock,
    handleLayoff,
    handleSkipLayoff,
    handleNextRound,
  } = useGinRummyGame();
  const { cardWidth } = useCardDimensions();

  const isDiscardPhaseForKbd = state?.phase === GinRummyPhase.DISCARD;
  const isLayoffPhaseForKbd = state?.phase === GinRummyPhase.LAYOFF;
  const isHumanTurnForKbd =
    (isDiscardPhaseForKbd || isLayoffPhaseForKbd) && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    if (isDiscardPhaseForKbd) {
      handleDiscard();
    } else if (isLayoffPhaseForKbd) {
      handleLayoff();
    }
  }, [isDiscardPhaseForKbd, isLayoffPhaseForKbd, handleDiscard, handleLayoff]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('ginrummy', GINRUMMY_PHASE_KEYS);

  if (!state) return <GinRummySkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isDrawPhase = state.phase === GinRummyPhase.DRAW;
  const isDiscardPhase = state.phase === GinRummyPhase.DISCARD;
  const isLayoffPhase = state.phase === GinRummyPhase.LAYOFF;
  const isRoundEnd = state.phase === GinRummyPhase.ROUND_END;
  const isGameEnd = state.phase === GinRummyPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn =
    (isDrawPhase || isDiscardPhase || isLayoffPhase) && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.ginrummy.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.ginrummy')} />
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanTurn}>
        <TutorialButton />
      </PhaseIndicator>

      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: ginRummyConfig.cpuDifficulty,
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
                value: ginRummyConfig.pointLimit,
                options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('pointLimit', v),
              },
            ],
          },
        ]}
      />

      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        <div className="text-white text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span>{t('drawPile', { count: state.drawPileCount })}</span>
        </div>

        {/* Discard pile top */}
        {state.discardTop && (
          <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3">
            <AnimatedCard card={state.discardTop} width={cardWidth} />
            <div className="text-white/70 text-sm">
              <div>{t('discardTop')}</div>
            </div>
          </div>
        )}

        {/* CPU player */}
        {state.players
          .filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-white/70 text-sm">
                {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}
              </div>
              {/* Show CPU cards during layoff/round end/game end */}
              {(isLayoffPhase || isRoundEnd || isGameEnd) && p.cards.length > 0 && (
                <div className="flex flex-wrap gap-1 mt-1">
                  {p.cards.map((card, idx) => (
                    <AnimatedCard key={`cpu-${card.design}-${card.value}-${idx}`} card={card} width={cardWidth * 0.8} />
                  ))}
                </div>
              )}
            </div>
          ))}

        {/* Knocker melds */}
        {state.knockerMelds.length > 0 && (
          <div className="my-3 p-2 rounded bg-black/30">
            <div className="text-white/70 text-sm mb-1">{t('knockerMelds')}</div>
            {state.knockerMelds.map((meld, meldIdx) => (
              <div key={`meld-${meldIdx}`} className="flex flex-wrap gap-1 mb-1">
                {meld.cards.map((card, cardIdx) => (
                  <AnimatedCard
                    key={`meld-${meldIdx}-${card.design}-${card.value}-${cardIdx}`}
                    card={card}
                    width={cardWidth * 0.7}
                  />
                ))}
              </div>
            ))}
          </div>
        )}

        {/* Score table */}
        <div className="my-3 p-2 rounded bg-black/30" data-tutorial="gr-score-table">
          <div className="text-white/70 text-sm mb-1">{t('scores')}</div>
          <table className="w-full text-sm text-white/70">
            <thead>
              <tr>
                <th scope="col" className="text-left">
                  {t('scoresPlayer')}
                </th>
                <th scope="col">{t('scoresRound')}</th>
                <th scope="col">{t('scoresTotal')}</th>
              </tr>
            </thead>
            <tbody>
              {state.players.map((p) => (
                <tr key={p.id} className={p.isHuman ? 'text-yellow-300' : ''}>
                  <td>{playerName(p.id, p.isHuman)}</td>
                  <td className="text-center">{p.roundScore}</td>
                  <td className="text-center">{p.cumulativeScore}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.ginrummy.footer} px-4 py-2.5`}>
        {humanPlayer && (
          <div className="flex flex-wrap gap-1 mb-2" data-tutorial="gr-player-hand">
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
                }}
              >
                <AnimatedCard card={card} width={cardWidth} />
              </button>
            ))}
          </div>
        )}

        <ErrorAlert message={error} />

        <div className="flex gap-2 items-center flex-wrap">
          {isDrawPhase && isHumanTurn && (
            <div className="flex gap-2" data-tutorial="gr-draw-area">
              <button type="button" className={btnPrimary} onClick={handleDrawStock} disabled={loading}>
                {t('drawStockButton')}
              </button>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleDrawDiscard}
                disabled={loading || !state.discardTop}
              >
                {t('drawDiscardButton')}
              </button>
            </div>
          )}
          {isDiscardPhase && isHumanTurn && (
            <>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleDiscard}
                disabled={loading || selectedCardIndices.length !== 1}
                data-tutorial="gr-discard-button"
              >
                {t('discardButton')}
              </button>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleKnock}
                disabled={loading || selectedCardIndices.length !== 1}
                data-tutorial="gr-knock-button"
              >
                {t('knockButton')}
              </button>
            </>
          )}
          {isLayoffPhase && isHumanTurn && (
            <>
              <button
                type="button"
                className={btnPrimary}
                onClick={handleLayoff}
                disabled={loading || selectedCardIndices.length === 0}
              >
                {t('layoffButton')}
              </button>
              <button type="button" className={btnPrimary} onClick={handleSkipLayoff} disabled={loading}>
                {t('skipLayoffButton')}
              </button>
            </>
          )}
          {isRoundEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
              {t('nextRound')}
            </button>
          )}
          <button
            type="button"
            className={btnWarning}
            data-tutorial="gr-reset-button"
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                return gameExec('reset', undefined, {
                  cpuDifficulty: ginRummyConfig.cpuDifficulty,
                  pointLimit: ginRummyConfig.pointLimit,
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
