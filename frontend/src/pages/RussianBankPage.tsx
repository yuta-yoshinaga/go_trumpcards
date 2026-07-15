import { useEffect, useState } from 'react';
import { russianbankApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { Card, RussianBankPlayer } from '../types/card';
import { RussianBankPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';

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

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  const phaseNames = usePhaseNames('russianbank', RB_PHASE_KEYS);
  const { cardWidth } = useCardDimensions();
  const w = Math.round(cardWidth * 0.62);

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
    exec('reset');
  };

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
  ) => {
    const cls = `rounded ${highlighted ? 'ring-2 ring-ds-warning' : ''} ${onClick ? 'cursor-pointer' : ''}`;
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
          <div className="flex gap-1 flex-wrap mt-0.5">
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
          <div className="flex gap-2 flex-wrap mt-0.5">
            {state.tableau.map((col, i) => (
              <button
                type="button"
                key={`tab-${i}`}
                className={`rounded ${selected ? 'ring-1 ring-ds-success' : ''} ${canAct ? 'cursor-pointer' : ''}`}
                onClick={canAct ? () => sendToTableau(i) : undefined}
                disabled={!canAct}
                data-testid={`tableau-${i}`}
                aria-label={
                  col.length > 0
                    ? t('slotCardZone', { card: cardAlt(col[col.length - 1]), zone: t('srcTableau', { col: i + 1 }) })
                    : t('slotEmptyZone', { zone: t('srcTableau', { col: i + 1 }) })
                }
              >
                {col.length > 0 ? (
                  <CardImage card={col[col.length - 1]} width={w} />
                ) : (
                  <div
                    className="rounded border border-dashed border-white/25 bg-black/20"
                    style={{ width: w, height: Math.round(w * 1.4) }}
                  />
                )}
              </button>
            ))}
          </div>
        </div>

        {/* Players */}
        <div className="grid gap-2 sm:grid-cols-2">
          {cpu && renderPlayer(cpu, true)}
          {human && renderPlayer(human, false)}
        </div>

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

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
