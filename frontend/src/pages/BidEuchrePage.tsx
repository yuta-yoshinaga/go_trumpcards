import { useEffect, useMemo, useState } from 'react';
import { bidEuchreApi } from '../api/gameApi';
import { ActionLogSection } from '../components/ActionLogSection';
import { CardImage } from '../components/CardImage';
import { CliTerminal } from '../components/cli/CliTerminal';
import { CliToggle } from '../components/cli/CliToggle';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { useCliGame } from '../hooks/useCliGame';
import { useCliMode } from '../hooks/useCliMode';
import { useGameApi } from '../hooks/useGameApi';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSuccess, btnWarning } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import type { BidEuchreResponse } from '../types/card';
import { BidEuchrePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { BIDEUCHRE_HELP, parseBidEuchreCommand } from '../utils/cli/commands/bideuchreCommands';
import { formatBidEuchreState } from '../utils/cli/formatters/bideuchreFormatter';
import { hintLocalCommand } from '../utils/cli/hintText';
import type { CliGameConfig } from '../utils/cli/types';

/** Declarations, in menu order. **Two of them are no-trump forms.** */
const TRUMPS = [0, 1, 2, 3, 4, 5];

/** The two no-trump declarations, which `allowNoTrump` can switch off. */
const NO_TRUMPS = [4, 5];

/** Bid Euchre tutorial step definitions. */
const BIDEUCHRE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="bideuchre-info"]',
    messageKey: 'tutorial.info',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bideuchre-players"]',
    messageKey: 'tutorial.players',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bideuchre-hand"]',
    messageKey: 'tutorial.hand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bideuchre-actions"]',
    messageKey: 'tutorial.actions',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="bideuchre-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const BIDEUCHRE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [BidEuchrePhase.BID]: 'bid',
  [BidEuchrePhase.CHOOSE_TRUMP]: 'chooseTrump',
  [BidEuchrePhase.PLAY]: 'play',
  [BidEuchrePhase.HAND_END]: 'handEnd',
  [BidEuchrePhase.GAME_END]: 'gameEnd',
};

/** Renders the Bid Euchre game page: 24-card partnership euchre with an auction. */
export const BidEuchrePage = withTutorial(BidEuchrePageContent, 'bideuchre', BIDEUCHRE_TUTORIAL_STEPS);

