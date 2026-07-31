import { useMemo, useState } from 'react';
import type { desmocheApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardBack } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { LandscapeBanner } from '../components/LandscapeBanner';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useDesmocheGame } from '../hooks/useDesmocheGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { gameTheme } from '../styles/gameTheme';
import type { DesmocheResponse } from '../types/card';
import { DesmochePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { DESMOCHE_HELP, parseDesmocheCommand } from '../utils/cli/commands/desmocheCommands';
import { formatDesmocheState } from '../utils/cli/formatters/desmocheFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const DS_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="ds-rule"]', messageKey: 'tutorial.goOut', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ds-rule"]', messageKey: 'tutorial.noPoker', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ds-melds"]', messageKey: 'tutorial.desmoche', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="ds-seats"]', messageKey: 'tutorial.pot', placement: 'bottom', advanceOn: 'next' },
];

/** Renders the Desmoche page: nine dealt, ten melded takes the pot. */
export const DesmochePage = withTutorial(DesmochePageContent, 'desmoche', DS_TUTORIAL_STEPS);

function DesmochePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('desmoche');
  const game = useDesmocheGame();
  const { state, loading, error, retry } = game;

  // Which hand cards are selected. A meld is several cards at once, so
  // selection has to be explicit rather than click-to-play.
  const [selected, setSelected] = useState<number[]>([]);
  const [meldTarget, setMeldTarget] = useState<number | null>(null);
  // The desmoche move needs a card *inside* a meld, which is a different
  // selection from the hand: {meld index, card index within it}.
  const [meldCard, setMeldCard] = useState<{ meld: number; card: number } | null>(null);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('desmoche');
  const cliConfig: CliGameConfig<DesmocheResponse, Parameters<typeof desmocheApi.exec>> = useMemo(
    () => ({
      gameName: 'desmoche',
      parseCommand: parseDesmocheCommand,
      formatResponse: formatDesmocheState,
      helpText: DESMOCHE_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(game.exec, cliConfig, state, { addInput, addOutput, addError, clearLog });
  const { cardWidth } = useCardDimensions();
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('desmoche', state);

  if (!state) {
    return <GameSkeleton gameKey="desmoche" layout={{ kind: 'tableau', topRow: 3, tableau: 4 }} />;
  }

  const ended = state.phase === DesmochePhase.GAME_END;
  const drawing = state.phase === DesmochePhase.DRAW;
  const acting = state.phase === DesmochePhase.ACT;
  const roundOver = state.phase === DesmochePhase.ROUND_END;
  const human = state.players.find((p) => p.isHuman);
  const opponents = state.players.filter((p) => !p.isHuman);
  const isHumanTurn = !ended && state.currentPlayerIdx === 0;

  const toggleCard = (i: number) => {
    setSelected((prev) => (prev.includes(i) ? prev.filter((n) => n !== i) : [...prev, i]));
  };

  const meldKindName = (kind: number) => (kind === 0 ? t('set') : t('run'));

  const phaseName = ended
    ? t('phase.end')
    : roundOver
      ? t('phase.roundEnd')
      : drawing
        ? t('phase.draw')
        : t('phase.act');

  return (
    <GamePageShell
      title={tc('nav.desmoche')}
      gameThemeBg={gameTheme.desmoche.bg}
      phaseName={phaseName}
      gamePath="/desmoche"
      gameEndFlag={ended}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span className="text-sm text-ds-text-muted">
            {t('round')}: {state.roundNo + 1} / {t('stock')}: {state.stockCount} / {t('pot', { n: state.pot })}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      <LandscapeBanner message={t('landscapeBanner')} />

      <SettingsPanel
        title={tc('settings.title')}
        groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
      />

      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className="flex-1 overflow-y-auto pt-3 px-2 sm:px-4 lg:px-8">
            {/* Permanent, not tutorial-only: "exactly ten, not the nine dealt"
                and "poker rankings play no part" are what a player gets wrong. */}
            <div className="text-center text-xs text-ds-warning mb-3 font-medium" data-tutorial="ds-rule">
              {t('ruleLine', { n: state.goOutSize })}
            </div>

            <div className="flex justify-center gap-4 mb-3 flex-wrap" data-tutorial="ds-seats">
              {opponents.map((o) => (
                <div key={`opp-${o.id.toString()}`} className="text-center">
                  <div className="text-game-text-muted text-xs mb-1">
                    {t('opponentHand', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                    {' · '}
                    {t('melded', { n: o.meldedCount, goal: state.goOutSize })}
                  </div>
                  <div
                    className="flex gap-1 justify-center flex-wrap"
                    role="img"
                    aria-label={t('opponentHandAriaLabel', { name: `CPU${o.id.toString()}`, n: o.cardCount })}
                  >
                    {Array.from({ length: o.cardCount }, (_, i) => (
                      <CardBack key={`opp-${o.id.toString()}-c${i.toString()}`} width={cardWidth} />
                    ))}
                  </div>
                </div>
              ))}
            </div>

            <div className="text-center mb-3">
              <div className="text-game-text-muted text-xs mb-1">{t('discard')}</div>
              {state.discardTop ? (
                <div className="flex justify-center">
                  <AnimatedCard card={state.discardTop} width={cardWidth} draggable={false} />
                </div>
              ) : (
                <span className="text-game-text-muted text-xs">—</span>
              )}
            </div>

            <div className="mb-4" data-tutorial="ds-melds">
              <div className="text-game-text-muted text-xs mb-1 text-center">{t('melds')}</div>
              {state.melds.length === 0 ? (
                <div className="text-center text-game-text-muted text-xs">{t('noMelds')}</div>
              ) : (
                <div className="flex flex-col gap-1 items-center">
                  {state.melds.map((m, i) => (
                    <div
                      key={`meld-${i.toString()}`}
                      className={[
                        'flex items-center gap-1 rounded px-2 py-1',
                        meldTarget === i ? 'ring-2 ring-ds-accent' : '',
                      ].join(' ')}
                    >
                      <button
                        type="button"
                        data-testid="desmoche-meld"
                        aria-pressed={meldTarget === i}
                        onClick={() => setMeldTarget(meldTarget === i ? null : i)}
                        className="text-[10px] text-ds-text-muted min-h-11 px-1"
                      >
                        {meldKindName(m.kind)} · {t('meldOwner', { n: m.owner })}
                      </button>
                      {m.cards.map((card, j) => (
                        <button
                          key={`meld-${i.toString()}-c${j.toString()}`}
                          type="button"
                          data-testid="desmoche-meld-card"
                          aria-pressed={meldCard?.meld === i && meldCard.card === j}
                          aria-disabled={m.owner !== 0}
                          onClick={() =>
                            m.owner === 0 &&
                            setMeldCard(meldCard?.meld === i && meldCard.card === j ? null : { meld: i, card: j })
                          }
                          className={[
                            'rounded min-h-11',
                            meldCard?.meld === i && meldCard.card === j ? 'ring-2 ring-ds-warning' : '',
                          ].join(' ')}
                        >
                          <AnimatedCard card={card} width={Math.round(cardWidth * 0.7)} draggable={false} />
                        </button>
                      ))}
                    </div>
                  ))}
                </div>
              )}
            </div>

            {roundOver && (
              <div className="text-center text-sm mb-3" data-testid="desmoche-round-result">
                {state.roundWinner >= 0
                  ? t('roundWinner', { n: state.roundWinner })
                  : t('roundNoWinner', { n: state.pot })}
              </div>
            )}

            <div className="text-center" data-tutorial="ds-hand">
              <div className="text-game-text-muted text-xs mb-1">
                {t('yourHand')}
                {' · '}
                {t('melded', { n: human?.meldedCount ?? 0, goal: state.goOutSize })}
                {selected.length > 0 && ` · ${t('selected', { n: selected.length })}`}
              </div>
              <div className="flex gap-1 justify-center flex-wrap">
                {(human?.cards ?? []).map((card, i) => (
                  <button
                    key={`hand-${i.toString()}`}
                    type="button"
                    data-hint-action="discard"
                    aria-pressed={selected.includes(i)}
                    aria-disabled={!isHumanTurn || !acting}
                    onClick={() => isHumanTurn && acting && toggleCard(i)}
                    className={[
                      'rounded transition-transform',
                      isHumanTurn && acting ? 'hover:-translate-y-2' : 'opacity-60',
                      selected.includes(i) ? 'ring-2 ring-ds-accent -translate-y-2' : '',
                      frontendHintEnabled && state.hint?.cardIndex === i ? 'ring-2 ring-ds-warning' : '',
                    ].join(' ')}
                  >
                    <AnimatedCard card={card} width={cardWidth} draggable={false} />
                  </button>
                ))}
              </div>
              {acting && isHumanTurn && <div className="text-[10px] text-ds-text-muted mt-1">{t('selectHint')}</div>}
              <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />
            </div>

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={ended}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          <GameFooter className={`${gameTheme.desmoche.footer} px-4 py-2.5`}>
            <ErrorAlert message={error} onRetry={retry} />
            <div className="flex gap-2 items-center flex-wrap">
              {drawing && isHumanTurn && (
                <>
                  <button
                    type="button"
                    data-hint-action="draw"
                    className={`${btnPrimary} min-h-11`}
                    onClick={game.handleDrawStock}
                  >
                    {t('drawStock')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-h-11`}
                    onClick={game.handleDrawDiscard}
                    disabled={!state.discardTop}
                  >
                    {t('drawDiscard')}
                  </button>
                </>
              )}
              {acting && isHumanTurn && (
                <>
                  <button
                    type="button"
                    data-hint-action="meld"
                    className={`${btnPrimary} min-h-11`}
                    disabled={selected.length < 3}
                    onClick={() => {
                      game.handleMeld(selected);
                      setSelected([]);
                    }}
                  >
                    {t('meld')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-h-11`}
                    disabled={selected.length !== 1 || meldTarget === null}
                    onClick={() => {
                      if (selected.length === 1 && meldTarget !== null) {
                        game.handleLayOff(selected[0], meldTarget);
                        setSelected([]);
                        setMeldTarget(null);
                      }
                    }}
                  >
                    {t('layOff')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-h-11`}
                    disabled={meldCard === null || meldTarget === null || meldCard.meld === meldTarget}
                    onClick={() => {
                      if (meldCard !== null && meldTarget !== null && meldCard.meld !== meldTarget) {
                        game.handleDesmoche(meldCard.meld, meldCard.card, meldTarget);
                        setMeldCard(null);
                        setMeldTarget(null);
                      }
                    }}
                  >
                    {t('desmocheBtn')}
                  </button>
                  <button
                    type="button"
                    className={`${btnSecondary} min-h-11`}
                    disabled={selected.length !== 1}
                    onClick={() => {
                      if (selected.length === 1) {
                        game.handleDiscard(selected[0]);
                        setSelected([]);
                      }
                    }}
                  >
                    {t('discardBtn')}
                  </button>
                </>
              )}
              {roundOver && !ended && (
                <button type="button" className={`${btnPrimary} min-h-11`} onClick={game.handleNextRound}>
                  {t('nextRound')}
                </button>
              )}
              <GameResetButton
                isGameEnd={ended}
                onReset={game.handleReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="ds-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
