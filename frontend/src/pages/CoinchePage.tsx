import { useCallback, useEffect, useRef, useState } from 'react';
import { ActionLogSection } from '../components/ActionLogSection';
import { SettingsPanel } from '../components/common/SettingsPanel';
import { ErrorAlert } from '../components/ErrorAlert';
import { GameFooter } from '../components/GameFooter';
import { GameMessageBox } from '../components/GameMessageBox';
import { GamePageShell } from '../components/GamePageShell';
import { GameResetButton } from '../components/GameResetButton';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { RoundScoreAnnouncement } from '../components/RoundScoreAnnouncement';
import { GameSkeleton } from '../components/skeleton/GameSkeleton';
import { TrickDisplay } from '../components/TrickDisplay';
import { withTutorial } from '../components/tutorial/withTutorial';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { CPU_DIFFICULTY_OPTIONS, TARGET_SCORE_OPTIONS, useCoincheGame } from '../hooks/useCoincheGame';
import { useGameHint } from '../hooks/useGameHint';
import { useGamePageSetup } from '../hooks/useGamePageSetup';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { useSound } from '../providers/SoundProvider';
import { btnPrimary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint } from '../styles/gameStyles';
import { gameTheme } from '../styles/gameTheme';
import { COINCHE_CAPOT_POINTS, CoinchePhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { coincheLegalPlayIndices } from '../utils/coincheLegal';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Coinche tutorial step definitions. */
const COINCHE_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="be-bid-controls"]',
    messageKey: 'tutorial.bidControls',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-play-button"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-score-table"]',
    messageKey: 'tutorial.scoreTable',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="be-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const COINCHE_PHASE_KEYS: Readonly<Record<number, string>> = {
  [CoinchePhase.BID]: 'bid',
  [CoinchePhase.DOUBLE]: 'double',
  [CoinchePhase.PLAY]: 'play',
  [CoinchePhase.TRICK_END]: 'trickEnd',
  [CoinchePhase.ROUND_END]: 'roundEnd',
  [CoinchePhase.GAME_END]: 'gameEnd',
};

const SUIT_LABEL_KEYS: Readonly<Record<number, string>> = {
  1: 'suitSpade',
  2: 'suitClover',
  3: 'suitHeart',
  4: 'suitDiamond',
};

/** Renders the Coinche game page (4-player partnership trick-taking, 32-card deck). */
export const CoinchePage = withTutorial(CoinchePageContent, 'coinche', COINCHE_TUTORIAL_STEPS);

