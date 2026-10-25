import type { TFunction } from 'i18next';
import { useEffect, useRef, useState } from 'react';
import type { sultanApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { StalemateEscapeButton } from '../components/StalemateEscapeButton';
import { useResponsiveTableau } from '../hooks/useResponsiveTableau';
import { useSultanGame } from '../hooks/useSultanGame';
import { btnDanger, btnPrimary, btnSuccess, focusRingWhite } from '../styles/buttonStyles';
import { SultanPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSultanCommand, SULTAN_HELP } from '../utils/cli/commands/sultanCommands';
import { formatSultanState } from '../utils/cli/formatters/sultanFormatter';
import { sultanFoundationInfo } from '../utils/sultanFoundation';

/**
 * Maximum number of waste redeals allowed in Sultan of Turkey.
 * Mirrors `domain.SultanMaxRedeal` (total of 3 passes through the stock).
 */
const SULTAN_MAX_REDEAL = 2;

/** Sultan of Turkey tutorial step definitions. */
const SULTAN_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sultan-stock-waste"]',
    messageKey: 'tutorial.stockWaste',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sultan-foundation"]',
    messageKey: 'tutorial.foundation',
    placement: 'bottom',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sultan-divan"]', messageKey: 'tutorial.divan', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="sultan-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
];

/** Format a frontend hint zone for display. */
function formatHintZone(t: (key: string, opts?: Record<string, unknown>) => string, zone: string, idx: number): string {
  if (zone === 'waste') return t('frontendHint.waste');
  return t('frontendHint.divan', { idx });
}

type SultanGame = ReturnType<typeof useSultanGame>;
type SultanView = { autoCompleteAnnouncement: string; phaseName: string };
type SultanContext = GamePageContext<Awaited<ReturnType<typeof sultanApi.exec>>, SultanGame, SultanView>;

function useSultanView({ game, t }: { game: SultanGame; t: TFunction }): SultanView {
  const [autoCompleteAnnouncement, setAutoCompleteAnnouncement] = useState('');
  const previousAutoCompleting = useRef(game.isAutoCompleting);
  useEffect(() => {
    if (previousAutoCompleting.current === game.isAutoCompleting) return;
    previousAutoCompleting.current = game.isAutoCompleting;
    setAutoCompleteAnnouncement(t(game.isAutoCompleting ? 'autoCompleteStarted' : 'autoCompleteCompleted'));
  }, [game.isAutoCompleting, t]);
  return {
    autoCompleteAnnouncement,
    phaseName:
      game.state?.phase === SultanPhase.GAME_CLEAR
        ? t('phase.gameClear')
        : game.state?.phase === SultanPhase.GAME_OVER
          ? t('phase.gameOver')
          : t('phase.playing'),
  };
}

