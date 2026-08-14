import { useCallback, useMemo } from 'react';
import { colourwhistApi } from '../api/gameApi';
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
import { KbdBadge } from '../components/KbdBadge';
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
import { btnPrimary, btnSecondary } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { ColourWhistResponse } from '../types/card';
import { COLOUR_WHIST_BIDDABLE, COLOUR_WHIST_NO_TRUMP } from '../types/games/colourwhist';
import { ColourWhistContract, ColourWhistPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { COLOURWHIST_CLI_HELP, parseColourWhistCommand } from '../utils/cli/commands/colourwhistCommands';
import { formatColourWhistState } from '../utils/cli/formatters/colourwhistFormatter';
import type { CliGameConfig } from '../utils/cli/types';
import { hintCheckboxItem } from '../utils/settingsItems';

const CW_TUTORIAL_STEPS: TutorialStep[] = [
  { target: '[data-tutorial="cw-bid"]', messageKey: 'tutorial.bid', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cw-hand"]', messageKey: 'tutorial.hand', placement: 'top', advanceOn: 'next' },
  { target: '[data-tutorial="cw-score"]', messageKey: 'tutorial.score', placement: 'bottom', advanceOn: 'next' },
];

/** The four trump choices. Suits are 1..4. */
const TRUMP_CHOICES = [
  { suit: 1, key: 'trump.spade' },
  { suit: 2, key: 'trump.clover' },
  { suit: 3, key: 'trump.heart' },
  { suit: 4, key: 'trump.diamond' },
] as const;

/** Maps a contract value to its i18n key, including the un-biddable Troel. */
const CONTRACT_KEYS: Record<number, string> = {
  [ColourWhistContract.NONE]: 'contract.none',
  [ColourWhistContract.SAMEN]: 'contract.samen',
  [ColourWhistContract.ALLEEN]: 'contract.alleen',
  [ColourWhistContract.MISERIE]: 'contract.miserie',
  [ColourWhistContract.TROEL]: 'contract.troel',
};

/** Renders the Colour Whist game page (#5231). */
export const ColourWhistPage = withTutorial(ColourWhistPageContent, 'colourwhist', CW_TUTORIAL_STEPS);

function ColourWhistPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('colourwhist');

  const { cardWidth } = useCardDimensions();
  const { state, loading, error, exec: execApi, retry } = useGameApi(colourwhistApi.exec);

  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('colourwhist');
  const cliConfig: CliGameConfig<ColourWhistResponse, Parameters<typeof colourwhistApi.exec>> = useMemo(
    () => ({
      gameName: 'colourwhist',
      parseCommand: parseColourWhistCommand,
      formatResponse: formatColourWhistState,
      helpText: COLOURWHIST_CLI_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(execApi, cliConfig, state, { addInput, addOutput, addError, clearLog });

  useMountReset(execApi);

  const phase = state?.phase;
  const isBidPhase = phase === ColourWhistPhase.BID;
  const isCallPhase = phase === ColourWhistPhase.CALL;
  const isPlayPhase = phase === ColourWhistPhase.PLAY;
  const isRoundEnd = phase === ColourWhistPhase.ROUND_END;

  // **パス (0) も契約値。** 値を必ず送ります。
  const handleBid = useCallback((contract: number) => execApi('bid', undefined, contract), [execApi]);
  const handleCall = useCallback((suit: number) => execApi('call', undefined, undefined, suit), [execApi]);
  const handlePlay = useCallback((cardIndex: number) => execApi('play', cardIndex), [execApi]);

  const actionBindings = useMemo(
    () => [
      { key: 'n', action: () => execApi('next'), enabled: isRoundEnd },
      { key: 'g', action: () => execApi('giveup'), enabled: !!state && !state.gameEndFlag },
    ],
    [execApi, isRoundEnd, state],
  );
  useActionKeyboardNav({ bindings: actionBindings, enabled: !!state && !loading });

  // **フックは早期 return より上。** 下に置くと初回レンダーだけフック数が変わります (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('colourwhist', state);

  if (!state) {
    return (
      <GameSkeleton gameKey="colourwhist" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />
    );
  }

  const human = state.players.find((p) => p.isHuman);
  const legal = new Set(state.validPlays);
  const phaseName = t(
    isBidPhase
      ? 'phase.bid'
      : isCallPhase
        ? 'phase.call'
        : isPlayPhase
          ? 'phase.play'
          : isRoundEnd
            ? 'phase.roundEnd'
            : 'phase.gameEnd',
  );
  const contractLabel = t(CONTRACT_KEYS[state.contract] ?? 'contract.none');
  const trumpLabel =
    state.trumpSuit === COLOUR_WHIST_NO_TRUMP
      ? t('trump.none')
      : t(TRUMP_CHOICES.find((c) => c.suit === state.trumpSuit)?.key ?? 'trump.none');

  return (
    <GamePageShell
      title={tc('nav.colourwhist')}
      gameThemeBg={gameTheme.colourwhist.bg}
      phaseName={phaseName}
      gamePath="/colourwhist"
      gameEndFlag={state.gameEndFlag}
      winShow={state.gameEndFlag && state.winnerIdx === 0}
      lossShow={state.gameEndFlag && state.winnerIdx !== 0}
      loading={loading}
      isHumanTurn={state.isHumanTurn}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={
        <>
          <span data-tutorial="cw-score">
            {t('label.round')}: {state.roundNumber}
            {state.config ? ` / ${state.config.rounds}` : ''}
          </span>
          <CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />
        </>
      }
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div data-testid="card-area" className={`overflow-y-auto pt-3 px-4 lg:px-8 flex-1 ${lgCardAreaConstraint}`}>
            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            {/* **競りが飛ばされた理由を出す。** 出さないと不具合に見えます。 */}
            {state.troelForced && (
              <p
                className="text-ds-warning text-center text-sm font-medium mb-2"
                data-testid="colourwhist-troel-notice"
              >
                {t('troelNotice')}
              </p>
            )}

            <div className="text-ds-text-primary text-center text-sm mb-2 space-y-1">
              <div data-testid="colourwhist-contract">
                {t('label.contract')}: {contractLabel}
                {state.contract !== ColourWhistContract.NONE && ` / ${t('label.trump')}: ${trumpLabel}`}
              </div>
              {state.declarerIdx >= 0 && (
                <div data-testid="colourwhist-declarer">
                  {t('label.declarer')}: #{state.declarerIdx} ({t('label.declarerTricks')} {state.declarerTricks})
                </div>
              )}
              {state.contract !== ColourWhistContract.NONE && (
                <div data-testid="colourwhist-partner">
                  {t('label.partner')}: {state.partnerIdx >= 0 ? `#${state.partnerIdx}` : t('label.hidden')}
                </div>
              )}
            </div>

            {state.currentTrick.length > 0 && (
              <div className="mb-4" data-testid="colourwhist-trick">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.trick')}</div>
                <div className="flex justify-center gap-3 flex-wrap">
                  {state.currentTrick.map((tc) => (
                    <div key={`trick-${tc.playerIdx}`} className="flex flex-col items-center">
                      <span className="text-ds-text-muted text-xs mb-1">#{tc.playerIdx}</span>
                      <AnimatedCard card={tc.card} width={cardWidth} />
                    </div>
                  ))}
                </div>
              </div>
            )}

            <div className="flex justify-center gap-4 flex-wrap mb-4" data-testid="colourwhist-seats">
              {state.players.map((p) => (
                <div key={`seat-${p.id}`} className="text-center text-xs text-ds-text-muted">
                  <div>
                    #{p.id} {p.isHuman ? t('label.you') : ''}
                    {p.isDeclarerSide && ` (${t('label.declarerSide')})`}
                    {p.hasPassed && ` (${t('label.passed')})`}
                  </div>
                  {/* **得点は負にもなります。** ゼロサムなので当然そうなります。 */}
                  <div>
                    {t('label.score')}: {p.score} / {p.cardCount} {t('label.cards')}
                  </div>
                </div>
              ))}
            </div>

            {human && human.cards.length > 0 && (
              <div data-tutorial="cw-hand">
                <div className="text-ds-text-primary text-center text-sm font-bold mb-1">{t('label.hand')}</div>
                <div className="flex justify-center gap-1 flex-wrap" data-testid="colourwhist-hand">
                  {human.cards.map((card, i) => {
                    const playable = isPlayPhase && state.isHumanTurn && legal.has(i);
                    return (
                      <button
                        key={`hand-${card.design}-${card.value}-${i}`}
                        type="button"
                        onClick={() => handlePlay(i)}
                        disabled={loading || !playable}
                        aria-disabled={!playable}
                        aria-label={`${card.design} ${card.value}`}
                        className={playable ? 'ring-2 ring-ds-success rounded' : 'opacity-60'}
                      >
                        <AnimatedCard card={card} width={cardWidth} />
                      </button>
                    );
                  })}
                </div>
                {isPlayPhase && <p className="text-ds-text-muted text-center text-xs mt-1">{t('playGuide')}</p>}
              </div>
            )}

            {actionLog && <ActionLogPanel entries={actionLog} onClose={hideActionLog} />}
          </div>

          <GameFooter className={`${gameTheme.colourwhist.footer} px-4 pt-3`}>
            <ErrorAlert message={error} onRetry={retry} />
            <SettingsPanel
              title={tc('settings.title')}
              groups={[{ items: [hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled)] }]}
            />
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            {isBidPhase && state.isHumanTurn && (
              <div className="flex flex-col items-center gap-2 pb-2" data-tutorial="cw-bid">
                <p className="text-ds-text-muted text-sm">{t('bidGuide')}</p>
                <div className="flex justify-center gap-2 flex-wrap">
                  {/* **トルールのボタンは無い。** 配りでしか成立しない契約です。 */}
                  {COLOUR_WHIST_BIDDABLE.map((c) => (
                    <button
                      key={`bid-${c.contract}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleBid(c.contract)}
                      disabled={loading || c.contract <= state.contract}
                      aria-disabled={c.contract <= state.contract}
                    >
                      {t('button.bid', { contract: t(c.key) })}
                    </button>
                  ))}
                  <button
                    type="button"
                    className={btnSecondary}
                    onClick={() => handleBid(ColourWhistContract.NONE)}
                    disabled={loading}
                  >
                    {t('button.pass')}
                  </button>
                </div>
              </div>
            )}

            {isCallPhase && state.isHumanTurn && (
              <div className="flex flex-col items-center gap-2 pb-2">
                <p className="text-ds-text-muted text-sm">{t('callGuide')}</p>
                <div className="flex justify-center gap-2 flex-wrap">
                  {TRUMP_CHOICES.map((c) => (
                    <button
                      key={`call-${c.suit}`}
                      type="button"
                      className={btnPrimary}
                      onClick={() => handleCall(c.suit)}
                      disabled={loading}
                    >
                      {t('button.call', { trump: t(c.key) })}
                    </button>
                  ))}
                </div>
              </div>
            )}

            <div className="flex justify-center gap-2 pb-2 flex-wrap">
              {isRoundEnd && !state.gameEndFlag && (
                <button
                  type="button"
                  className={btnPrimary}
                  onClick={() => execApi('next')}
                  disabled={loading}
                  data-testid="cw-next-button"
                  aria-keyshortcuts="n"
                >
                  {t('button.next')}
                  <KbdBadge label={t('kbd.next')} />
                </button>
              )}
              <GameResetButton
                isGameEnd={state.gameEndFlag}
                onReset={() => execApi('reset')}
                requestConfirm={requestConfirm}
                loading={loading}
              />
              {!state.gameEndFlag && (
                <button
                  type="button"
                  className={btnSecondary}
                  onClick={() => execApi('giveup')}
                  disabled={loading}
                  aria-keyshortcuts="g"
                >
                  {t('button.giveUp')}
                  <KbdBadge label={t('kbd.giveUp')} />
                </button>
              )}
              <button type="button" className={btnSecondary} onClick={showActionLog} disabled={loading}>
                {tc('actionLog.view')}
              </button>
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