function CoinchePageContent() {
  const { t, tc, actionLog, showActionLog, hideActionLog, confirmOpen, requestConfirm, confirmReset, cancelReset } =
    useGamePageSetup('coinche');
  const {
    state,
    loading,
    error,
    exec: dispatch,
    retry,
    coincheConfig,
    selectedCardIndices,
    toggleCard,
    handleConfigChange,
    handleToggle,
    handlePlay,
    handleNextTrick,
    handleNextRound,
    handleBid,
    handleCoinche,
    handleSurcoinche,
    handleDeclineDouble,
    handlePass,
    hint,
    hintError,
    hintLoading,
    handleHint,
  } = useCoincheGame();

  // 目標点は先に選んでからスートを押す。契約は「点 + 切り札」の対なので、
  // 片方だけで送ると残りに既定値が入って別の契約になる。
  const [selectedPoints, setSelectedPoints] = useState<number | null>(null);

  const {
    hint: frontendHint,
    hintEnabled: frontendHintEnabled,
    setHintEnabled: setFrontendHintEnabled,
  } = useGameHint('coinche', state);
  const { cardWidth, isMobile } = useCardDimensions();

  const phaseNames = usePhaseNames('coinche', COINCHE_PHASE_KEYS);

  const { playSound } = useSound();
  const coincheTotal = state ? state.roundBeloteBonus[0] + state.roundBeloteBonus[1] : 0;
  const prevCoincheTotalRef = useRef<number | null>(null);
  const [coincheJustConfirmed, setCoincheJustConfirmed] = useState(false);
  useEffect(() => {
    if (!state) return;
    if (prevCoincheTotalRef.current === null) {
      prevCoincheTotalRef.current = coincheTotal;
      return;
    }
    if (coincheTotal > prevCoincheTotalRef.current) {
      setCoincheJustConfirmed(true);
      playSound('winFanfare');
    }
    prevCoincheTotalRef.current = coincheTotal;
  }, [coincheTotal, playSound, state]);

  // Clear timer keyed only on the flag so an unrelated state update mid-window can't cancel it.
  useEffect(() => {
    if (!coincheJustConfirmed) return;
    const id = setTimeout(() => setCoincheJustConfirmed(false), 2500);
    return () => clearTimeout(id);
  }, [coincheJustConfirmed]);

  const handleManualReset = useCallback(() => {
    hideActionLog();
    void (
      dispatch as unknown as (command: string, ci?: number, s?: number, cfg?: typeof coincheConfig) => Promise<void>
    )('reset', undefined, undefined, coincheConfig);
  }, [dispatch, hideActionLog, coincheConfig]);

  if (!state)
    return <GameSkeleton gameKey="coinche" layout={{ kind: 'trick-taking', trickArea: true, footerHandSize: 8 }} />;

  const humanPlayer = state.players.find((p) => p.isHuman);
  const isBidPhase = state.phase === CoinchePhase.BID;
  const isDoublePhase = state.phase === CoinchePhase.DOUBLE;
  const isPlayPhase = state.phase === CoinchePhase.PLAY;
  const isTrickEnd = state.phase === CoinchePhase.TRICK_END;
  const isRoundEnd = state.phase === CoinchePhase.ROUND_END;
  const isGameEnd = state.phase === CoinchePhase.GAME_END || state.gameEndFlag;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanBidTurn = isBidPhase && state.bidPlayerIdx === humanIdx;
  const isHumanDoubleTurn = isDoublePhase && state.currentPlayerIdx === humanIdx;
  // **上回れる契約だけをボタンにする。** 打てば必ず拒否される値を並べると、
  // 押せるのに何も起きない操作面ができる。
  const biddablePoints = state.biddablePoints ?? [];
  // 倍化できる側は立場で決まる: 守備側はコワンシュ、コワンシュされた
  // 宣言側はシュルコワンシュ。
  const onMakerTeam = humanPlayer ? humanPlayer.team === state.makerTeam : false;
  const canCoinche = isHumanDoubleTurn && !onMakerTeam && state.double === 0;
  const canSurcoinche = isHumanDoubleTurn && onMakerTeam && state.double === 1;
  const isHumanTurn = isPlayPhase && state.players[state.currentPlayerIdx]?.isHuman === true;
  // Legal-play highlight: compute the follow-suit / trump-obligation legal set
  // on the human's play turn only (mirrors internal/domain/Coinche.go validatePlay).
  // Ring-only, additive — illegal cards stay clickable and the backend still validates.
  const legalPlayIndices =
    isHumanTurn && humanPlayer
      ? coincheLegalPlayIndices(humanPlayer.cards, state.currentTrick, state.trumpSuit, humanIdx)
      : undefined;
  // Coinche の切り札はどのスートでも宣言できる (ベロートのような表向き札の
  // 除外は無い)。
  const allSuits = [1, 2, 3, 4];

  return (
    <GamePageShell
      title={tc('nav.coinche')}
      gameThemeBg={gameTheme.coinche.bg}
      phaseName={phaseNames[state.phase]}
      isHumanTurn={isHumanTurn || isHumanBidTurn}
      gamePath="/coinche"
      gameEndFlag={!!state?.gameEndFlag}
      loading={loading}
      confirmOpen={confirmOpen}
      confirmReset={confirmReset}
      cancelReset={cancelReset}
    >
      <SettingsPanel
        title={t('settings.title')}
        groups={[
          {
            items: [
              {
                type: 'select',
                id: 'cpuDifficulty',
                label: t('settings.cpuDifficulty'),
                value: coincheConfig.cpuDifficulty,
                options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
                  value: o.value,
                  label: t(`settings.${o.label.toLowerCase()}`),
                })),
                onSelect: (v) => handleConfigChange('cpuDifficulty', v),
              },
              {
                type: 'select',
                id: 'targetScore',
                label: t('settings.targetScore'),
                value: coincheConfig.targetScore,
                options: TARGET_SCORE_OPTIONS.map((v) => ({ value: v, label: String(v) })),
                onSelect: (v) => handleConfigChange('targetScore', v),
              },
              {
                type: 'checkbox',
                id: 'enableBeloteRebelote',
                label: t('settings.enableBeloteRebelote'),
                checked: coincheConfig.enableBeloteRebelote,
                onToggle: (v) => handleToggle('enableBeloteRebelote', v),
              },
              hintCheckboxItem(tc, frontendHintEnabled, setFrontendHintEnabled),
            ],
          },
        ]}
      />

      <div className={`flex-1 overflow-y-auto pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`}>
        {/* Round/Trick/Trump info */}
        <div className="text-ds-text-primary text-center mb-2">
          <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
          <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
          <span>
            {state.trumpSuit > 0 ? t('trumpSuit', { suit: t(SUIT_LABEL_KEYS[state.trumpSuit]) }) : t('noTrump')}
          </span>
        </div>
        {/* Bonus trackers — Dix de Der lights up on trick 8; Belote/Rebelote
            lights up as soon as ANY team finishes playing K + Q of trump
            (the bonus is awarded to whichever team holds those cards, not
            specifically the maker — see internal/domain/Coinche.go #864). */}
        {state.trumpSuit > 0 &&
          (() => {
            const isLastTrick = state.trickNumber === 8;
            const hasCoincheBonus = state.roundBeloteBonus.some((b) => b > 0);
            return (
              <div
                className="text-center mb-2 flex justify-center gap-2 flex-wrap text-xs"
                data-testid="coinche-bonus-trackers"
              >
                {state.config.dixDeDer > 0 && (
                  <span
                    data-testid="dix-de-der-badge"
                    data-active={isLastTrick ? 'true' : undefined}
                    className={
                      isLastTrick
                        ? 'px-2 py-0.5 rounded-full font-medium border bg-ds-accent text-ds-text-on-accent border-ds-accent animate-pulse'
                        : 'px-2 py-0.5 rounded-full font-medium border bg-ds-surface text-ds-text-muted border-ds-border'
                    }
                  >
                    {/* 点数は設定から。訳文に +10 と書くと、設定を変えたとき
                        バッジだけが古い数字を出す (#5592)。 */}
                    👑 {t('tracker.dixDeDer', { points: state.config.dixDeDer })}
                  </span>
                )}
                {state.config.enableBeloteRebelote && (
                  <span
                    data-testid="coinche-rebelote-badge"
                    data-active={hasCoincheBonus ? 'true' : undefined}
                    className={`${
                      hasCoincheBonus
                        ? 'px-2 py-0.5 rounded-full font-medium border bg-ds-success text-ds-text-on-accent border-ds-success'
                        : 'px-2 py-0.5 rounded-full font-medium border bg-ds-surface text-ds-text-muted border-ds-border'
                    }${coincheJustConfirmed ? ' ring-2 ring-ds-success motion-safe:animate-pulse' : ''}`}
                  >
                    {t('tracker.coincheKing')} · {t('tracker.coincheQueen')}
                    {hasCoincheBonus ? ` ${t('tracker.coincheBonus')}` : ''}
                  </span>
                )}
              </div>
            );
          })()}

        {coincheJustConfirmed && (
          <div
            role="status"
            aria-live="polite"
            data-testid="coinche-bonus-confirmed"
            className="text-center mb-2 text-sm font-semibold text-ds-success motion-safe:animate-pulse"
          >
            {t('tracker.coincheConfirmed')}
          </div>
        )}

        {/* **契約と倍率は精算そのもの。** 出さないと、同じカード点でも
            勝ち負けが変わる理由が画面から読めない。 */}
        {state.contractPoints > 0 && (
          <div className="mb-3 text-center text-sm text-ds-text-muted" data-testid="co-contract">
            {t('contractLine', {
              points: state.contractPoints,
              team: state.makerTeam,
              mult: state.multiplier,
            })}
          </div>
        )}

        {/* CPU players */}
        <div className="mb-3">
          {state.players
            .filter((p) => !p.isHuman)
            .map((p) => (
              <div key={p.id} className="mb-1 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
                {playerName(p.id, p.isHuman)}: {t('team', { n: p.team })} | {t('cards', { count: p.cardCount })} |{' '}
                {t('trickCount', { count: p.trickCount })}
              </div>
            ))}
        </div>

        {/* Current trick */}
        <TrickDisplay
          currentTrick={state.currentTrick}
          players={state.players}
          cardWidth={cardWidth}
          label={t('currentTrick')}
          dataTutorial="be-trick-display"
        />

        {/* Team scores */}
        <div className="my-3 p-2 rounded bg-black/30" data-tutorial="be-score-table">
          <div className="text-ds-text-muted text-sm mb-1">{t('teamScores')}</div>
          <table className="w-full text-sm text-ds-text-muted">
            <thead>
              <tr>
                <th scope="col" className="text-left">
                  {t('team', { n: 0 })}
                </th>
                <th scope="col" className="text-center">
                  {t('team', { n: 1 })}
                </th>
              </tr>
            </thead>
            <tbody>
              <tr>
                <td className="text-ds-accent">{state.teamScores[0]}</td>
                <td className="text-center">{state.teamScores[1]}</td>
              </tr>
              <tr>
                <td className="text-xs">{t('roundPoints', { points: state.roundPoints[0] })}</td>
                <td className="text-center text-xs">{t('roundPoints', { points: state.roundPoints[1] })}</td>
              </tr>
            </tbody>
          </table>
          {(state.roundBeloteBonus[0] > 0 || state.roundBeloteBonus[1] > 0) && (
            <div className="text-xs text-ds-warning mt-1">{t('coincheRebelote')}</div>
          )}
        </div>

        <RoundScoreAnnouncement
          active={isRoundEnd || isGameEnd}
          entries={[
            {
              name: t('team', { n: 0 }),
              roundScore: state.roundPoints[0] + state.roundBeloteBonus[0],
              cumulativeScore: state.teamScores[0],
            },
            {
              name: t('team', { n: 1 }),
              roundScore: state.roundPoints[1] + state.roundBeloteBonus[1],
              cumulativeScore: state.teamScores[1],
            },
          ]}
        />

        <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />

        <FrontendHintTooltip hint={frontendHint} enabled={frontendHintEnabled} t={t} />

        <ActionLogSection
          isEndPhase={isGameEnd}
          actionLog={actionLog}
          showActionLog={showActionLog}
          hideActionLog={hideActionLog}
        />
      </div>

      <GameFooter className={`${gameTheme.coinche.footer} px-4 py-2.5`}>
        {humanPlayer && (
          <PlayerHandSection
            humanPlayer={humanPlayer}
            selectedCardIndices={selectedCardIndices}
            toggleCard={toggleCard}
            cardWidth={cardWidth}
            isMobile={isMobile}
            dataTutorialPrefix="be"
            legalIndices={legalPlayIndices}
          />
        )}

        <ErrorAlert message={error ?? hintError} onRetry={retry} />

        {/* ライブ領域は**常設**。hint がある間だけ現れる内側の要素に role/aria-live を
            付けると、領域と中身が同じコミットで DOM に入るので変化として扱われず、
            読み上げられないことがある (#5955, #6663)。 */}
        <div data-testid="coinche-hint-live" role="status" aria-live="polite">
          {hint && (
            <div className="text-ds-warning text-sm mb-2">
              {/* hint.reason is a raw backend identifier; translate via hintReason.*,
                  falling back to a generic label. The hint shape depends on the phase:
                  orderUp (take/pass) and suit (call trump) during bidding, cardIndex in play. */}
              {(() => {
                const reason = t(`hintReason.${hint.reason}`, { defaultValue: t('hintReason.fallback') });
                // **点だけ言って何で取るのか言わない助言にしない。** 契約は
                // 目標点と切り札の対なので、どちらか欠けたら助言にならない。
                if (hint.bid !== undefined && hint.suit !== undefined) {
                  return `${t('hintBid', { points: hint.bid, suit: t(SUIT_LABEL_KEYS[hint.suit]) })} (${reason})`;
                }
                if (hint.cardIndex !== undefined) {
                  return `${t('hintPlay')}: [${hint.cardIndex}] (${reason})`;
                }
                return reason;
              })()}
            </div>
          )}
        </div>

        <div className="flex gap-2 items-center flex-wrap" data-tutorial="be-play-button">
          {isHumanBidTurn && (
            <span data-tutorial="be-bid-controls" className="flex gap-2 flex-wrap items-center">
              <select
                className="rounded bg-black/40 px-2 py-1 text-ds-text-primary"
                value={selectedPoints ?? ''}
                onChange={(e) => setSelectedPoints(Number(e.target.value))}
                disabled={loading || biddablePoints.length === 0}
                aria-label={t('bidPointsLabel')}
              >
                <option value="">{t('bidPointsLabel')}</option>
                {biddablePoints.map((pts) => (
                  <option key={pts} value={pts}>
                    {pts === COINCHE_CAPOT_POINTS ? t('capotOption', { points: pts }) : pts}
                  </option>
                ))}
              </select>
              {/* 目標点を選ぶまでスートは押せない。契約は対なので、片方だけ
                  送ると残りに既定値が入って別の契約になる。 */}
              {allSuits.map((s) => (
                <button
                  key={s}
                  type="button"
                  className={btnPrimary}
                  onClick={() => selectedPoints !== null && handleBid(selectedPoints, s)}
                  disabled={loading || selectedPoints === null}
                  aria-label={t('bidButtonAriaLabel', { suit: t(SUIT_LABEL_KEYS[s]) })}
                >
                  {t(SUIT_LABEL_KEYS[s])}
                </button>
              ))}
              <button type="button" className={btnSuccess} onClick={handlePass} disabled={loading}>
                {t('passButton')}
              </button>
            </span>
          )}

          {isHumanDoubleTurn && (
            <span data-tutorial="be-bid-controls" className="flex gap-2 flex-wrap">
              {canCoinche && (
                <button type="button" className={btnPrimary} onClick={handleCoinche} disabled={loading}>
                  {t('coincheButton')}
                </button>
              )}
              {canSurcoinche && (
                <button type="button" className={btnPrimary} onClick={handleSurcoinche} disabled={loading}>
                  {t('surcoincheButton')}
                </button>
              )}
              <button type="button" className={btnSuccess} onClick={handleDeclineDouble} disabled={loading}>
                {t('declineDoubleButton')}
              </button>
            </span>
          )}

          {(isHumanTurn || isHumanBidTurn) && (
            <button type="button" className={btnSuccess} onClick={handleHint} disabled={loading || hintLoading}>
              {tc('button.hint')}
            </button>
          )}
          {isHumanTurn && (
            <button
              type="button"
              className={btnPrimary}
              onClick={handlePlay}
              disabled={loading || selectedCardIndices.length !== 1}
            >
              {t('playButton')}
            </button>
          )}
          {isTrickEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextTrick} disabled={loading}>
              {t('nextTrick')}
            </button>
          )}
          {isRoundEnd && (
            <button type="button" className={btnSuccess} onClick={handleNextRound} disabled={loading}>
              {t('nextRound')}
            </button>
          )}
          <GameResetButton
            isGameEnd={!!isGameEnd}
            onReset={handleManualReset}
            requestConfirm={requestConfirm}
            loading={loading}
            dataTutorial="be-reset-button"
          />
        </div>
      </GameFooter>
    </GamePageShell>
  );
}