/** Inner content of the Bid Euchre page, wrapped by TutorialProvider. */
function BidEuchrePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('bideuchre');
  const { state, loading, error, exec, retry } = useGameApi(bidEuchreApi.exec);

  const [bidValue, setBidValue] = useState(3);
  const [trumpChoice, setTrumpChoice] = useState(0);
  const [selected, setSelected] = useState<number | null>(null);

  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: run once on mount.
  useEffect(() => {
    exec('reset');
  }, []);

  // CLI mode
  const { cliEnabled, toggleCli, logEntries, addInput, addOutput, addError, clearLog } = useCliMode('bideuchre');

  // **フックは早期 return より上。**`if (!state)` の下に置くと、初回レンダー
  // だけフック数が変わってページが骨組みのまま固まる (#4561)。
  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('bideuchre', state);
  const cliConfig: CliGameConfig<BidEuchreResponse, Parameters<typeof bidEuchreApi.exec>> = useMemo(
    () => ({
      gameName: 'bideuchre',
      parseCommand: parseBidEuchreCommand,
      formatResponse: formatBidEuchreState,
      helpText: BIDEUCHRE_HELP,
      localCommand: hintLocalCommand(frontendHint),
    }),
    [frontendHint],
  );
  const { handleCommand } = useCliGame(exec, cliConfig, state, { addInput, addOutput, addError, clearLog });

  const { cardWidth } = useCardDimensions();
  const phaseNames = usePhaseNames('bideuchre', BIDEUCHRE_PHASE_KEYS);

  if (!state)
    return <GameSkeleton gameKey="bideuchre" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 6 }} />;

  const human = state.players.find((p) => p.isHuman);
  const isBid = state.phase === BidEuchrePhase.BID;
  const isChooseTrump = state.phase === BidEuchrePhase.CHOOSE_TRUMP;
  const isPlay = state.phase === BidEuchrePhase.PLAY;
  const isHandEnd = state.phase === BidEuchrePhase.HAND_END;
  const isGameEnd = state.phase === BidEuchrePhase.GAME_END || state.gameEndFlag;
  // 人間は席 0 = チーム 0。勝敗はチームで判定する。
  const humanWon = isGameEnd && state.winnerTeam === 0;
  const isHumanBid = isBid && state.bidPlayerIdx === 0 && !isGameEnd;
  // **切札を名乗るのは落札者。**手番ではない。
  const isHumanTrump = isChooseTrump && state.declarerIdx === 0 && !isGameEnd;
  const isHumanPlay = isPlay && state.currentPlayerIdx === 0 && !isGameEnd;

  const playerLabel = (id: number, isHuman: boolean): string => (isHuman ? t('you') : t('cpu', { id }));
  const trumpLabel = (trump: number): string => t(`trumpName.${trump}`);

  const canPlay = (i: number) => state.validPlays.includes(i);

  // **立っている宣言を上回るものだけ出す。ただし親は同額でも奪える。**
  // 上限だけ出していると、非ディーラーが必ず弾かれる値を選べてしまう。
  const humanIsDealer = state.dealerIdx === 0;
  const floor = state.highBid === null ? state.minBid : humanIsDealer ? state.highBid.value : state.highBid.value + 1;
  const bids: number[] = [];
  for (let v = Math.max(state.minBid, floor); v <= state.maxBid; v++) bids.push(v);

  // **ノートランプは設定で切れる。**サーバーが弾く選択肢は出さない。
  const trumpOptions = state.config?.allowNoTrump === false ? TRUMPS.filter((tr) => !NO_TRUMPS.includes(tr)) : TRUMPS;

  // **床が上がると 3 は選べなくなる。**選択値を出せる範囲へ寄せる。
  const selectedBid = bids.length > 0 && !bids.includes(bidValue) ? bids[0] : bidValue;

  const handleBid = () => {
    exec('bid', { value: selectedBid });
  };

  const handleTrump = () => {
    exec('trump', { trump: trumpChoice });
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
      title={tc('nav.bideuchre')}
      gameThemeBg={gameTheme.bideuchre.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanBid || isHumanTrump || isHumanPlay}
      gamePath="/bideuchre"
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
            <div className="text-ds-text-primary text-center mb-2" data-tutorial="bideuchre-info">
              <span className="mr-4">{t('hand', { n: state.handNumber })}</span>
              {state.highBid && (
                <span className="mr-4" data-testid="bideuchre-contract">
                  {t('contract')}: {t('contractTricks', { n: state.highBid.value })} / {t('trump')}:{' '}
                  {state.trumpChosen ? trumpLabel(state.trump) : t('trumpUndecided')}
                </span>
              )}
            </div>

            {/* No kitty — 24 / 4 = 6 leaves no remainder. */}
            <div className="mb-2 text-center text-ds-text-muted text-xs" data-testid="bideuchre-no-kitty">
              {t('noKittyNote')}
            </div>

            {/* Score sheet */}
            <div className="mb-2 p-2 rounded bg-black/20 text-sm" data-testid="bideuchre-scores">
              <div className="mb-1 text-ds-text-primary">{t('scoreTitle')}</div>
              <div>
                {t('team', { n: 0 })}: {state.scores[0]} / {t('team', { n: 1 })}: {state.scores[1]}
              </div>
              <div className="text-xs text-ds-text-muted">{t('gameTargetNote', { n: state.gameTarget })}</div>
            </div>

            {/* Players */}
            <div className="mb-2 p-2 rounded bg-black/30" data-tutorial="bideuchre-players">
              <div className="mb-1 text-ds-text-primary text-sm">{t('playersTitle')}</div>
              {state.players.map((p) => (
                <div
                  key={`player-${p.id}`}
                  data-testid="bideuchre-player"
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
              <div className="mb-2 flex items-center gap-2" data-testid="bideuchre-trick">
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
                data-testid="bideuchre-settlement"
                role="status"
                aria-live="polite"
              >
                <div className="mb-1">{t('settlementTitle')}</div>
                <div>{result.made ? t('madeLine', { bid: result.bid }) : t('setLine', { bid: result.bid })}</div>
                <div>{t('pointsLine', { t0: result.points[0], t1: result.points[1] })}</div>
                <div>{t('tricksLine', { t0: result.tricks[0], t1: result.tricks[1] })}</div>
                {/* **未達で失うのは宣言額。**取ったトリック数ではない。 */}
                {!result.made && (
                  <div className="text-xs" data-testid="bideuchre-set-note">
                    {t('setCostNote')}
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
          <GameFooter className={`${gameTheme.bideuchre.footer} px-4 py-2.5`}>
            <div className="mb-2" data-tutorial="bideuchre-hand">
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
              <div className="text-ds-text-muted text-xs mb-2" data-testid="bideuchre-bid-notice">
                <div>{t('bidNotice')}</div>
                {/* **親だけは同額で奪える。** */}
                <div data-testid="bideuchre-dealer-note">{t('dealerEqualNote')}</div>
              </div>
            )}
            {isHumanTrump && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="bideuchre-trump-notice">
                <div>{t('trumpNotice')}</div>
                {/* **NT ローは序列が逆転する。** */}
                <div data-testid="bideuchre-ntlow-note">{t('noTrumpLowNote')}</div>
              </div>
            )}
            {isHumanPlay && (
              <div className="text-ds-text-muted text-xs mb-2" data-testid="bideuchre-play-notice">
                {t('playNotice')}
              </div>
            )}

            <label className="flex items-center gap-1 text-ds-text-primary text-xs w-full justify-center cursor-pointer min-h-[44px]">
              <input
                type="checkbox"
                checked={frontendHintEnabled}
                onChange={(e) => setFrontendHintEnabled(e.target.checked)}
              />
              {tc('hint.toggle', { ns: 'tutorial' })}
            </label>
            <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

            <div className="flex flex-wrap gap-2 items-center" data-tutorial="bideuchre-actions">
              {isHumanBid && (
                <>
                  <label className="text-ds-text-muted text-xs flex items-center gap-1" htmlFor="bideuchre-bid-select">
                    {t('bidLabel')}
                    <select
                      id="bideuchre-bid-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={selectedBid}
                      onChange={(e) => setBidValue(Number(e.target.value))}
                    >
                      {bids.map((v) => (
                        <option key={v} value={v}>
                          {v}
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

              {isHumanTrump && (
                <>
                  <label
                    className="text-ds-text-muted text-xs flex items-center gap-1"
                    htmlFor="bideuchre-trump-select"
                  >
                    {t('trumpLabel')}
                    <select
                      id="bideuchre-trump-select"
                      className="bg-black/30 text-ds-text-primary rounded px-1 min-h-[44px]"
                      value={trumpChoice}
                      onChange={(e) => setTrumpChoice(Number(e.target.value))}
                    >
                      {trumpOptions.map((tr) => (
                        <option key={tr} value={tr}>
                          {trumpLabel(tr)}
                        </option>
                      ))}
                    </select>
                  </label>
                  <button type="button" className={btnPrimary} onClick={handleTrump} disabled={loading}>
                    {t('trumpButton')}
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
                dataTutorial="bideuchre-reset-button"
              />
            </div>
          </GameFooter>
        </>
      )}
    </GamePageShell>
  );
}
