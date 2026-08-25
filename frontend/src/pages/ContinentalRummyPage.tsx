import { useCallback, useMemo } from 'react';
import { continentalrummyApi } from '../api/gameApi';
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
import { PlayerHandSection } from '../components/PlayerHandSection';
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
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ContinentalRummyResponse } from '../types/card';
import { CONTINENTAL_RUMMY_PHASE } from '../types/games/continentalrummy';
import type { TutorialStep } from '../types/tutorial';
import {
  CONTINENTALRUMMY_CLI_HELP,
  parseContinentalRummyCommand,
} from '../utils/cli/commands/continentalrummyCommands';
import { formatContinentalRummyState } from '../utils/cli/formatters/continentalrummyFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const CONT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="continentalrummy-layouts"]',
    messageKey: 'tutorial.layouts',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="continentalrummy-player-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="continentalrummy-controls"]',
    messageKey: 'tutorial.draw',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Continental Rummy game page (#5464). */
export const ContinentalRummyPage = withTutorial(ContinentalRummyPageContent, 'continentalrummy', CONT_TUTORIAL_STEPS);

function ContinentalRummyPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('continentalrummy');

  const { cardWidth, isMobile } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(continentalrummyApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('continentalrummy');
  const cliConfig: CliGameConfig<ContinentalRummyResponse, Parameters<typeof continentalrummyApi.exec>> = useMemo(
    () => ({
      gameName: 'continentalrummy',
      parseCommand: parseContinentalRummyCommand,
      formatResponse: formatContinentalRummyState,
      helpText: CONTINENTALRUMMY_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const gameOver = !!state?.gameEndFlag;
  const myTurn = !!state?.isHumanTurn;
  const canDraw = myTurn && phase === CONTINENTAL_RUMMY_PHASE.DRAW;
  const canDiscard = myTurn && phase === CONTINENTAL_RUMMY_PHASE.DISCARD;
  const isRoundEnd = phase === CONTINENTAL_RUMMY_PHASE.ROUND_END && !gameOver;
  // **上がれるかはサーバが解いた答えをそのまま使う。** 15 枚の分割問題を
  // ページ側で解き直すと、規則が 2 か所に増えてどこかで食い違う。
  const goOutIdx = state?.goOutIdx ?? -1;

  const handleGoOut = useCallback(() => {
    if (goOutIdx >= 0) execApi('goout', { handIndex: goOutIdx });
  }, [execApi, goOutIdx]);

  const actionBindings = useMemo(
    () => [
      { key: 's', action: () => execApi('stock'), enabled: canDraw },
      { key: 't', action: () => execApi('take'), enabled: canDraw && !!state?.discardTop },
      { key: 'g', action: handleGoOut, enabled: canDiscard && goOutIdx >= 0 },
      { key: 'n', action: () => execApi('next'), enabled: isRoundEnd },
    ],
    [execApi, handleGoOut, canDraw, canDiscard, goOutIdx, isRoundEnd, state?.discardTop],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('continentalrummy', state);

  if (!state)
    return (
      <GameSkeleton
        gameKey="continentalrummy"
        layout={{ kind: 'card-grid', count: 4, cols: 'grid-cols-2 sm:grid-cols-4', topPills: 4, footerHandSize: 15 }}
      />
    );

  const phaseName =
    {
      [CONTINENTAL_RUMMY_PHASE.DRAW]: t('phase.draw'),
      [CONTINENTAL_RUMMY_PHASE.DISCARD]: t('phase.discard'),
      [CONTINENTAL_RUMMY_PHASE.ROUND_END]: t('phase.roundEnd'),
      [CONTINENTAL_RUMMY_PHASE.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const me = state.players.find((p) => p.isHuman);
  // 捨てられるのは自分の番だけ。番でなければ 1 枚も押せない。
  const legalIndices = canDiscard ? (me?.cards ?? []).map((_, i) => i) : [];

  return (
    <GamePageShell
      title={tc('nav.continentalrummy')}
      gameThemeBg={gameTheme.continentalrummy.bg}
      phaseName={phaseName}
      gamePath="/continentalrummy"
      gameEndFlag={gameOver}
      isHumanTurn={myTurn}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="cont-round">
            {t('label.round')}: {state.roundNumber}/{state.totalRounds}
          </span>
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

            {/* **上がれる形は常に見えていること。** 15 枚をどう割るかが全部で、
                5+5+5 がそこに無いのが肝。サーバが返した並びだけを出す。 */}
            <div
              className="text-ds-text-primary text-center text-sm mb-1"
              data-tutorial="continentalrummy-layouts"
              data-testid="cont-layouts"
            >
              {t('label.layouts')}: {state.layouts.map((l) => l.join('+')).join(' / ')}
            </div>
            <div className="text-ds-text-muted text-center text-xs mb-3" data-testid="cont-nosets">
              {t('noSets')}
            </div>

            <div className="flex items-center justify-center gap-6 mb-4" data-testid="cont-table">
              <div className="text-center">
                <div className="text-ds-text-muted text-xs mb-1">{t('label.stock')}</div>
                <div className="text-ds-text-primary text-lg font-bold" data-testid="cont-stock">
                  {state.stockCount}
                </div>
              </div>
              <div className="text-center">
                <div className="text-ds-text-muted text-xs mb-1">{t('label.discard')}</div>
                <div data-testid="cont-discard-top">
                  {state.discardTop ? (
                    <AnimatedCard card={state.discardTop} width={cardWidth} />
                  ) : (
                    <span className="text-ds-text-muted text-xs">—</span>
                  )}
                </div>
              </div>
            </div>

            <div className="grid grid-cols-2 sm:grid-cols-4 gap-2 mb-4" data-testid="cont-seats">
              {state.players.map((p) => (
                <div
                  key={`seat-${p.id}`}
                  data-testid={`cont-seat-${p.id}`}
                  className={`rounded border px-2 py-1 text-xs ${
                    p.id === state.currentPlayerIdx && !gameOver ? 'border-ds-success' : 'border-ds-border'
                  }`}
                >
                  <div className="font-bold text-ds-text-primary">
                    {p.isHuman ? t('label.you') : t('label.cpu', { n: p.id })}
                    {p.isDealer && ' ★'}
                  </div>
                  <div className="text-ds-text-muted">
                    {t('label.cards', { n: p.cardCount })} · {t('label.score')}: {p.score}
                  </div>
                  {p.melds.map((run, ri) => (
                    <div
                      key={`meld-${p.id}-${run.map((c) => `${c.design}${c.value}`).join('')}`}
                      className="flex gap-0.5 mt-1"
                      data-testid={`cont-meld-${p.id}-${ri}`}
                    >
                      {run.map((card, ci) => (
                        <AnimatedCard
                          key={`${card.design}-${card.value}-${ci}`}
                          card={card}
                          width={Math.round(cardWidth * 0.5)}
                        />
                      ))}
                    </div>
                  ))}
                </div>
              ))}
            </div>

            {/* **加点は内訳で見せる。** 合計だけだと、どう上がると得なのかが伝わらない。 */}
            {state.lastResult && (
              <div className="mb-3" data-testid="cont-result">
                {state.lastResult.winnerIdx < 0 ? (
                  <div className="text-ds-text-muted text-center text-sm">{t('washout')}</div>
                ) : (
                  <>
                    <div className="text-ds-text-primary text-center text-base font-bold mb-1">
                      {t('wentOut', {
                        name:
                          state.lastResult.winnerIdx === 0
                            ? t('label.you')
                            : t('label.cpu', { n: state.lastResult.winnerIdx }),
                      })}
                    </div>
                    {state.lastResult.bonuses.map((b) => (
                      <div
                        key={`bonus-${b.key}`}
                        className="text-ds-text-primary text-center text-sm"
                        data-testid={`cont-bonus-${b.key}`}
                      >
                        {t(`bonus.${b.key}`)}: {b.points}
                      </div>
                    ))}
                    <div className="text-ds-success text-center text-sm font-medium mt-1" data-testid="cont-collected">
                      {t('collected', { per: state.lastResult.perOpponent, total: state.lastResult.total })}
                    </div>
                  </>
                )}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.continentalrummy.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {me && (
              <PlayerHandSection
                humanPlayer={me}
                selectedCardIndices={[]}
                toggleCard={(i) => execApi('discard', { handIndex: i })}
                cardWidth={cardWidth}
                isMobile={isMobile}
                dataTutorialPrefix="continentalrummy"
                validIndices={canDiscard ? legalIndices : undefined}
                legalIndices={canDiscard ? legalIndices : undefined}
                restrictedTooltip={t('restrictedTooltip')}
              />
            )}

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="continentalrummy-controls">
              {canDraw && (
                <div className="flex gap-2">
                  <button
                    type="button"
                    className={btnPrimary}
                    data-hint-action="stock"
                    aria-keyshortcuts="s"
                    onClick={() => execApi('stock')}
                    disabled={loading}
                  >
                    {t('button.stock')}
                  </button>
                  <button
                    type="button"
                    className={btnSecondary}
                    data-hint-action="take"
                    aria-keyshortcuts="t"
                    onClick={() => execApi('take')}
                    disabled={loading || !state.discardTop}
                  >
                    {t('button.take')}
                  </button>
                </div>
              )}

              {/* **上がれるときは黙っていない。** 15 枚の分割は目で追いきれない
                  ので、見落としたまま捨ててしまうのが一番つまらない負け方になる。 */}
              {canDiscard && goOutIdx >= 0 && (
                <button
                  type="button"
                  className={btnSuccess}
                  data-hint-action="goout"
                  aria-keyshortcuts="g"
                  onClick={handleGoOut}
                  disabled={loading}
                  data-testid="cont-goout"
                >
                  {t('button.goOut')}
                </button>
              )}

              {canDiscard && goOutIdx < 0 && (
                <p className="text-ds-text-muted text-sm" data-testid="cont-discard-notice">
                  {t('discardNotice')}
                </p>
              )}

              {isRoundEnd && (
                <button
                  type="button"
                  className={btnPrimary}
                  aria-keyshortcuts="n"
                  onClick={() => execApi('next')}
                  disabled={loading}
                >
                  {t('button.next')}
                </button>
              )}

              {!myTurn && !isRoundEnd && !gameOver && (
                <p className="text-ds-text-muted text-sm" data-testid="cont-wait-notice">
                  {t('waitNotice')}
                </p>
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