function Board({ game, state, t, frontendHint, view }: SultanContext) {
  const { loading, hint, handleDraw, handlePlay, isAutoCompleting } = game;
  const frontendHintEnabled = frontendHint.hintEnabled;
  const { autoCompleteAnnouncement } = view;
  const isPlaying = state.phase === SultanPhase.PLAYING;
  // Waste display: show top card only
  const wasteDisplay = state.waste.slice(-1);
  const wasteTargets = state.legalTargets.waste;
  const wastePlayable =
    wasteDisplay[0] !== undefined && wasteTargets !== undefined && wasteTargets.foundation.length > 0;
  const sultan = useResponsiveTableau(8);
  return (
    <>
      {/* Stock/Waste + Foundation row */}
      <div className="flex gap-1 sm:gap-2 mb-3 items-start flex-wrap">
        {/* Stock + Waste */}
        <div className="flex gap-1 sm:gap-2" data-tutorial="sultan-stock-waste">
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">
              {t('stock')} ({state.stockCount})
            </div>
            {state.stockCount > 0 ? (
              <AnimatedCardBack width={sultan.cw} onClick={isPlaying ? handleDraw : undefined} ariaLabel={t('draw')} />
            ) : (
              <div
                style={{ width: sultan.cw, height: sultan.ch }}
                className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex items-center justify-center"
              >
                {t('empty')}
              </div>
            )}
          </div>

          {/* Waste */}
          <div className="text-center">
            <div className="text-game-text-muted text-xs mb-1">{t('waste')}</div>
            {wasteDisplay.length > 0 ? (
              <button
                type="button"
                onClick={() => handlePlay({ zone: 'waste' })}
                disabled={!isPlaying || loading || isAutoCompleting}
                aria-label={t(wastePlayable ? 'wastePlayableCard' : 'wasteUnplayableCard', {
                  label: cardAlt(wasteDisplay[0]),
                })}
                className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}${wastePlayable ? ' ring-2 ring-ds-success' : ''}`}
                data-playable={wastePlayable ? 'true' : undefined}
              >
                <AnimatedCard card={wasteDisplay[0]} width={sultan.cw} draggable={false} />
              </button>
            ) : (
              <div
                style={{ width: sultan.cw, height: sultan.ch }}
                className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
              >
                {t('empty')}
              </div>
            )}
          </div>
        </div>

        <div className="w-2 sm:w-4" />

        {/* Foundation piles (8 King-based piles) */}
        <div className="flex gap-1 sm:gap-2 flex-wrap" data-tutorial="sultan-foundation">
          {state.foundation.map((pile, idx) => {
            const info = sultanFoundationInfo(pile);
            return (
              <div key={`f-${idx.toString()}`} className="text-center">
                <div
                  className="text-game-text-muted text-xs mb-1"
                  data-testid={`sultan-foundation-label-${idx.toString()}`}
                >
                  {info.suit ? `${idx + 1} ${info.suit}` : idx + 1}
                </div>
                {pile.length > 0 ? (
                  <div
                    role="img"
                    aria-label={t(info.complete ? 'foundationSlotComplete' : 'foundationSlot', {
                      idx: idx + 1,
                      top: cardAlt(pile[pile.length - 1]),
                    })}
                    className={`relative inline-block rounded ${info.complete ? 'ring-2 ring-ds-success' : ''}`}
                  >
                    <AnimatedCard
                      card={pile[pile.length - 1]}
                      width={sultan.cw}
                      draggable={false}
                      dealDelay={isAutoCompleting ? idx * 0.15 : 0}
                    />
                    {info.complete && (
                      <span
                        aria-hidden="true"
                        data-testid={`sultan-foundation-complete-${idx.toString()}`}
                        className="absolute -top-1 -right-1 rounded-full bg-ds-success text-ds-text-on-accent text-[10px] leading-none px-1 py-0.5"
                      >
                        ✓
                      </span>
                    )}
                  </div>
                ) : (
                  <div
                    role="img"
                    aria-label={t('emptyFoundationSlot', { idx: idx + 1 })}
                    style={{ width: sultan.cw, height: sultan.ch }}
                    className="rounded border-2 border-dashed border-white/30 text-game-text-muted text-xs flex flex-col items-center justify-center leading-tight"
                  >
                    <span>{idx + 1}</span>
                    <span>K</span>
                  </div>
                )}
              </div>
            );
          })}
        </div>
      </div>

      {/* Divan reserve (8 slots; null = played/empty) */}
      <div className="flex gap-1 sm:gap-2 mb-3 flex-wrap" data-tutorial="sultan-divan">
        {state.divan.map((dcard, idx) => {
          const divanTargets = state.legalTargets[`divan-${idx}`];
          const divanPlayable = dcard !== null && divanTargets !== undefined && divanTargets.foundation.length > 0;
          return (
            <div key={`divan-${idx.toString()}`} className="text-center">
              <div className="text-game-text-muted text-xs mb-1">{idx}</div>
              {dcard ? (
                <button
                  type="button"
                  onClick={() => handlePlay({ zone: 'divan', divanIdx: idx })}
                  disabled={!isPlaying || loading || isAutoCompleting}
                  aria-label={t(divanPlayable ? 'divanPlayableCard' : 'divanUnplayableCard', {
                    label: t('divanCard', { idx, card: cardAlt(dcard) }),
                  })}
                  className={`p-0 border-0 bg-transparent cursor-pointer rounded ${focusRingWhite}${divanPlayable ? ' ring-2 ring-ds-success' : ''}`}
                  data-playable={divanPlayable ? 'true' : undefined}
                >
                  <AnimatedCard card={dcard} width={sultan.cw} draggable={false} />
                </button>
              ) : (
                <div
                  role="img"
                  aria-label={t('emptyDivanSlot', { idx })}
                  style={{ width: sultan.cw, height: sultan.ch }}
                  className="rounded border border-white/20 flex items-center justify-center text-game-text-muted text-xs"
                >
                  {t('empty')}
                </div>
              )}
            </div>
          );
        })}
      </div>

      {/* Hint display */}
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで現れるので変化として扱われず、読み上げられない
        ことがある (#5602)。
      */}
      <div data-tutorial="sultan-hint-display" data-testid="sultan-hint-live" role="status" aria-live="polite">
        {hint && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {formatHintZone(t, hint.fromZone, hint.fromIdx)} →{' '}
            {t('frontendHint.foundation', { idx: hint.toFoundation })}
          </div>
        )}
      </div>
      <div data-testid="sultan-autocomplete-live" className="sr-only" role="status" aria-live="polite">
        {autoCompleteAnnouncement}
      </div>
      <div className="flex justify-center">
        <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHintEnabled} t={t} />
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function HeaderExtra({ state, t }: SultanContext) {
  return (
    <span>
      {t('moveCount')}: {state.moveCount}
    </span>
  );
}

function FooterActions({ game, state, t, busy, confirmGiveUpAction }: SultanContext) {
  const { handleDraw, handleRedeal, handleUndo, handleUndoEscape, handleHint, handleAutoComplete, isAutoCompleting } =
    game;
  const isPlaying = state.phase === SultanPhase.PLAYING;
  // Auto-complete is offered once the stock and waste are empty.
  const autoCompleteReady = state.stockCount === 0 && state.waste.length === 0;
  return (
    <>
      {isPlaying && (
        <div className="flex flex-wrap gap-2" data-tutorial="sultan-controls">
          <button
            type="button"
            className={btnPrimary}
            onClick={handleDraw}
            disabled={busy || isAutoCompleting || state.stockCount === 0}
          >
            {t('draw')}
          </button>
          {state.canRedeal && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handleRedeal}
              disabled={busy || isAutoCompleting}
              data-testid="sultan-redeal-count"
            >
              {t('redealWithCount', {
                count: SULTAN_MAX_REDEAL - state.redealCount,
              })}
            </button>
          )}
          <button
            type="button"
            className={btnPrimary}
            onClick={handleUndo}
            disabled={busy || isAutoCompleting || !state.canUndo}
          >
            {t('undo')}
          </button>
          {state.isStalemate && (
            <StalemateEscapeButton
              undoToEscape={state.undoToEscape ?? 0}
              onEscape={handleUndoEscape}
              disabled={busy || isAutoCompleting}
            />
          )}
          <button type="button" className={btnSuccess} onClick={handleHint} disabled={busy || isAutoCompleting}>
            {t('hint')}
          </button>
          <button
            type="button"
            className={`${btnSuccess}${autoCompleteReady && !busy && !isAutoCompleting ? ' animate-pulse ring-2 ring-ds-success' : ''}`}
            onClick={handleAutoComplete}
            disabled={busy || isAutoCompleting || !autoCompleteReady}
            data-testid="autocomplete-button"
            title={autoCompleteReady ? undefined : t('autoCompleteNotReady')}
          >
            {t('autoComplete')}
          </button>
          <button type="button" className={btnDanger} onClick={confirmGiveUpAction} disabled={busy || isAutoCompleting}>
            {t('giveup')}
          </button>
        </div>
      )}
    </>
  );
}

const sultanDefinition: GameDefinition<
  Awaited<ReturnType<typeof sultanApi.exec>>,
  SultanGame,
  Parameters<typeof sultanApi.exec>,
  SultanView
> = {
  key: 'sultan',
  useGame: useSultanGame,
  useView: useSultanView,
  tutorial: SULTAN_TUTORIAL_STEPS,
  skeleton: { kind: 'tableau', topRow: 10, tableau: 8 },
  cli: () => ({ parseCommand: parseSultanCommand, formatResponse: formatSultanState, helpText: SULTAN_HELP }),
  phase: ({ state, view }) => ({
    name: view.phaseName,
    ended: state.phase === SultanPhase.GAME_CLEAR || state.phase === SultanPhase.GAME_OVER,
    won: state.phase === SultanPhase.GAME_CLEAR,
  }),
  landscapeBanner: true,
  contentClassName: 'pt-3 px-2 sm:px-4 lg:px-8',
  resetTutorialId: 'sultan-reset-button',
  // Give-up is irreversible, so route both the button and the `g` key through
  // the confirm dialog — matching reset's guard (issue #2099).
  shortcuts: ({ game, confirmGiveUpAction }) => [
    { key: 'd', action: game.handleDraw, label: 'draw' },
    { key: 'h', action: game.handleHint, label: 'hint' },
    { key: 'a', action: game.handleAutoComplete, label: 'autoComplete' },
    { key: 'g', action: confirmGiveUpAction ?? (() => undefined), label: 'giveUp' },
    { key: 'z', action: game.handleUndo, label: 'undo' },
  ],
  shortcutsEnabled: ({ game }) => game.state?.phase === SultanPhase.PLAYING,
  shortcutsTestId: 'sultan-kbd-shortcuts',
  settingsPlacement: 'bottom',
  Board,
  HeaderExtra,
  FooterActions,
};

/** Renders the Sultan of Turkey solitaire game page with foundations, divan reserve, and stock/waste. */
export const SultanPage = createGamePage(sultanDefinition);
