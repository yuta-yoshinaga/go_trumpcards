import { useCallback } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GameResetDialog } from '../components/GameResetDialog';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { CrazyEightsSkeleton } from '../components/skeleton/CrazyEightsSkeleton';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCardKeyboardNav } from '../hooks/useCardKeyboardNav';
import { CPU_DIFFICULTY_OPTIONS, POINT_LIMIT_OPTIONS, useCrazyEightsGame } from '../hooks/useCrazyEightsGame';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { TutorialProvider, useTutorialContext } from '../providers/TutorialProvider';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { selectedCardStyle } from '../styles/cardStyles';
import { CrazyEightsPhase, CrazyEightsSuit } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { playerName } from '../utils/playerUtils';

const CRAZYEIGHTS_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CrazyEightsPhase.PLAY]: 'play',
  [CrazyEightsPhase.CHOOSE_SUIT]: 'chooseSuit',
  [CrazyEightsPhase.ROUND_END]: 'roundEnd',
  [CrazyEightsPhase.GAME_END]: 'gameEnd',
};

const SUIT_BUTTONS = [
  { suit: CrazyEightsSuit.SPADE, key: 'suitSpade' },
  { suit: CrazyEightsSuit.CLOVER, key: 'suitClover' },
  { suit: CrazyEightsSuit.HEART, key: 'suitHeart' },
  { suit: CrazyEightsSuit.DIAMOND, key: 'suitDiamond' },
] as const;

const SUIT_SYMBOLS: Record<number, string> = {
  [CrazyEightsSuit.SPADE]: '♠',
  [CrazyEightsSuit.CLOVER]: '♣',
  [CrazyEightsSuit.HEART]: '♥',
  [CrazyEightsSuit.DIAMOND]: '♦',
};

/** Crazy Eights tutorial step definitions. */
const CE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="ce-discard-pile"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ce-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="ce-play-draw"]', messageKey: 'tutorial.playDraw', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="ce-suit-choice"]',
    messageKey: 'tutorial.suitChoice',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="ce-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Crazy Eights tutorial configuration. */
const CE_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'crazyeights',
  steps: CE_TUTORIAL_STEPS,
};

/** Tutorial button that starts the Crazy Eights tutorial. */
function TutorialButton() {
  const { t } = useTranslation('tutorial');
  const { start } = useTutorialContext();
  return (
    <button
      type="button"
      className={`${btnSecondary} text-xs`}
      onClick={start}
      aria-label={t('tutorialButton')}
      title={t('tutorialButton')}
    >
      ?
    </button>
  );
}

