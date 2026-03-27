import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import { ActionLogSection } from '../components/ActionLogSection';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageHeading } from '../components/GamePageHeading';
import { GameResetDialog } from '../components/GameResetDialog';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { WinCelebration } from '../components/motion/WinCelebration';
import { PhaseIndicator } from '../components/PhaseIndicator';
import { ScrollFadeHint } from '../components/ScrollFadeHint';
import { FreeCellSkeleton } from '../components/skeleton/FreeCellSkeleton';
import { TutorialButton } from '../components/tutorial/TutorialButton';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useFreeCellGame } from '../hooks/useFreeCellGame';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { TutorialProvider } from '../providers/TutorialProvider';
import { btnDanger, btnPrimary, btnSuccess, btnWarning, focusRingWhite } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card } from '../types/card';
import { FreeCellPhase } from '../types/phases';
import type { TutorialConfig, TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

const FOUNDATION_SUITS = ['♠', '♣', '♥', '♦'] as const;

/** FreeCell tutorial step definitions. */
const FC_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="fc-free-cells"]',
    messageKey: 'tutorial.freeCells',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-tableau"]',
    messageKey: 'tutorial.tableau',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-controls"]',
    messageKey: 'tutorial.controls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="fc-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** FreeCell tutorial configuration. */
const FC_TUTORIAL_CONFIG: TutorialConfig = {
  gameName: 'freecell',
  steps: FC_TUTORIAL_STEPS,
};

/** Renders the FreeCell solitaire game page with tableau, free cells, and foundation. */
export function FreeCellPage() {
  const { t: tFc } = useTranslation('freecell');
  return (
    <TutorialProvider config={FC_TUTORIAL_CONFIG} translateMessage={tFc}>
      <FreeCellPageContent />
    </TutorialProvider>
  );
}

