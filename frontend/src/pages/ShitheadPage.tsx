import { useCallback } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { TutorialWrapper } from '../components/tutorial/TutorialWrapper';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useShitheadGame } from '../hooks/useShitheadGame';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, CardDesign } from '../types/card';
import type { TutorialStep } from '../types/tutorial';

/** Shithead tutorial step definitions. */
const SHITHEAD_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sh-discard"]',
    messageKey: 'tutorial.discardPile',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Source identifier strings (sync with internal/domain/Shithead.go). */
const SOURCE_HAND = 'hand';
const SOURCE_FACE_UP = 'faceup';
const SOURCE_FACE_DOWN = 'facedown';

/** Renders the Shithead / Karma game page. */
export function ShitheadPage() {
  return (
    <TutorialWrapper gameName="shithead" steps={SHITHEAD_TUTORIAL_STEPS}>
      <ShitheadPageContent />
    </TutorialWrapper>
  );
}

/** Inner content of the Shithead page. */
function ShitheadPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('shithead');
  const { state, loading, error, selectedCardIndices, toggleCard, handlePlay, handlePickup, retry } = useShitheadGame();
  const { isMobile: _isMobile } = useCardDimensions();

  const handleManualReset = useCallback(() => {
    hideActionLog();
    window.location.reload();
  }, [hideActionLog]);

  if (!state) {
    return (
      <div className={`flex-1 flex flex-col min-h-0 ${gameTheme.shithead.bg}`}>
        <div className="flex-1 flex items-center justify-center text-white">
          <p>{tc('common.loading')}</p>
        </div>
      </div>
    );
  }

  const isGameEnd = state.gameEndFlag;
  const humanPlayer = state.players.find((p) => p.isHuman);
  const isHumanTurn = state.players[state.currentTurn]?.isHuman === true && !isGameEnd;
  const phaseName = isGameEnd ? t('phase.gameEnd') : t('phase.play');

  return (
    <GamePageShell
      title={tc('nav.shithead')}
      gameThemeBg={gameTheme.shithead.bg}
      phaseName={phaseName}
      isHumanTurn={isHumanTurn}
      gamePath="/shithead"
      gameEndFlag={isGameEnd}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto px-4 py-2 space-y-4">
        {error && <ErrorAlert message={error} onRetry={retry} />}

        {/* CPU player rows */}
        <div className="bg-black/30 text-white p-3 rounded text-sm space-y-1">
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="flex justify-between items-center">
                <span>
                  CPU {p.id}
                  {p.isFinished && ` — ${t('labels.rank')} ${p.rank}`}
                  {p.rank === state.players.length && ` (${t('labels.shithead')})`}
                </span>
                <span>
                  {t('labels.hand')}:{p.handCount} / {t('labels.faceUp')}:{p.faceUpCards.length} /{' '}
                  {t('labels.faceDown')}:{p.faceDownCount}
                </span>
              </div>
            ))}
        </div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        {/* Discard pile and stock */}
        <div data-tutorial="sh-discard" className="bg-black/30 text-white p-3 rounded">
          <div className="flex flex-wrap items-center gap-3 text-sm">
            <span>
              {t('labels.discardPile')}: {state.discardPile.length === 0 ? '—' : describeTopCard(state.discardPile)}
            </span>
            <span>
              {t('labels.stock')}: {state.stockSize}
            </span>
            {state.sevenActive && <span className="text-ds-warning">{t('rules.magicSeven')}</span>}
            {state.skipNext && <span className="text-ds-warning">{t('rules.magicEight')}</span>}
          </div>
        </div>

        {/* Human player area */}
        {humanPlayer && (
          <div data-tutorial="sh-player-hand" className="bg-black/30 text-white p-3 rounded space-y-2">
            <div className="text-sm">
              {t('labels.you')} —{' '}
              {humanPlayer.isFinished
                ? `${t('labels.rank')} ${humanPlayer.rank}`
                : `${t('labels.currentSource')}: ${state.currentSource}`}
            </div>
            {humanPlayer.handCards.length > 0 && (
              <CardRow
                label={t('labels.hand')}
                cards={humanPlayer.handCards}
                selectable={isHumanTurn && state.currentSource === SOURCE_HAND}
                selected={selectedCardIndices}
                onToggle={toggleCard}
              />
            )}
            {humanPlayer.faceUpCards.length > 0 && (
              <CardRow
                label={t('labels.faceUp')}
                cards={humanPlayer.faceUpCards}
                selectable={isHumanTurn && state.currentSource === SOURCE_FACE_UP}
                selected={selectedCardIndices}
                onToggle={toggleCard}
              />
            )}
            {humanPlayer.faceDownCount > 0 && (
              <FaceDownRow
                label={t('labels.faceDown')}
                count={humanPlayer.faceDownCount}
                selectable={isHumanTurn && state.currentSource === SOURCE_FACE_DOWN}
                selected={selectedCardIndices}
                onToggle={toggleCard}
              />
            )}
          </div>
        )}

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.shithead.footer} px-4 py-2.5`}>
        <div className="flex flex-wrap gap-2 items-center">
          <GameResetButton
            isGameEnd={isGameEnd}
            onReset={handleManualReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="sh-reset-button"
          />

          {isHumanTurn && (
            <>
              <button
                type="button"
                className={btnPrimary}
                onClick={handlePlay}
                disabled={loading || selectedCardIndices.length === 0}
              >
                {state.currentSource === SOURCE_FACE_DOWN ? t('actions.blindPlay') : t('actions.play')}
              </button>
              <button type="button" className={btnSecondary} onClick={handlePickup} disabled={loading}>
                {t('actions.pickup')}
              </button>
            </>
          )}
        </div>
      </GameFooter>
    </GamePageShell>
  );
}

/** Returns a textual description of the top discard card. */
function describeTopCard(pile: Card[]): string {
  const top = pile[pile.length - 1];
  if (!top) return '—';
  return `${suitSymbol(top.design)}${top.value}`;
}

/** Returns a Unicode suit symbol for the given suit identifier. */
function suitSymbol(design: CardDesign): string {
  switch (design) {
    case 'SPADE':
      return '♠';
    case 'CLOVER':
      return '♣';
    case 'HEART':
      return '♥';
    case 'DIAMOND':
      return '♦';
    default:
      return '?';
  }
}

interface CardRowProps {
  label: string;
  cards: Card[];
  selectable: boolean;
  selected: number[];
  onToggle: (index: number) => void;
}

/** Row of selectable cards with index labels. */
function CardRow({ label, cards, selectable, selected, onToggle }: CardRowProps) {
  return (
    <div className="space-y-1">
      <div className="text-xs uppercase tracking-wide text-ds-text-muted">{label}</div>
      <div className="flex flex-wrap gap-2">
        {cards.map((c, i) => {
          const isSelected = selected.includes(i);
          const cls = isSelected
            ? 'bg-ds-warning text-ds-text-on-accent border-ds-warning'
            : selectable
              ? 'bg-ds-surface-elevated text-ds-text-primary border-ds-border-subtle hover:bg-ds-surface-elevated-hover'
              : 'bg-ds-surface text-ds-text-muted border-ds-border-subtle';
          return (
            <button
              key={`hand-${i}`}
              type="button"
              disabled={!selectable}
              onClick={() => onToggle(i)}
              className={`min-w-[3rem] px-2 py-2 rounded border text-sm ${cls}`}
            >
              <span className="block leading-none">{suitSymbol(c.design)}</span>
              <span className="block text-base font-bold">{c.value}</span>
            </button>
          );
        })}
      </div>
    </div>
  );
}

interface FaceDownRowProps {
  label: string;
  count: number;
  selectable: boolean;
  selected: number[];
  onToggle: (index: number) => void;
}

/** Row of face-down (hidden) cards selectable for blind play. */
function FaceDownRow({ label, count, selectable, selected, onToggle }: FaceDownRowProps) {
  return (
    <div className="space-y-1">
      <div className="text-xs uppercase tracking-wide text-ds-text-muted">{label}</div>
      <div className="flex flex-wrap gap-2">
        {Array.from({ length: count }).map((_, i) => {
          const isSelected = selected.includes(i);
          const cls = isSelected
            ? 'bg-ds-warning text-ds-text-on-accent border-ds-warning'
            : selectable
              ? 'bg-ds-accent text-ds-text-on-accent border-ds-border-subtle hover:bg-ds-accent-hover'
              : 'bg-ds-surface text-ds-text-muted border-ds-border-subtle';
          return (
            <button
              key={`fd-${i}`}
              type="button"
              disabled={!selectable}
              onClick={() => onToggle(i)}
              className={`min-w-[3rem] px-2 py-2 rounded border text-sm ${cls}`}
            >
              ?
            </button>
          );
        })}
      </div>
    </div>
  );
}
