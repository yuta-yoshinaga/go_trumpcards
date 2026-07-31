import { useEffect, useMemo, useState } from 'react';
import { vintApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { VintResponse } from '../types/card';
import { VintPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseVintCommand, VINT_HELP } from '../utils/cli/commands/vintCommands';
import { formatVintState } from '../utils/cli/formatters/vintFormatter';
import type { CliGameConfig } from '../utils/cli/types';

/** Denominations in bidding order — spades LOWEST, no trump highest. */
const DENOMS = [0, 1, 2, 3, 4];

/** Vint tutorial step definitions. */
const VINT_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="vint-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vint-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vint-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vint-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="vint-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const VINT_PHASE_KEYS: Readonly<Record<number, string>> = {
  [VintPhase.BID]: 'bid',
  [VintPhase.PLAY]: 'play',
  [VintPhase.HAND_END]: 'handEnd',
  [VintPhase.GAME_END]: 'gameEnd',
};

/** Renders the Vint game page: Russian bridge, played without a dummy. */
export const VintPage = withTutorial(VintPageContent, 'vint', VINT_TUTORIAL_STEPS);

/** Inner content of the Vint page, wrapped by TutorialProvider. */
function VintPageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('vint');
  const { state, loading, error, exec, retry } = useGameApi(vintApi.exec);

  const [bidLevel, setBidLevel] = useState(1);
  const [bidDenom, setBidDenom] = useState(0);
  const [selected, setSelected] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('vint');
  const cliConfig: CliGameConfig<VintResponse, Parameters<typeof vintApi.exec>> = useMemo(
    () => ({
      gameName: 'vint',
      parseCommand: parseVintCommand,
      formatResponse: formatVintState,
      helpText: VINT_HELP,
    }),
    [],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('vint', VINT_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="vint" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 13 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === VintPhase.BID;
  const isPlay = state.phase === VintPhase.PLAY;
  const isHandEnd = state.phase === VintPhase.HAND_END;
  const isGameEnd = state.phase === VintPhase.GAME_END || state.gameEndFlag;
  // 人間は席 0 = チーム 0。勝敗はチームで判定する。
  const humanWon = isGameEnd && state.winnerTeam === 0;
  const isHumanBid = isBid && state.bidPlayerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const denomLabel = (denom: number): string => t(`denom.${denom}`);

  const canPlay = (i: number) => state.validPlays.includes(i);

  const levels: number[] = [];
  for (let l = state.minLevel; l <= state.maxLevel; l++) levels.push(l);

  const handleBid = () => {
    exec('bid', { level: bidLevel, denom: bidDenom });
  };

  const handlePlay = () => {
    if (selected === null) return;
    exec('play', { cardIndex: selected });
    setSelected(null);
  };

  const handleManualReset = () => {
    hideActionLog();
    setSelected(null);
    exec('reset');
  };

  const result = state.lastResult;

  return (
    <GamePageShell
      title={tc('nav.vint')}
      gameThemeBg={gameTheme.vint.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBid || isHumanPlay}
      gamePath="/vint"
      gameEndFlag={isGameEnd}
      winShow={humanWon}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
      headerExtra={<CliToggle cliEnabled={cliEnabled} onToggle={toggleCli} />}
    >
      {cliEnabled ? (
        <CliTerminal logEntries={logEntries} onCommand={handleCommand} disabled={loading} />
      ) : (
        <>
          <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="vint-info">
              <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
              {state.highBid && (
                <span className="mr-4" data-testid="vint-contract">
                  {t('contract')}: {state.highBid.level} {denomLabel(state.highBid.denom)} (
                  {t('trickValue', { n: state.highBid.trickValue })})
                </span>
              )}
            </div>

            {/* No dummy — the rule that separates this from bridge. */}
            <div className="mb-2 text-center text-ds-text-muted text-xs" data-testid="vint-no-dummy">
              {t('noDummyNote')}
            </div>

            {/* The bidding order — reversed from bridge. */}
            <div className="mb-2 p-2 rounded bg-black/20 text-xs" data-testid="vint-ladder">
              <div className="mb-1 text-ds-text-primary">{t('ladderTitle')}</div>
              <div className="flex flex-wrap gap-x-3">
                {DENOMS.map((d) => (
                  <span key={d} className={d === 0 ? 'text-ds-text-muted' : 'text-ds-text-primary'}>
                    {denomLabel(d)} ({state.trickValues[d]})
                  </span>
                ))}
              </div>
              <div className="mt-1 text-ds-text-muted">{t('ladderNote')}</div>
            </div>

            {/* Score sheet */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="vint-scores">
              <div className="mb-1 text-ds-text-primary">{t('scoreTitle')}</div>
              <div>
                {t('team', { n: 0 })}: {t('below')} {state.below[0]} / {t('above')} {state.above[0]} / {t('games')}{' '}
                {state.gamesWon[0]}
              </div>
              <div>
                {t('team', { n: 1 })}: {t('below')} {state.below[1]} / {t('above')} {state.above[1]} / {t('games')}{' '}
                {state.gamesWon[1]}
              </div>
              <div className="text-xs text-ds-text-muted">{t('gameTargetNote', { n: state.gameTarget })}</div>
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="vint-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="vint-player"
                  className={`text-sm py-0.5 flex items-center gap-2 ${
                    p.isCurrentTurn && !isGameEnd ? 'text-ds-warning' : 'text-ds-text-muted'
                  } ${p.isHuman ? 'font-semibold' : ''}`}
                >
                  <span>{playerLabel(p.id, p.isHuman)}</span>
                  <span>({t('team', { n: p.team })})</span>
                  {p.isDealer && <span className="text-ds-accent">[{t('dealer')}]</span>}
                  {p.isDeclarer && <span className="text-ds-success">[{t('declarer')}]</span>}
                  <span>{t('tricksWon', { n: p.tricksWon })}</span>
                  {!p.isHuman && p.cards.length === 0 && <span>{t('hiddenHand', { count: p.cardCount })}</span>}
                </div>
              ))}
            </div>

            {/* Trick */}
            {state.trick.length > 0 && (
              <div className="mb-2 flex items-center gap-2" data-testid="vint-trick">
                <span className="text-ds-text-muted text-sm">{t('trick')}</span>
                {state.trick.map((c, i) => (
                  <CardImage key={`trick-${c.design}-${c.value}-${i}`} card={c} width={cardWidth} />
                ))}
              </div>
            )}

            {/* Settlement */}
            {(isHandEnd || isGameEnd) && result && (
              <div
                className={`mb-2 p-2 rounded text-sm ${badgeWarningColors}`}
                data-testid="vint-settlement"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('settlementTitle')}</div>
                <div>
                  {result.made
                    ? t('madeLine', { tricks: result.declarerTricks })
                    : t('setLine', { tricks: result.declarerTricks })}
                </div>
                <div>{t('trickPointsLine', { t0: result.trickPoints[0], t1: result.trickPoints[1] })}</div>
                {/* **両チームとも得点する。**ブリッジとの決定的な違い。 */}
                <div className="text-xs">{t('bothScoreNote')}</div>
                <div>{t('honourLine', { t0: result.honourPoints[0], t1: result.honourPoints[1] })}</div>
                <div>{t('aceLine', { t0: result.acePoints[0], t1: result.acePoints[1] })}</div>
                {(result.penalty[0] > 0 || result.penalty[1] > 0) && (
                  <div data-testid="vint-penalty">
                    {t('penaltyLine', { t0: result.penalty[0], t1: result.penalty[1] })}
                  </div>
                )}
              </div>
            )}

            <GameMessageBox
              message={state.message}
              messageCode={state.messageCode}
              messageParams={state.messageParams}
            />

            <ActionLogSection
              isEndPhase={isGameEnd}
              actionLog={actionLog}
              showActionLog={showActionLog}
              hideActionLog={hideActionLog}
            />
          </div>

          {/* Footer */}
          <GameFooter className={`${gameTheme.vint.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="vint-hand">
              <div className="text-ds-text-muted text-xs mb-1">{t('yourHand')}</div>
              <div className="flex flex-wrap gap-1">
                {(human?.cards ?? []).map((c, i) => (
                  <button
                    key={`hand-${c.design}-${c.value}-${i}`}
                    type="button"
                    data-hint-action="play"
                    onClick={() => setSelected(i)}
                    disabled={loading || (isPlay && !canPlay(i))}
                    className={`rounded ${selected === i ? 'ring-2 ring-ds-accent' : ''} ${
                      isPlay && !canPlay(i) ? 'opacity-40' : ''
                    }`}
                  >
                    <CardImage card={c} width={cardWidth} />
                  </button>
                ))}
              </div>
            </div>

            <ErrorAlert message={error} onRetry={retry} />

            {isHumanBid && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="vint-bid-notice">
                {t('bidNotice')}
              </div>
            )}
            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="vint-play-notice">
                {t('playNotice')}
              </div>
            )}

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="vint-actions">
              {isHumanBid && (
                <>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="vint-level-select">
                    {t('levelLabel')}
                    <select
                      id="vint-level-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={bidLevel}
                      onChange={(e) => setBidLevel(Number(e.target.value))}
                    >
                      {levels.map((l) => (
                        <option key={l} value={l}>
                          {l}
                        </option>
                      ))}
                    </select>
                  </label>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="vint-denom-select">
                    {t('denomLabel')}
                    <select
                      id="vint-denom-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={bidDenom}
                      onChange={(e) => setBidDenom(Number(e.target.value))}
                    >
                      {DENOMS.map((d) => (
                        <option key={d} value={d}>
                          {denomLabel(d)}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button type="button" className={btnPrimary} onClick={handleBid} disabled={loading}>
                    {t('bidButton')}
                  </button>
                  <button type="button" className={btnWarning} onClick={() => exec('pass')} disabled={loading}>
                    {t('passButton')}
                  </button>
                </>
              )}

              {isHumanPlay && (
                <button
                  type="button"
                  className={btnSuccess}
                  onClick={handlePlay}
                  disabled={loading || selected === null}
                >
                  {t('playButton')}
                </button>
              )}

              {isHandEnd && !isGameEnd && (
                <button type="button" className={btnSuccess} onClick={() => exec('next')} disabled={loading}>
                  {t('nextHand')}
                </button>
              )}

              {isGameEnd && (
                <span className="text-ds-text-primary text-sm font-semibold mr-1">
                  {humanWon ? t('win') : t('lose')}
                </span>
              )}

              <GameResetButton
                isGameEnd={isGameEnd}
                onReset={handleManualReset}
                requestConfirm={requestConfirm}
                loading={loading}
                dataTutorial="vint-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