/** Renders the Crazy Eights game page with card play and suit selection. */
export function CrazyEightsPage() {
  const { t: tCe } = useTranslation('crazyeights');
  return (
    <TutorialProvider config={CE_TUTORIAL_CONFIG} translateMessage={tCe}>
      <CrazyEightsPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the Crazy Eights page, wrapped by TutorialProvider. */
function CrazyEightsPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('crazyeights');
  const {
    state,
    loading,
    error,
    exec: gameExec,
    crazyEightsConfig,
    selectedCardIndices,
    toggleCard,
    clearSelection,
    handleConfigChange,
    handlePlay,
    handleDraw,
    handleChooseSuit,
    handleNextRound,
  } = useCrazyEightsGame();
  const { cardWidth } = useCardDimensions();

  const isPlayPhaseForKbd = state?.phase === CrazyEightsPhase.PLAY;
  const isHumanTurnForKbd = isPlayPhaseForKbd && state?.players[state.currentPlayerIdx]?.isHuman === true;
  const humanCardCountForKbd = state?.players.find((p) => p.isHuman)?.cards?.length ?? 0;

  const confirmAction = useCallback(() => {
    handlePlay();
  }, [handlePlay]);

  useCardKeyboardNav({
    cardCount: humanCardCountForKbd,
    onToggle: toggleCard,
    onConfirm: confirmAction,
    onClear: clearSelection,
    enabled: !!isHumanTurnForKbd && !loading,
  });

  const phaseNames = usePhaseNames('crazyeights', CRAZYEIGHTS_PHASE_KEYS);

  if (!state) return <CrazyEightsSkeleton />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isPlayPhase = state.phase === CrazyEightsPhase.PLAY;
  const isChooseSuit = state.phase === CrazyEightsPhase.CHOOSE_SUIT;
  const isRoundEnd = state.phase === CrazyEightsPhase.ROUND_END;
  const isGameEnd = state.phase === CrazyEightsPhase.GAME_END || state.gameEndFlag;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;

  return (
    <div className="flex-1 flex flex-col min-h-0 bg-game-bg-blue" aria-busy={loading}>
      <PhaseIndicator phaseName={phaseNames[state.phase]} isHumanTurn={isHumanTurn || isChooseSuit}>
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
                value: crazyEightsConfig.cpuDifficulty,
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
                value: crazyEightsConfig.pointLimit,
                options: POINT_LIMIT_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('pointLimit', v),
              },
            ],
          },
        ]}
      />

      <div className="flex-1 overflow-y-auto pt-3 px-4">
        <div className="text-white text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span>{t('drawPile', { count: state.drawPileCount })}</span>
        </div>

        {/* Discard pile top */}
        {state.discardTop && (
          <div className="my-3 p-3 rounded bg-black/40 flex items-center gap-3" data-tutorial="ce-discard-pile">
            <AnimatedCard card={state.discardTop} width={cardWidth} />
            <div className="text-white/70 text-sm">
              <div>{t('discardTop')}</div>
              {state.chosenSuit > 0 && (
                <div className="text-yellow-300">
                  {t('chosenSuit')}: {SUIT_SYMBOLS[state.chosenSuit] ?? '?'}
                </div>
              )}
            </div>
          </div>
        )}

        {/* CPU players */}
        {state.players
          .filter((p) => !p.isHuman)
          .map((p) => (
            <div key={p.id} className="mb-2 p-2 rounded bg-black/30">
              <div className="text-white/70 text-sm">
                {playerName(p.id, p.isHuman)}: {t('cards', { count: p.cardCount })} |{' '}
                {t('cumulativeScore', { score: p.cumulativeScore })} | {t('roundScore', { score: p.roundScore })}
              </div>
            </div>
          ))}

        {/* Score table */}
        <div className="my-3 p-2 rounded bg-black/30">
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

      <GameFooter className="bg-game-bg-blue-dark border-white/20 px-4 py-2.5">
        {humanPlayer && (
          <div className="flex flex-wrap gap-1 mb-2" data-tutorial="ce-player-hand">
            {humanPlayer.cards.map((card, idx) => (
              <button
                type="button"
                key={`${card.design}-${card.value}-${idx}`}
                onClick={() => toggleCard(idx)}
                aria-label={cardAlt(card)}
                aria-pressed={selectedCardIndices.includes(idx)}
                className="transition-transform"
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
          {isHumanTurn && (
            <div className="flex gap-2" data-tutorial="ce-play-draw">
              <button
                type="button"
                className={btnPrimary}
                onClick={handlePlay}
                disabled={loading || selectedCardIndices.length !== 1}
              >
                {t('playButton')}
              </button>
              <button type="button" className={btnPrimary} onClick={handleDraw} disabled={loading}>
                {t('drawButton')}
              </button>
            </div>
          )}
          {isChooseSuit && (
            <div className="flex gap-1" data-tutorial="ce-suit-choice">
              {SUIT_BUTTONS.map(({ suit, key }) => (
                <button
                  key={suit}
                  type="button"
                  className={btnPrimary}
                  onClick={() => handleChooseSuit(suit)}
                  disabled={loading}
                >
                  {t(key)}
                </button>
              ))}
            </div>
          )}
          {isRoundEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
              {t('nextRound')}
            </button>
          )}
          <button
            type="button"
            className={btnWarning}
            data-tutorial="ce-reset-button"
            onClick={() =>
              requestConfirm(() => {
                hideActionLog();
                return gameExec('reset', undefined, undefined, {
                  cpuDifficulty: crazyEightsConfig.cpuDifficulty,
                  pointLimit: crazyEightsConfig.pointLimit,
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
