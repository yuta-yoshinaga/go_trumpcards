import { useCallback, useMemo, useState } from 'react';
import { freebetApi } from '../api/gameApi';
import { ActionLogPanel } from '../components/ActionLogPanel';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ChipBetInput } from '../components/common/ChipBetInput';
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
import type { FreeBetResponse } from '../types/card';
import { FREE_BET_RESULT } from '../types/games/freebet';
import { FreeBetPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { FREEBET_CLI_HELP, parseFreeBetCommand } from '../utils/cli/commands/freebetCommands';
import { formatFreeBetState } from '../utils/cli/formatters/freebetFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const FB_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="fb-bet"]', messageKey: 'tutorial.bet', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="fb-hands"]', messageKey: 'tutorial.hands', placement: 'bottom', advanceOn: 'next' },
  { target: '[data-tutorial="fb-actions"]', messageKey: 'tutorial.free', placement: 'top', advanceOn: 'next' },
];

/** Renders the Free Bet Blackjack game page (#5262). */
export const FreeBetPage = withTutorial(FreeBetPageContent, 'freebet', FB_TUTORIAL_STEPS);

function FreeBetPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('freebet');

  const [ante, setAnte] = useState(50);
  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(freebetApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('freebet');
  const cliConfig: CliGameConfig<FreeBetResponse, Parameters<typeof freebetApi.exec>> = useMemo(
    () => ({
      gameName: 'freebet',
      parseCommand: parseFreeBetCommand,
      formatResponse: formatFreeBetState,
      helpText: FREEBET_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBetPhase = phase === FreeBetPhase.BET;
  const isPlayPhase = phase === FreeBetPhase.PLAY;
  const isResultPhase = phase === FreeBetPhase.RESULT;
  const gameOver = !!state?.gameEndFlag;

  const handleDeal = useCallback(() => execApi('bet', { ante }), [execApi, ante]);

  const actionBindings = useMemo(
    () => [
      { key: 'h', action: () => execApi('hit'), enabled: isPlayPhase },
      { key: 's', action: () => execApi('stand'), enabled: isPlayPhase },
      { key: 'n', action: () => execApi('next'), enabled: isResultPhase && !gameOver },
    ],
    [execApi, isPlayPhase, isResultPhase, gameOver],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** `if (!state)` の下に置くと初回レンダーだけ
  // フック数が変わってページが骨組みのまま固まります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('freebet', state);

  if (!state) return <GameSkeleton gameKey="freebet" layout={{ kind: 'casino-table', sections: [1, 1] }} />;

  const phaseName =
    {
      [FreeBetPhase.BET]: t('phase.bet'),
      [FreeBetPhase.PLAY]: t('phase.play'),
      [FreeBetPhase.RESULT]: t('phase.result'),
    }[state.phase] ?? '';

  // **自分が出した金だけを数える。** ハウスの出資は失う対象ではないので、
  // ここに `freeBet` を足すと無料ダブルで負けたラウンドが倍の損に見える。
  const staked = state.hands.reduce((sum, h) => sum + h.bet, 0);
  const net = state.payout - staked;
  const won = state.hands.some((h) => h.result === FREE_BET_RESULT.win || h.result === FREE_BET_RESULT.blackjack);
  const resultKeyOf = (r: number) => Object.entries(FREE_BET_RESULT).find(([, v]) => v === r)?.[0] ?? 'none';

  return (
    <GamePageShell
      title={tc('nav.freebet')}
      gameThemeBg={gameTheme.freebet.bg}
      phaseName={phaseName}
      gamePath="/freebet"
      gameEndFlag={gameOver}
      winShow={isResultPhase && won}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-testid="fb-chips">
            {t('label.chips')}: {state.chips}
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

            <div className="text-ds-text-primary text-center text-sm mb-2" data-testid="fb-bet-line">
              {t('label.round')}: {state.roundNumber}
              {state.anteBet > 0 && (
                <>
                  {' · '}
                  {t('label.ante')}: {state.anteBet}
                </>
              )}
            </div>

            {state.dealerCards.length > 0 && (
              <div className="mb-3">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.dealer')}</div>
                <div className="flex justify-center gap-1 flex-wrap" data-testid="fb-dealer-cards">
                  {state.dealerCards.map((card, i) => (
                    <AnimatedCard key={`dealer-${card.design}-${card.value}-${i}`} card={card} width={cardWidth} />
                  ))}
                </div>
                <div className="text-ds-text-primary text-center text-sm mt-1" data-testid="fb-dealer-score">
                  {state.dealerScore} {t('label.score')}
                </div>
              </div>
            )}

            {/* **22 は名指しする。** 無料ダブル / 無料スプリットの対価がこれ。 */}
            {state.dealerPushed22 && (
              <p className="text-ds-text-muted text-center text-xs mb-2" data-testid="fb-dealer22">
                {t('dealer22Notice')}
              </p>
            )}

            {state.hands.length > 0 && (
              <div data-tutorial="fb-hands">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.yourHands')}</div>
                {state.hands.map((h, i) => (
                  <div
                    key={`hand-${i}-${h.cards.length}-${h.score}`}
                    data-testid={`fb-hand-${i}`}
                    className={`mb-2 rounded px-2 py-1 ${
                      isPlayPhase && i === state.activeHand ? 'ring-2 ring-ds-success' : ''
                    }`}
                  >
                    <div className="flex justify-center gap-1 flex-wrap">
                      {h.cards.map((card, k) => (
                        <AnimatedCard key={`h${i}-${card.design}-${card.value}-${k}`} card={card} width={cardWidth} />
                      ))}
                    </div>
                    <div className="text-ds-text-primary text-center text-sm mt-1">
                      {state.hands.length > 1 && `${t('label.hand', { idx: i + 1 })} · `}
                      {h.score} {t('label.score')} · {t('label.bet')} {h.bet}
                      {/* **ハウス持ちは別建てで出す。** 足すと「いくら失うのか」が消える。 */}
                      {h.freeBet > 0 && (
                        <span className="text-ds-text-muted" data-testid={`fb-free-${i}`}>
                          {' '}
                          + {t('label.freeBet')} {h.freeBet}
                        </span>
                      )}
                      {isResultPhase && ` · ${t(`result.${resultKeyOf(h.result)}`)}`}
                    </div>
                  </div>
                ))}
              </div>
            )}

            {isResultPhase && (
              <div className="text-center mb-2" data-testid="fb-result">
                <div className={`text-sm font-medium ${net >= 0 ? 'text-ds-success' : 'text-ds-error'}`}>
                  {t('label.net')}: {net}
                </div>
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.freebet.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-col items-center gap-2 pb-2">
              {isBetPhase && !gameOver && (
                <div className="flex flex-col items-center gap-2" data-tutorial="fb-bet">
                  <p className="text-ds-text-muted text-sm">{t('betGuide')}</p>
                  <ChipBetInput
                    id="freebet-ante"
                    label={t('label.ante')}
                    value={ante}
                    onChange={setAnte}
                    max={state.chips}
                  />
                  <button type="button" className={btnPrimary} onClick={handleDeal} disabled={loading}>
                    {t('button.deal')}
                  </button>
                </div>
              )}

              {isPlayPhase && (
                <div className="flex flex-col items-center gap-2" data-tutorial="fb-actions">
                  <p className="text-ds-text-muted text-sm">{t('playGuide')}</p>
                  <p className="text-ds-text-muted text-xs" data-testid="fb-free-notice">
                    {t('freeNotice')}
                  </p>
                  <div className="flex gap-2 flex-wrap justify-center">
                    <button
                      type="button"
                      className={btnSuccess}
                      data-hint-action="hit"
                      onClick={() => execApi('hit')}
                      disabled={loading}
                    >
                      {t('button.hit')}
                    </button>
                    <button
                      type="button"
                      className={btnSecondary}
                      data-hint-action="stand"
                      onClick={() => execApi('stand')}
                      disabled={loading}
                    >
                      {t('button.stand')}
                    </button>
                    {/* **押せるかどうかはサーバが決める。** 手札から計算し直さない。 */}
                    {state.canFreeDouble && (
                      <button
                        type="button"
                        className={btnWarning}
                        data-testid="fb-freedouble"
                        data-hint-action="freedouble"
                        onClick={() => execApi('freedouble')}
                        disabled={loading}
                      >
                        {t('button.freeDouble')}
                      </button>
                    )}
                    {state.canFreeSplit && (
                      <button
                        type="button"
                        className={btnWarning}
                        data-testid="fb-freesplit"
                        data-hint-action="freesplit"
                        onClick={() => execApi('freesplit')}
                        disabled={loading}
                      >
                        {t('button.freeSplit')}
                      </button>
                    )}
                  </div>
                </div>
              )}

              {isResultPhase && !gameOver && (
                <button type="button" className={btnPrimary} onClick={() => execApi('next')} disabled={loading}>
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
