import { useCallback, useMemo } from 'react';
import { baccaratbanqueApi } from '../api/gameApi';
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
import type { BaccaratBanquePlayer, BaccaratBanqueResponse } from '../types/card';
import { BACCARAT_BANQUE_PHASE } from '../types/games/baccaratbanque';
import type { TutorialStep } from '../types/tutorial';
import { BACCARATBANQUE_CLI_HELP, parseBaccaratBanqueCommand } from '../utils/cli/commands/baccaratbanqueCommands';
import { formatBaccaratBanqueState } from '../utils/cli/formatters/baccaratbanqueFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const BANQUE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="baccaratbanque-tableaux"]',
    messageKey: 'tutorial.tableaux',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="baccaratbanque-bank"]',
    messageKey: 'tutorial.bank',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="baccaratbanque-controls"]',
    messageKey: 'tutorial.draw',
    placement: 'top',
    advanceOn: 'next',
  },
];

/** Renders the Baccarat Banque game page (#5462). */
export const BaccaratBanquePage = withTutorial(BaccaratBanquePageContent, 'baccaratbanque', BANQUE_TUTORIAL_STEPS);

function BaccaratBanquePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('baccaratbanque');

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(baccaratbanqueApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('baccaratbanque');
  const cliConfig: CliGameConfig<BaccaratBanqueResponse, Parameters<typeof baccaratbanqueApi.exec>> = useMemo(
    () => ({
      gameName: 'baccaratbanque',
      parseCommand: parseBaccaratBanqueCommand,
      formatResponse: formatBaccaratBanqueState,
      helpText: BACCARATBANQUE_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  // **出せる操作はサーバの phase だけで決める。** 引き際の規則をページ側で
  // 組み直すと必ずドメインとずれる。とくに親はどの合計でも自由なので、
  // 合計から「引けるはず」を導いてはいけない。
  const phase = state?.phase;
  const gameOver = !!state?.gameEndFlag;
  const canDecide = !!state?.isHumanTurn && phase === BACCARAT_BANQUE_PHASE.BANKER && !gameOver;
  const isCoupEnd = phase === BACCARAT_BANQUE_PHASE.RESULT && !gameOver;

  const handleDraw = useCallback(() => execApi('draw'), [execApi]);
  const handleStand = useCallback(() => execApi('stand'), [execApi]);

  const actionBindings = useMemo(
    () => [
      { key: 'd', action: handleDraw, enabled: canDecide },
      { key: 's', action: handleStand, enabled: canDecide },
      { key: 'n', action: () => execApi('nextcoup'), enabled: isCoupEnd },
      { key: 'r', action: () => execApi('retire'), enabled: isCoupEnd },
    ],
    [handleDraw, handleStand, execApi, canDecide, isCoupEnd],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('baccaratbanque', state);

  if (!state) return <GameSkeleton gameKey="baccaratbanque" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [BACCARAT_BANQUE_PHASE.PUNTERS]: t('phase.punters'),
      [BACCARAT_BANQUE_PHASE.BANKER]: t('phase.banker'),
      [BACCARAT_BANQUE_PHASE.RESULT]: t('phase.result'),
      [BACCARAT_BANQUE_PHASE.GAME_END]: t('phase.gameEnd'),
    }[state.phase] ?? '';

  const banker = state.players.find((p) => p.role === 'banker');
  const tableaux = state.players.filter((p) => p.role !== 'banker');

  const seatBox = (p: BaccaratBanquePlayer, tutorial?: string) => (
    <div
      key={`seat-${p.id}`}
      data-testid={`banque-seat-${p.role}`}
      data-tutorial={tutorial}
      className="flex-1 min-w-0 rounded border border-ds-border px-2 py-2"
    >
      <div className="text-ds-text-primary text-center text-sm font-bold mb-1">
        {t(`role.${p.role}`)}
        {/* **ナチュラルはその席が打ち止めという意味。** 印が無いと、なぜ 3 枚目が
            無いのか盤面から読めない。 */}
        {p.cards.length === 2 && p.total >= 8 && <span data-testid={`banque-natural-${p.role}`}> ★</span>}
      </div>
      <div className="flex justify-center gap-1 flex-wrap" data-testid={`banque-hand-${p.role}`}>
        {p.cards.length === 0 ? (
          <span className="text-ds-text-muted text-xs">—</span>
        ) : (
          p.cards.map((card, i) => (
            <AnimatedCard key={`${p.role}-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
          ))
        )}
      </div>
      {p.cards.length > 0 && (
        <div className="text-ds-text-primary text-center text-lg font-bold mt-1" data-testid={`banque-total-${p.role}`}>
          {p.total}
        </div>
      )}
      <div className="text-ds-text-muted text-center text-xs mt-1">
        {t('label.chips')}: {p.chips}
        {p.bet > 0 && ` (${t('label.bet')}: ${p.bet})`}
      </div>
    </div>
  );

  return (
    <GamePageShell
      title={tc('nav.baccaratbanque')}
      gameThemeBg={gameTheme.baccaratbanque.bg}
      phaseName={phaseName}
      gamePath="/baccaratbanque"
      gameEndFlag={gameOver}
      isHumanTurn={canDecide}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="banque-my-chips">
            {t('label.chips')}: {banker?.chips ?? 0}
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

            {/* **バンクが何回続いているかを見せる。** 1 回負けても席が動かないのが
                この形式の要で、それは残高からは読み取れない。 */}
            <div className="text-ds-text-primary text-center text-sm mb-1" data-testid="banque-coup-line">
              {t('label.coup')}: {state.coupNumber} · {t('label.bankHeld', { n: state.bankHeld })}
            </div>
            <div className="text-ds-text-muted text-center text-xs mb-3" data-testid="banque-shoe-line">
              {t('label.shoe', { n: state.shoeRemaining })}
            </div>

            <div className="flex gap-3 mb-3 flex-wrap sm:flex-nowrap" data-tutorial="baccaratbanque-tableaux">
              {tableaux.map((p) => seatBox(p))}
            </div>

            <div className="flex mb-4" data-tutorial="baccaratbanque-bank">
              {banker && seatBox(banker)}
            </div>

            {/* **左右は別勘定。** 片方に払いながらもう片方から取るクーがあるので、
                1 行にまとめると何が起きたか読めない。 */}
            {state.lastResult && (
              <div className="mb-3" data-testid="banque-result">
                <div className="text-ds-text-primary text-center text-base font-bold mb-1">
                  {t('result.title', { total: state.lastResult.bankerTotal })}
                </div>
                {state.lastResult.sides.map((s) => (
                  <div
                    key={`side-${s.seatIdx}`}
                    className="text-ds-text-primary text-center text-sm"
                    data-testid={`banque-side-${s.seatIdx}`}
                  >
                    {t(`role.${s.seatIdx === 1 ? 'right' : 'left'}`)}: {t(`outcome.${s.outcome}`)} ({s.delta})
                  </div>
                ))}
                <div
                  className={`text-center text-sm font-medium mt-1 ${
                    state.lastResult.bankerDelta > 0
                      ? 'text-ds-success'
                      : state.lastResult.bankerDelta < 0
                        ? 'text-ds-error'
                        : 'text-ds-text-muted'
                  }`}
                  data-testid="banque-net"
                >
                  {t('result.bankerDelta', { n: state.lastResult.bankerDelta })}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.baccaratbanque.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="baccaratbanque-controls">
              {canDecide && (
                <>
                  {/* **固定表が無いことを毎回言う。** バカラで「引ける合計」を
                      覚えている人ほど、ここで表を探して手が止まる。 */}
                  <p className="text-ds-text-muted text-sm" data-testid="banque-free-notice">
                    {t('bankerGuide', { total: banker?.total ?? 0 })}
                  </p>
                  <div className="flex gap-2">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-hint-action="draw"
                      aria-keyshortcuts="d"
                      onClick={handleDraw}
                      disabled={loading}
                    >
                      {t('button.draw')}
                    </button>
                    <button
                      type="button"
                      className={btnSecondary}
                      data-hint-action="stand"
                      aria-keyshortcuts="s"
                      onClick={handleStand}
                      disabled={loading}
                    >
                      {t('button.stand')}
                    </button>
                  </div>
                </>
              )}

              {isCoupEnd && (
                <div className="flex gap-2">
                  <button
                    type="button"
                    className={btnPrimary}
                    aria-keyshortcuts="n"
                    onClick={() => execApi('nextcoup')}
                    disabled={loading}
                  >
                    {t('button.nextCoup')}
                  </button>
                  <button
                    type="button"
                    className={btnWarning}
                    aria-keyshortcuts="r"
                    onClick={() => execApi('retire')}
                    disabled={loading}
                  >
                    {t('button.retire')}
                  </button>
                </div>
              )}

              {!canDecide && !isCoupEnd && !gameOver && (
                <p className="text-ds-text-muted text-sm" data-testid="banque-wait-notice">
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