/** Inner content of the FreeCell page, wrapped by TutorialProvider. */
function FreeCellPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('freecell');
  const {
    state,
    loading,
    error,
    hintError,
    selectedSource,
    hint,
    handleReset,
    handleGiveUp,
    handleHint,
    handleAutoComplete,
    handleUndo,
    handleSelectSource,
    handleSelectTarget,
  } = useFreeCellGame();
  const { cardHeight, cardOverlap, cardWidth, isMobile, solitaireMinColWidth } = useCardDimensions();

  const isPlayingForKbd = state?.phase === FreeCellPhase.PLAYING;

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: handleHint },
      { key: 'a', action: handleAutoComplete },
      { key: 'g', action: handleGiveUp },
      { key: 'z', action: handleUndo },
    ],
    [handleHint, handleAutoComplete, handleGiveUp, handleUndo],
  );

  useActionKeyboardNav({
    bindings: actionBindings,
    enabled: !!isPlayingForKbd && !loading,
  });

  if (!state) return <FreeCellSkeleton />;

  const isPlaying = state.phase === FreeCellPhase.PLAYING;
  const isGameClear = state.phase === FreeCellPhase.GAME_CLEAR;
  const isGameOver = state.phase === FreeCellPhase.GAME_OVER;
  const isEnded = isGameClear || isGameOver;

  const isSourceSelected = (zone: string, col?: number, cell?: number, cardIndex?: number) =>
    selectedSource !== null &&
    selectedSource.zone === zone &&
    selectedSource.col === col &&
    selectedSource.cell === cell &&
    selectedSource.cardIndex === cardIndex;

  return (
    <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.freecell.bg}`} aria-busy={loading} aria-live="polite">
      <GamePageHeading title={tc('nav.freecell')} />
      <PhaseIndicator
        phaseName={isGameClear ? t('phase.gameClear') : isGameOver ? t('phase.gameOver') : t('phase.playing')}
      >
        <span>
          {t('moveCount')}: {state.moveCount}
        </span>
        <TutorialButton />
      </PhaseIndicator>

      <LandscapeBanner message={t('landscapeBanner')} />

      <div className="flex-1 overflow-y-auto pt-3 px-4 lg:px-8">
        {/* Free cells + Foundation row */}
        <div className="flex gap-2 mb-3 items-start flex-wrap">
          {/* Free cells */}
          <div className="flex gap-2" data-tutorial="fc-free-cells">
            {state.freeCells.map((card: Card | null, idx: number) => (
              <div key={`fc-${idx.toString()}`} className="text-center">
                <div className="text-game-text-muted text-xs mb-1">
                  <span className="hidden sm:inline">
                    {t('freecell')} {idx}
                  </span>
                  <span className="sm:hidden">
                    {t('freecellShort')}
                    {idx}
                  </span>
                </div>
                {card ? (
                  <button
                    type="button"
                    onClick={() => handleSelectSource({ zone: 'freecell', cell: idx })}
                    disabled={!isPlaying || loading}
                    aria-label={cardAlt(card)}
                    aria-pressed={isSourceSelected('freecell', undefined, idx)}
                    className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite} ${isSourceSelected('freecell', undefined, idx) ? 'ring-2 ring-yellow-400' : ''}`}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => handleSelectTarget({ zone: 'freecell', cell: idx })}
                    disabled={!isPlaying || loading || !selectedSource}
                    aria-label={t('emptyFreecellAriaLabel', { idx: String(idx) })}
                    style={{ width: cardWidth, height: cardHeight }}
                    className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    {t('empty')}
                  </button>
                )}
              </div>
            ))}
          </div>

          <div className="w-4" />

          {/* Foundation piles */}
          <div className="flex gap-2" data-tutorial="fc-foundation">
            {state.foundation.map((pile: Card[], idx: number) => (
              <div key={`f-${idx.toString()}`} className="text-center">
                <div className="text-game-text-muted text-xs mb-1">{FOUNDATION_SUITS[idx]}</div>
                {pile.length > 0 ? (
                  <button
                    type="button"
                    onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                    disabled={!isPlaying || loading || !selectedSource}
                    aria-label={t('foundationAriaLabel', {
                      suit: FOUNDATION_SUITS[idx],
                      cardCount: String(pile.length),
                    })}
                    className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}`}
                  >
                    <AnimatedCard card={pile[pile.length - 1]} width={cardWidth} />
                  </button>
                ) : (
                  <button
                    type="button"
                    onClick={() => handleSelectTarget({ zone: 'foundation', col: idx })}
                    disabled={!isPlaying || loading || !selectedSource}
                    aria-label={t('emptyFoundationAriaLabel', { suit: FOUNDATION_SUITS[idx] })}
                    style={{ width: cardWidth, height: cardHeight }}
                    className={`rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                  >
                    A
                  </button>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Tableau */}
        <div className="relative">
          <div className="flex gap-2 mb-3 overflow-x-auto" data-tutorial="fc-tableau">
            {state.tableau.map((col: (Card | null)[], colIdx: number) => (
              <div
                key={`col-${colIdx.toString()}`}
                className={isMobile ? 'flex-shrink-0' : 'flex-1'}
                style={isMobile ? { width: solitaireMinColWidth } : undefined}
              >
                <div className="text-game-text-muted text-xs text-center mb-1">{colIdx}</div>
                <div className="relative" style={{ minHeight: cardHeight }}>
                  {col.length === 0 ? (
                    <button
                      type="button"
                      onClick={() => handleSelectTarget({ zone: 'tableau', col: colIdx })}
                      disabled={!isPlaying || loading || !selectedSource}
                      style={{ height: cardHeight }}
                      className={`w-full rounded border-2 border-dashed border-white/20 text-game-text-muted text-xs flex items-center justify-center ${focusRingWhite}`}
                    >
                      K
                    </button>
                  ) : (
                    col.map((card: Card | null, cardIdx: number) => (
                      <div
                        key={`tc-${colIdx.toString()}-${cardIdx.toString()}`}
                        className="absolute left-0 right-0"
                        style={{ top: cardIdx * cardOverlap }}
                      >
                        {card ? (
                          <button
                            type="button"
                            onClick={() => {
                              if (selectedSource) {
                                handleSelectTarget({ zone: 'tableau', col: colIdx });
                              } else {
                                handleSelectSource({ zone: 'tableau', col: colIdx, cardIndex: cardIdx });
                              }
                            }}
                            disabled={!isPlaying || loading}
                            aria-label={cardAlt(card)}
                            aria-pressed={isSourceSelected('tableau', colIdx, undefined, cardIdx)}
                            className={`p-0 border-0 bg-transparent cursor-pointer w-full rounded ${focusRingWhite} ${isSourceSelected('tableau', colIdx, undefined, cardIdx) ? 'ring-2 ring-yellow-400' : ''}`}
                          >
                            <AnimatedCard card={card} width={cardWidth} style={{ width: '100%' }} />
                          </button>
                        ) : (
                          <div style={{ width: cardWidth, height: cardHeight }} />
                        )}
                      </div>
                    ))
                  )}
                  {col.length > 0 && <div style={{ height: (col.length - 1) * cardOverlap + cardHeight }} />}
                </div>
              </div>
            ))}
          </div>
          {isMobile && <ScrollFadeHint />}
        </div>

        {/* Hint display */}
        {hint && (
          <div className="text-yellow-300 text-sm mb-2">
            {t('hintAvailable')}: {hint.fromZone}
            {hint.fromCol >= 0 ? ` ${hint.fromCol}` : ''} → {hint.toZone}
            {hint.toCol >= 0 ? ` ${hint.toCol}` : ''}
          </div>
        )}

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <ActionLogSection
          isEndPhase={isEnded}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.freecell.footer} px-4 py-2.5`}>
        <ErrorAlert message={error ?? hintError} />
        <div className="flex gap-2 items-center flex-wrap">
          {isPlaying && (
            <div data-tutorial="fc-controls">
              <button type="button" className={btnPrimary} onClick={handleUndo} disabled={loading || !state.canUndo}>
                {t('undo')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading}>
                {t('hint')}
              </button>
              <button type="button" className={btnSuccess} onClick={handleAutoComplete} disabled={loading}>
                {t('autoComplete')}
              </button>
              <button type="button" className={btnDanger} onClick={handleGiveUp} disabled={loading}>
                {t('giveup')}
              </button>
            </div>
          )}
          <div data-tutorial="fc-reset-button">
            <button
              type="button"
              className={btnWarning}
              onClick={() =>
                requestConfirm(() => {
                  hideActionLog();
                  return handleReset();
                })
              }
              disabled={loading}
            >
              {tc('button.reset')}
            </button>
          </div>
        </div>
      </GameFooter>
      <WinCelebration show={state.phase === FreeCellPhase.GAME_CLEAR} />
      <GameResetDialog confirmOpen={confirmOpen} confirmReset={confirmReset} cancelReset={cancelReset} />
    </div>
  );
}
