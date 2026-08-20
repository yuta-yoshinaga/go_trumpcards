import { useCallback, useMemo, useState } from 'react';
import { tusacApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCard } from '../components/motion/AnimatedCard';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useActionKeyboardNav } from '../hooks/useActionKeyboardNav';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { useMountReset } from '../hooks/useMountReset';
import { btnPrimary, btnSecondary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { TuSacResponse } from '../types/card';
import { TuSacPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseTuSacCommand, TUSAC_CLI_HELP } from '../utils/cli/commands/tusacCommands';
import { formatTuSacState } from '../utils/cli/formatters/tusacFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const TUSAC_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="tusac-hand"]', messageKey: 'tutorial.deck', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="tusac-seats"]', messageKey: 'tutorial.meld', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="tusac-actions"]', messageKey: 'tutorial.score', placement: 'top', advanceOn: 'next' },
];

/** Meld i18n keys, matching the Go domain's kind order. */
const MELD_KEYS = ['meld.sameColorSet', 'meld.sameColorSet', 'meld.chariotTrio', 'meld.soldierSet'] as const;

/**
 * Kinds shown in the points table, in the order the rules introduce them.
 *
 * Index 0 is "not a meld", so it never appears here.
 */
const MELD_KINDS = [1, 2, 3] as const;

/** Renders the Tu Sac game page (#5281). */
export const TuSacPage = withTutorial(TuSacPageContent, 'tusac', TUSAC_TUTORIAL_STEPS);

function TuSacPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('tusac');

  // **選ぶのは手札の位置。** 同じ色・同じ駒が 4 枚あるので、札そのものでは
  // どの 1 枚か決まらない。
  const [selected, setSelected] = useState<number[]>([]);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(tusacApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('tusac');
  const cliConfig: CliGameConfig<TuSacResponse, Parameters<typeof tusacApi.exec>> = useMemo(
    () => ({
      gameName: 'tusac',
      parseCommand: parseTuSacCommand,
      formatResponse: formatTuSacState,
      helpText: TUSAC_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isDraw = phase === TuSacPhase.DRAW;
  const isDiscard = phase === TuSacPhase.DISCARD;
  const isRoundEnd = phase === TuSacPhase.ROUND_END;
  const gameOver = !!state?.gameEndFlag;
  const canAct = !!state?.isHumanTurn && !gameOver;

  const clearSelection = useCallback(() => setSelected([]), []);

  const toggleCard = useCallback((index: number) => {
    setSelected((prev) => (prev.includes(index) ? prev.filter((i) => i !== index) : [...prev, index]));
  }, []);

  const handleDraw = useCallback(
    (fromDiscard: boolean) => {
      clearSelection();
      return execApi(fromDiscard ? 'take' : 'draw');
    },
    [execApi, clearSelection],
  );

  const handleMeld = useCallback(async () => {
    await execApi('meld', { indexes: selected });
    clearSelection();
  }, [execApi, selected, clearSelection]);

  const handleDiscard = useCallback(async () => {
    if (selected.length === 0) return;
    await execApi('discard', { index: selected[0] });
    clearSelection();
  }, [execApi, selected, clearSelection]);

  const actionBindings = useMemo(
    () => [{ key: 'n', action: () => execApi('next'), enabled: isRoundEnd && !gameOver }],
    [execApi, isRoundEnd, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('tusac', state);

  if (!state) return <GameSkeleton gameKey="tusac" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [TuSacPhase.DRAW]: t('phase.draw'),
      [TuSacPhase.DISCARD]: t('phase.discard'),
      [TuSacPhase.ROUND_END]: t('phase.roundEnd'),
      [TuSacPhase.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const human = state.seats[state.humanSeat];
  const humanWon = gameOver && state.winnerSeat === state.humanSeat;

  return (
    <GamePageShell
      title={tc('nav.tusac')}
      gameThemeBg={gameTheme.tusac.bg}
      phaseName={phaseName}
      gamePath="/tusac"
      gameEndFlag={gameOver}
      isHumanTurn={state.isHumanTurn}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="tusac-score">{t('label.score', { score: human?.score ?? 0 })}</span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div data-testid="card-area" className={`overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="tusac-round">
              {t('label.round', { round: state.roundNumber, total: state.rounds })}
              {' · '}
              <span data-testid="tusac-stock">{t('label.stock', { count: state.stockCount })}</span>
            </div>

            <p className="text-ds-text-muted text-center text-xs mb-1" data-testid="tusac-notice">
              {t('notice')}
            </p>

            {/* **5 枚の卒を揃える価値は、狙う前に知りたい** (#5784)。点数は
                サーバが送る meldPointsByKind から出す——写した表は、確率から
                導いた点数が変わったときに片方だけ嘘になる。 */}
            <p className="text-ds-text-muted text-center text-xs mb-3" data-testid="tusac-meld-points">
              {MELD_KINDS.map(
                (kind) =>
                  `${t(MELD_KEYS[kind] ?? 'meld.sameColorSet')} ${t('meld.points', { points: state.meldPointsByKind[kind] ?? 0 })}`,
              ).join(' · ')}
            </p>

            {/* **捨て札の一番上は拾う判断の材料。** 何が取れるかを出す。 */}
            <div className="mb-3">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.discard')}</div>
              <div className="flex justify-center" data-testid="tusac-discard-top">
                {state.discardTop ? (
                  <AnimatedCard card={state.discardTop} width={cardWidth} />
                ) : (
                  <span className="text-ds-text-muted text-xs">{t('label.discardEmpty')}</span>
                )}
              </div>
            </div>

            <div data-tutorial="tusac-seats">
              {state.seats.map((seat, i) => (
                <div
                  key={`seat-${seat.name}-${i}`}
                  data-testid={`tusac-seat-${i}`}
                  className={`mb-2 rounded px-2 py-1 text-sm ${seat.isTurn ? 'ring-2 ring-ds-success' : ''}`}
                >
                  <span className="text-ds-text-primary">
                    {seat.name}
                    {' · '}
                    {/* **相手の手札は届いていない。** 枚数だけが分かる。 */}
                    <span data-testid={`tusac-count-${i}`}>{t('label.handCount', { count: seat.handCount })}</span>
                    {' · '}
                    {t('label.score', { score: seat.score })}
                    {seat.wentOut && (
                      <span className="text-ds-success" data-testid={`tusac-wentout-${i}`}>
                        {' · '}
                        {t('label.wentOut')}
                      </span>
                    )}
                  </span>
                  {/* **場に出た組み合わせは全員ぶん見える。** ここから読む。 */}
                  <div className="mt-1" data-testid={`tusac-melds-${i}`}>
                    {seat.melds.length === 0 ? (
                      <span className="text-ds-text-muted text-xs">{t('label.noMelds')}</span>
                    ) : (
                      seat.melds.map((meld, k) => (
                        <div key={`m${i}-${k}-${meld.kind}`} className="flex items-center gap-1 flex-wrap mb-1">
                          <span className="text-ds-text-muted text-xs">
                            {t(MELD_KEYS[meld.kind] ?? 'meld.sameColorSet')} {t('meld.points', { points: meld.points })}
                          </span>
                          {meld.cards.map((card, c) => (
                            <AnimatedCard
                              key={`m${i}-${k}-${c}-${card.value}`}
                              card={card}
                              width={Math.round(cardWidth * 0.6)}
                            />
                          ))}
                        </div>
                      ))
                    )}
                  </div>
                  {isRoundEnd && (
                    <div className="text-ds-text-muted text-xs" data-testid={`tusac-roundscore-${i}`}>
                      {t('label.roundScore', {
                        meld: seat.meldPoints,
                        penalty: seat.handCount,
                        score: seat.roundScore,
                      })}
                    </div>
                  )}
                </div>
              ))}
            </div>

            {/* **手札は位置で選ぶ。** 押した札の番号がそのままワイヤに乗る。 */}
            <div className="mb-3" data-tutorial="tusac-hand">
              <div className="text-ds-text-primary text-center text-sm font-bold mb-1">
                {t('label.yourHand')}
                {selected.length > 0 && (
                  <span data-testid="tusac-selected">
                    {' · '}
                    {t('label.selected', { count: selected.length })}
                  </span>
                )}
              </div>
              <div className="flex justify-center gap-1 flex-wrap" data-testid="tusac-hand">
                {(human?.cards ?? []).map((card, i) => (
                  <button
                    key={`hand-${i}-${card.design}-${card.value}`}
                    type="button"
                    data-testid={`tusac-card-${i}`}
                    aria-pressed={selected.includes(i)}
                    className={`rounded transition-transform ${selected.includes(i) ? '-translate-y-2 ring-2 ring-ds-success' : ''}`}
                    onClick={() => toggleCard(i)}
                    disabled={loading || !canAct}
                  >
                    <AnimatedCard card={card} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            {gameOver && (
              <div className="text-ds-text-primary text-center text-sm font-bold mt-2" data-testid="tusac-winner">
                {t('label.winner')}: {state.seats[state.winnerSeat]?.name ?? '?'}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.tusac.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="tusac-actions">
              {/*
                **引く場面と出す・捨てる場面でボタンが入れ替わる。** サーバの
                フェーズに従う ── 引く前に捨てられる見た目にすると、押しても
                通らない操作を見せることになる。
              */}
              {canAct && isDraw && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="tusac-draw-guide">
                    {t('label.drawPrompt')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-testid="tusac-draw"
                      data-hint-action="draw"
                      onClick={() => handleDraw(false)}
                      disabled={loading}
                    >
                      {t('button.draw')}
                    </button>
                    <button
                      type="button"
                      className={btnWarning}
                      data-testid="tusac-take"
                      data-hint-action="draw"
                      onClick={() => handleDraw(true)}
                      disabled={loading || !state.discardTop}
                    >
                      {t('button.take')}
                    </button>
                  </div>
                </>
              )}

              {canAct && isDiscard && (
                <>
                  <p className="text-ds-text-muted text-sm" data-testid="tusac-discard-guide">
                    {t('label.discardPrompt')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-testid="tusac-meld"
                      data-hint-action="meld"
                      onClick={handleMeld}
                      disabled={loading || selected.length === 0}
                    >
                      {t('button.meld')}
                    </button>
                    <button
                      type="button"
                      className={btnWarning}
                      data-testid="tusac-discard"
                      data-hint-action="discard"
                      onClick={handleDiscard}
                      disabled={loading || selected.length !== 1}
                    >
                      {t('button.discard')}
                    </button>
                  </div>
                </>
              )}

              {isRoundEnd && !gameOver && (
                <button
                  type="button"
                  className={btnPrimary}
                  data-testid="tusac-next"
                  data-hint-action="next"
                  onClick={() => execApi('next')}
                  disabled={loading}
                >
                  {t('button.next')}
                </button>
              )}

              <div className="flex gap-2">
                <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                  {tc('button.actionLog')}
                </button>
                <GameResetButton
                  isGameEnd={gameOver}
                  onReset={() => execApi('reset')}
                  requestConfirm={requestConfirm}
                  loading={loading}
                />
              </div>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
