import { useEffect, useRef, useState } from 'react';
import { russianbankApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, RussianBankPlayer } from '../types/card';
import { RussianBankPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Source-zone codes matching the Go `RussianBankZone` enum. */
const ZONE_RESERVE = 0;
const ZONE_WASTE = 1;
const ZONE_TABLEAU = 2;

/** A selected move source (the human picks a source, then a destination). */
interface SelectedSource {
  zone: number;
  fromOpp: boolean;
  col: number;
  label: string;
}

/** Russian Bank tutorial step definitions. */
const RB_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="rb-board"]', messageKey: 'tutorial.board', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="rb-controls"]', messageKey: 'tutorial.controls', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="rb-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Maps numeric Russian Bank phases to i18n phase-label keys. */
const RB_PHASE_KEYS: Readonly<Record<number, string>> = {
  [RussianBankPhase.PLAYING]: 'playing',
  [RussianBankPhase.GAME_END]: 'gameEnd',
};

/** Renders the Russian Bank (Crapette) game page. */
export const RussianBankPage = withTutorial(RussianBankPageContent, 'russianbank', RB_TUTORIAL_STEPS);

/** Inner content of the Russian Bank page, wrapped by TutorialProvider. */
function RussianBankPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('russianbank');
  const { state, loading, error, exec, retry } = useGameApi(russianbankApi.exec);
  const [selected, setSelected] = useState<SelectedSource | null>(null);
  // Whether the last hint's suggested move is currently highlighted on the board.
  // The move coordinates come from `state.hint` (set by the server on a `hint`
  // command); this flag just gates the rings so they only show after the player
  // asks for a hint, then auto-dismiss.
  const [showHint, setShowHint] = useState(false);
  const hintTimerRef = useRef<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // Clear any hint highlight (and stale source selection) whenever the board
  // changes (move, discard, undo, CPU turn) so the rings can't point at a
  // different card than the current hint.
  // biome-ignore lint/correctness/useExhaustiveDependencies: deps are the change-trigger, not read in the body.
  useEffect(() => {
    setSelected(null);
    setShowHint(false);
  }, [state?.moveCount]);

  // Cancel a pending hint auto-dismiss timer on unmount.
  useEffect(() => () => window.clearTimeout(hintTimerRef.current ?? undefined), []);

  const phaseNames = usePhaseNames('russianbank', RB_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const w = Math.round(cardWidth * 0.62);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('russianbank', state);

  if (!state) return <GameSkeleton gameKey="russianbank" layout={{ kind: 'tableau', topRow: 8, tableau: 4 }} />;

  const isGameEnd = state.phase === RussianBankPhase.GAME_END || state.gameEndFlag;
  const phaseName = phaseNames[state.phase] ?? '';
  const human = state.players.find((p) => p.isHuman) ?? null;
  const cpu = state.players.find((p) => !p.isHuman) ?? null;
  const humanWon = isGameEnd && human !== null && state.winnerIdx === human.id;
  const canAct = state.isHumanTurn && !isGameEnd;

  const handleReset = () => {
    hideActionLog();
    setSelected(null);
    // A reset keeps moveCount at 0, so the board-change effect may not fire —
    // clear any stale hint highlight here.
    setShowHint(false);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    exec('reset');
  };

  // Ask the server for a hint, then highlight the suggested move's source and
  // destination for a few seconds. Does not execute the move.
  const handleHint = () => {
    exec('hint');
    setShowHint(true);
    window.clearTimeout(hintTimerRef.current ?? undefined);
    hintTimerRef.current = window.setTimeout(() => setShowHint(false), 4000);
  };

  // The move to highlight, only while the highlight is active.
  const hint = showHint ? state.hint : undefined;
  const hintFoundation = hint?.toFoundation === true;
  // A hint was requested but the server found no legal move to suggest.
  const hintEmpty = showHint && state.hint === undefined;

  /** Pick (or toggle off) a move source. */
  const pickSource = (zone: number, fromOpp: boolean, col: number, label: string) => {
    if (!canAct) return;
    if (selected && selected.zone === zone && selected.fromOpp === fromOpp && selected.col === col) {
      setSelected(null);
      return;
    }
    setSelected({ zone, fromOpp, col, label });
  };

  const sendToFoundation = () => {
    if (!selected) return;
    exec('pf', { zone: selected.zone, fromOpp: selected.fromOpp, col: selected.col });
    setSelected(null);
  };

  const sendToTableau = (toCol: number) => {
    if (!selected) {
      // No source selected: a tableau column click selects that column as a source.
      pickSource(ZONE_TABLEAU, false, toCol, t('srcTableau', { col: toCol }));
      return;
    }
    exec('mt', { zone: selected.zone, fromOpp: selected.fromOpp, col: selected.col, toCol });
    setSelected(null);
  };

  /** Renders a single card (or a dashed empty slot). */
  const slot = (
    card: Card | undefined,
    key: string,
    onClick?: () => void,
    highlighted = false,
    zoneName?: string,
    // `source` slots (reserve/waste) are selectable move-origins → expose aria-pressed.
    source = false,
    // Hint rings: `hintSource` marks the suggested move's origin, `hintDest` its target.
    hintSource = false,
    hintDest = false,
  ) => {
    // Source ring (info) and destination ring (success) are additive hint cues that
    // sit on top of the selection ring (warning).
    const hintRing = hintSource
      ? ' ring-2 ring-ds-info motion-safe:animate-pulse'
      : hintDest
        ? ' ring-2 ring-ds-success motion-safe:animate-pulse'
        : '';
    const cls = `rounded ${highlighted ? 'ring-2 ring-ds-warning' : ''}${hintRing} ${onClick ? 'cursor-pointer' : ''}`;
    const label = card
      ? zoneName
        ? t('slotCardZone', { card: cardAlt(card), zone: zoneName })
        : cardAlt(card)
      : zoneName
        ? t('slotEmptyZone', { zone: zoneName })
        : t('empty');
    const ariaPressed = source ? highlighted : undefined;
    if (card) {
      return (
        <button
          type="button"
          key={key}
          className={cls}
          onClick={onClick}
          disabled={!onClick}
          data-testid={key}
          data-hint-source={hintSource ? 'true' : undefined}
          data-hint-dest={hintDest ? 'true' : undefined}
          aria-label={label}
          aria-pressed={ariaPressed}
        >
          <CardImage card={card} width={w} />
        </button>
      );
    }
    return (
      <button
        type="button"
        key={key}
        className={`${cls} border border-dashed border-white/25 bg-black/20`}
        style={{ width: w, height: Math.round(w * 1.4) }}
        onClick={onClick}
        disabled={!onClick}
        data-testid={key}
        data-hint-source={hintSource ? 'true' : undefined}
        data-hint-dest={hintDest ? 'true' : undefined}
        aria-label={label}
        aria-pressed={ariaPressed}
      />
    );
  };

  const playerName = (p: RussianBankPlayer) => (p.isHuman ? t('you') : t('cpu'));

  /** Renders a player's reserve / waste / hand summary. */
  const renderPlayer = (p: RussianBankPlayer, opponent: boolean) => (
    <div
      className={`p-2 rounded bg-black/20 ${p.id === state.currentPlayerIdx && !isGameEnd ? 'ring-1 ring-ds-warning' : ''}`}
      data-testid={`player-${p.id}`}
    >
      <div className="flex items-center justify-between mb-1">
        <span className="text-ds-text-primary text-sm font-semibold">{playerName(p)}</span>
        <span className="text-ds-text-muted text-xs">{t('stopPoints', { n: p.stopPoints })}</span>
      </div>
      <div className="flex gap-3 items-end">
        <div className="flex flex-col items-center gap-0.5">
          <span className="text-ds-text-muted text-[11px]">{t('reserve', { n: p.reserveCount })}</span>
          {slot(
            p.reserveTop,
            `reserve-${p.id}`,
            canAct
              ? () => pickSource(ZONE_RESERVE, opponent, 0, t(opponent ? 'srcOppReserve' : 'srcReserve'))
              : undefined,
            selected?.zone === ZONE_RESERVE && selected.fromOpp === opponent,
            t(opponent ? 'srcOppReserve' : 'srcReserve'),
            true,
            hint?.zone === ZONE_RESERVE && hint.fromOpponent === opponent,
          )}
        </div>
        <div className="flex flex-col items-center gap-0.5">
          <span className="text-ds-text-muted text-[11px]">{t('waste', { n: p.wasteCount })}</span>
          {slot(
            p.wasteTop,
            `waste-${p.id}`,
            canAct ? () => pickSource(ZONE_WASTE, opponent, 0, t(opponent ? 'srcOppWaste' : 'srcWaste')) : undefined,
            selected?.zone === ZONE_WASTE && selected.fromOpp === opponent,
            t(opponent ? 'srcOppWaste' : 'srcWaste'),
            true,
            hint?.zone === ZONE_WASTE && hint.fromOpponent === opponent,
          )}
        </div>
        <div className="flex flex-col items-center gap-0.5">
          <span className="text-ds-text-muted text-[11px]">{t('hand', { n: p.handCount })}</span>
          <div
            className="rounded border border-white/15 bg-ds-primary/40"
            style={{ width: w, height: Math.round(w * 1.4) }}
          />
        </div>
      </div>
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.russianbank')}
      gameThemeBg={gameTheme.russianbank.bg}
      phaseName={phaseName}
      gamePath="/russianbank"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      isHumanTurn={state.isHumanTurn}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <div className="flex-1 overflow-y-auto pt-3 px-3 lg:px-8" data-tutorial="rb-board">
        {/* Foundations */}
        <div className="mb-2">
          <span className="text-ds-text-muted text-[11px]">{t('foundationsLabel')}</span>
          <div
            className={`flex gap-1 flex-wrap mt-0.5 rounded${hintFoundation ? ' ring-2 ring-ds-success motion-safe:animate-pulse p-1' : ''}`}
            data-hint-foundation={hintFoundation ? 'true' : undefined}
          >
            {state.foundations.map((f, i) =>
              slot(
                f[f.length - 1],
                `foundation-${i}`,
                selected ? sendToFoundation : undefined,
                false,
                t('foundationZone', { n: i + 1 }),
              ),
            )}
          </div>
        </div>

        {/* Shared tableau */}
        <div className="mb-3">
          <span className="text-ds-text-muted text-[11px]">{t('tableauLabel')}</span>
          <div className="flex gap-2 flex-wrap mt-0.5 items-start">
            {state.tableau.map((col, i) => {
              const n = col.length;
              const cardH = Math.round(w * 1.4);
              const isHintSource = hint?.zone === ZONE_TABLEAU && hint.col === i;
              const isHintDest = hint !== undefined && !hintFoundation && hint.toCol === i;
              // Source ring (info) / destination ring (success) are additive hint cues,
              // layered on top of the shared move-target ring.
              const hintRing = isHintSource
                ? ' ring-2 ring-ds-info motion-safe:animate-pulse'
                : isHintDest
                  ? ' ring-2 ring-ds-success motion-safe:animate-pulse'
                  : '';
              // Reveal every buried card's rank+suit corner via a vertical cascade
              // instead of showing only the top card. The visible strip auto-compresses
              // for tall columns so the pile never overruns the board / footer (#3574).
              const baseStrip = Math.round(w * 0.34);
              const maxColH = Math.round(cardH * 2.8);
              const strip = n > 1 ? Math.min(baseStrip, Math.max(1, Math.round((maxColH - cardH) / (n - 1)))) : 0;
              return (
                <button
                  type="button"
                  key={`tab-${i}`}
                  className={`relative flex flex-col items-center rounded ${selected ? 'ring-1 ring-ds-success' : ''}${hintRing} ${canAct ? 'cursor-pointer' : ''}`}
                  onClick={canAct ? () => sendToTableau(i) : undefined}
                  disabled={!canAct}
                  data-testid={`tableau-${i}`}
                  data-hint-source={isHintSource ? 'true' : undefined}
                  data-hint-dest={isHintDest ? 'true' : undefined}
                  aria-label={
                    n > 0
                      ? t('slotCardZone', { card: cardAlt(col[n - 1]), zone: t('srcTableau', { col: i + 1 }) })
                      : t('slotEmptyZone', { zone: t('srcTableau', { col: i + 1 }) })
                  }
                >
                  {n === 0 ? (
                    <div
                      className="rounded border border-dashed border-white/25 bg-black/20"
                      style={{ width: w, height: cardH }}
                    />
                  ) : (
                    col.map((c, ci) => (
                      <div
                        key={`tab-${i}-card-${ci}`}
                        data-testid={`tableau-${i}-card-${ci}`}
                        style={{ marginTop: ci === 0 ? 0 : -(cardH - strip) }}
                      >
                        <CardImage card={c} width={w} />
                      </div>
                    ))
                  )}
                </button>
              );
            })}
          </div>
        </div>

        {/* Players */}
        <div className="grid gap-2 sm:grid-cols-2">
          {cpu && renderPlayer(cpu, true)}
          {human && renderPlayer(human, false)}
        </div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <SettingsPanel
          title={tc('settings.title')}
          groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
        />

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      {/* Footer: move controls */}
      <GameFooter className={`${gameTheme.russianbank.footer} px-3 py-2.5`}>
        <ErrorAlert message={error} onRetry={retry} />
        <div className="flex flex-wrap gap-2 items-center" data-tutorial="rb-controls">
          {isGameEnd && <span className="text-ds-text-primary text-sm font-semibold mr-1">{t('gameEnd')}</span>}

          {!isGameEnd && selected && (
            <>
              <span className="text-ds-text-primary text-xs" role="status" data-testid="rb-selected-source">
                {t('selectedSource', { src: selected.label })}
              </span>
              <button
                type="button"
                className={btnSuccess}
                onClick={sendToFoundation}
                disabled={loading}
                data-testid="to-foundation"
              >
                {t('toFoundation')}
              </button>
              <button
                type="button"
                className={btnSecondary}
                onClick={() => setSelected(null)}
                disabled={loading}
                data-testid="cancel-select"
              >
                {t('cancelSelect')}
              </button>
            </>
          )}

          {canAct && !selected && <span className="text-ds-text-muted text-xs">{t('pickSourceHint')}</span>}

          {canAct && (
            <button
              type="button"
              className={btnPrimary}
              onClick={() => exec('d')}
              disabled={loading}
              data-testid="discard-button"
            >
              {t('discard')}
            </button>
          )}
          {canAct && (
            <button
              type="button"
              className={btnSecondary}
              onClick={handleHint}
              disabled={loading}
              data-testid="hint-button"
            >
              {t('hint')}
            </button>
          )}
          {hintEmpty && (
            <span className="text-ds-text-muted text-xs" role="status" data-testid="rb-hint-none">
              {t('hintNone')}
            </span>
          )}
          {canAct && state.canCallStop && (
            <button
              type="button"
              className={btnWarning}
              onClick={() => exec('s')}
              disabled={loading}
              data-testid="stop-button"
            >
              {t('stop')}
            </button>
          )}
          {canAct && state.canUndo && (
            <button
              type="button"
              className={btnSecondary}
              onClick={() => exec('u')}
              disabled={loading}
              data-testid="undo-button"
            >
              {t('undo')}
            </button>
          )}

          <GameResetButton
            isGameEnd={isGameEnd}
            onReset={handleReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="rb-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
