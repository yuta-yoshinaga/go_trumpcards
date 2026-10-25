import { useEffect, useRef, useState } from 'react';
import type { sheepsheadApi } from '../api/gameApi';
import { CardImage } from '../components/CardImage';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { AnimatedCardBack } from '../components/motion/AnimatedCardBack';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import {
  BASE_CHIPS_OPTIONS,
  CPU_DIFFICULTY_OPTIONS,
  SHEEPSHEAD_BURY_COUNT,
  TARGET_CHIPS_OPTIONS,
  useSheepsheadGame,
} from '../hooks/useSheepsheadGame';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { SheepsheadResponse } from '../types/card';
import { SheepsheadPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { parseSheepsheadCommand, SHEEPSHEAD_HELP } from '../utils/cli/commands/sheepsheadCommands';
import { formatSheepsheadState } from '../utils/cli/formatters/sheepsheadFormatter';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Sheepshead tutorial step definitions. */
const SH_TUTORIAL_STEPS: TutorialStep[] = [
  {
    target: '[data-tutorial="sh-trick-display"]',
    messageKey: 'tutorial.trickDisplay',
    placement: 'bottom',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-player-hand"]',
    messageKey: 'tutorial.playerHand',
    placement: 'top',
    advanceOn: 'next',
  },
  {
    target: '[data-tutorial="sh-action-buttons"]',
    messageKey: 'tutorial.playButton',
    placement: 'top',
    advanceOn: 'next',
  },
  { target: '[data-tutorial="sh-info"]', messageKey: 'tutorial.info', placement: 'top', advanceOn: 'next' },
  {
    target: '[data-tutorial="sh-reset-button"]',
    messageKey: 'tutorial.resetButton',
    placement: 'top',
    advanceOn: 'next',
  },
];

const SHEEPSHEAD_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SheepsheadPhase.PICK]: 'pick',
  [SheepsheadPhase.BURY]: 'bury',
  [SheepsheadPhase.CALL]: 'call',
  [SheepsheadPhase.PLAY]: 'play',
  [SheepsheadPhase.TRICK_END]: 'trickEnd',
  [SheepsheadPhase.ROUND_END]: 'roundEnd',
  [SheepsheadPhase.GAME_END]: 'gameEnd',
};

/** Maps a called-suit id (1=♠ 2=♣ 3=♥) to its i18n suit key. */
const SUIT_KEYS: Readonly<Record<number, string>> = { 1: 'spade', 2: 'club', 3: 'heart' };

type SheepsheadGame = ReturnType<typeof useSheepsheadGame>;
type SheepsheadView = ReturnType<typeof useSheepsheadView>;
type SheepsheadContext = GamePageContext<SheepsheadResponse, SheepsheadGame, SheepsheadView>;

function useSheepsheadView({
  game,
  t,
}: {
  game: SheepsheadGame;
  t: (key: string, opts?: Record<string, unknown>) => string;
}) {
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('sheepshead', SHEEPSHEAD_PHASE_KEYS);
  const [trickAnnouncement, setTrickAnnouncement] = useState('');
  const previousCompletedTrickCount = useRef<number | null>(null);

  useEffect(() => {
    const state = game.state;
    if (!state) return;
    if (previousCompletedTrickCount.current === null) {
      previousCompletedTrickCount.current = state.completedTrickCount;
      return;
    }
    if (state.completedTrickCount > previousCompletedTrickCount.current && state.lastTrickWinner >= 0) {
      setTrickAnnouncement(
        t('trickWonAnnouncement', {
          name: playerName(state.lastTrickWinner, state.players[state.lastTrickWinner]?.isHuman ?? false),
        }),
      );
    }
    previousCompletedTrickCount.current = state.completedTrickCount;
  }, [game.state, t]);

  return { cardWidth, isMobile, phaseNames, trickAnnouncement };
}

function LiveRegion({ view }: SheepsheadContext) {
  return (
    <div className="sr-only" role="status" aria-live="polite" data-testid="sheepshead-trick-live">
      {view.trickAnnouncement}
    </div>
  );
}

function Board({ state, t, view }: SheepsheadContext) {
  const { cardWidth, isMobile } = view;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isPickPhase = state.phase === SheepsheadPhase.PICK;
  const isTrickEnd = state.phase === SheepsheadPhase.TRICK_END;
  const isRoundEnd = state.phase === SheepsheadPhase.ROUND_END;
  const isGameEnd = state.phase === SheepsheadPhase.GAME_END || state.gameEndFlag;
  const isPlayPhase = state.phase === SheepsheadPhase.PLAY;

  // Partner is only shown once revealed (or at round/game end).
  const showPartner = state.partnerRevealed || isRoundEnd || isGameEnd;
  const pickerLabel = state.pickerIdx >= 0 ? playerName(state.pickerIdx, state.pickerIdx === humanIdx) : '-';
  const partnerLabel =
    showPartner && state.partnerIdx >= 0
      ? playerName(state.partnerIdx, state.partnerIdx === humanIdx)
      : t('partnerHidden');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('blindCount', { count: state.blindCount })}</span>
      </div>

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          {/* Blind: face-down cards, shown only during the Pick phase before a picker takes them. */}
          {isPickPhase && state.blindCount > 0 && (
            <div className="mb-3" data-testid="sh-blind-display">
              <div className="text-ds-text-muted text-sm mb-1 text-center">
                {t('blind')} ({state.blindCount})
              </div>
              <div className="flex justify-center gap-2">
                {Array.from({ length: state.blindCount }).map((_, i) => (
                  <AnimatedCardBack key={i} width={cardWidth} dealDelay={i * 0.08} />
                ))}
              </div>
            </div>
          )}
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            dataTutorial="sh-trick-display"
            cardDetailFor={(_card, trickCard) => {
              return t('trickCardDetail', {
                points: trickCard.points,
                trump: t(trickCard.isTrump ? 'trump' : 'notTrump'),
              });
            }}
            cardDetailClassName="break-keep [overflow-wrap:anywhere]"
          />
          {isTrickEnd && state.leadPlayerIdx >= 0 && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm" data-testid="sh-trick-result">
              <div>
                {t('trickResult.winner', {
                  player: playerName(state.leadPlayerIdx, state.leadPlayerIdx === humanIdx),
                })}
              </div>
              <div>
                {t('trickResult.nextLead', {
                  player: playerName(state.leadPlayerIdx, state.leadPlayerIdx === humanIdx),
                })}
              </div>
            </div>
          )}
        </div>

        {/* Right: info sidebar */}
        <div data-tutorial="sh-info">
          {/* Picker / partner / called suit */}
          <div className="mb-2 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
            <div>
              {t('picker')}: {pickerLabel}
            </div>
            <div>
              {t('partner')}: {partnerLabel}
            </div>
            <div>
              {t('calledSuit')}:{' '}
              {state.calledSuit > 0 && SUIT_KEYS[state.calledSuit]
                ? t(`suit.${SUIT_KEYS[state.calledSuit]}`)
                : t('none')}
            </div>
            {isPickPhase && <div>{t('passCount', { count: state.passCount })}</div>}
          </div>

          {/* Players: chips / cards / tricks */}
          {isMobile ? (
            <details className="mb-2 p-2 rounded bg-black/30">
              <summary className="cursor-pointer select-none text-ds-text-muted text-sm">{t('players')}</summary>
              <div className="mt-1">
                {state.players.map((p) => (
                  <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                    {playerName(p.id, p.isHuman)}: {t('chips', { count: p.chips })} |{' '}
                    {t('tricks', { count: p.trickCount })}
                    {showPartner && (
                      <span>
                        {' '}
                        | {p.id === state.pickerIdx || p.id === state.partnerIdx ? t('pickerTeam') : t('defenderTeam')}
                      </span>
                    )}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('chips', { count: p.chips })} |{' '}
                  {t('tricks', { count: p.trickCount })}
                  {showPartner && (
                    <span>
                      {' '}
                      | {p.id === state.pickerIdx || p.id === state.partnerIdx ? t('pickerTeam') : t('defenderTeam')}
                    </span>
                  )}
                </div>
              ))}
            </div>
          )}

          {(isPlayPhase || isTrickEnd) && (
            <div
              className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm"
              role="status"
              aria-live="polite"
              data-testid="sh-live-points"
            >
              <div className="mb-1 text-ds-text-primary">{t('livePoints.title')}</div>
              <div>{t('livePoints.picker', { points: state.livePickerPoints, target: 61 })}</div>
              <div>{t('livePoints.defenders', { points: state.liveDefenderPoints })}</div>
            </div>
          )}

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.pickerPoints', { points: state.roundPickerPoints })}</div>
              <div>{t('roundResult.multiplier', { multiplier: state.roundMultiplier })}</div>
              <div>{state.roundPickerWon ? t('roundResult.pickerWon') : t('roundResult.pickerLost')}</div>
              {/* 埋め札はラウンド終了で公開される。得点の内訳そのものなので出す (#5638)。 */}
              {state.buried.length > 0 && (
                <div className="mt-1 flex items-center gap-2">
                  <span>{t('roundResult.buried')}</span>
                  {state.buried.map((c) => (
                    <CardImage key={`buried-${c.design}-${c.value}`} card={c} width={cardWidth * 0.6} />
                  ))}
                </div>
              )}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: SheepsheadContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const isPlayPhase = state.phase === SheepsheadPhase.PLAY;
  const canPlay = isPlayPhase && isHumanTurn;

  if (!humanPlayer) return null;

  return (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="sh"
      validIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  );
}

function FooterStatus({ state, t, frontendHint }: SheepsheadContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="sheepshead-hint-live" role="status" aria-live="polite">
        {state.hint && isRequestedHint(state) && (
          <div className="text-ds-warning text-sm mb-2">
            {t('hintAvailable')}: {t(`hint.${state.hint.reason}`)}
            {state.hint.cardIndices.length > 0 && ` (${state.hint.cardIndices.map((i) => `[${i}]`).join(', ')})`}
          </div>
        )}
      </div>
      <FrontendHintTooltip hint={frontendHint.hint} enabled={frontendHint.hintEnabled} t={t} />
    </>
  );
}

function FooterActions({ game, state, t, busy }: SheepsheadContext) {
  const loading = busy;
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const isPickPhase = state.phase === SheepsheadPhase.PICK;
  const isBuryPhase = state.phase === SheepsheadPhase.BURY;
  const isCallPhase = state.phase === SheepsheadPhase.CALL;
  const isPlayPhase = state.phase === SheepsheadPhase.PLAY;
  const isTrickEnd = state.phase === SheepsheadPhase.TRICK_END;
  const isRoundEnd = state.phase === SheepsheadPhase.ROUND_END;

  // Phase-specific "it's the human's actionable turn" flags.
  const canPick = isPickPhase && isHumanTurn;
  const canBury = isBuryPhase && state.pickerIdx === humanIdx;
  const canCall = isCallPhase && state.pickerIdx === humanIdx && state.callableSuits.length > 0;
  const canPlay = isPlayPhase && isHumanTurn;

  return (
    <>
      {canPick && (
        <>
          <button type="button" className={btnPrimary} onClick={game.handlePick} disabled={loading}>
            {t('pickButton')}
          </button>
          <button type="button" className={btnSecondary} onClick={game.handlePass} disabled={loading}>
            {t('passButton')}
          </button>
        </>
      )}
      {canBury && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handleBury}
          disabled={loading || game.selectedCardIndices.length !== SHEEPSHEAD_BURY_COUNT}
        >
          {t('buryButton')} ({t('burySelected', { count: game.selectedCardIndices.length })})
        </button>
      )}
      {canCall &&
        state.callableSuits.map((suit) => {
          // callableSuits are always 1-3 (♠♣♥), all present in SUIT_KEYS.
          const suitName = t(`suit.${SUIT_KEYS[suit]}`);
          return (
            <button
              key={`call-${suit}`}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleCall(suit)}
              disabled={loading}
              aria-label={t('callButtonAriaLabel', { suit: suitName })}
            >
              {t('callButton', { suit: suitName })}
            </button>
          );
        })}
      {canPlay && (
        <button
          type="button"
          className={btnPrimary}
          onClick={game.handlePlay}
          disabled={loading || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={loading}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={loading}>
          {t('nextRound')}
        </button>
      )}
    </>
  );
}

const sheepsheadDefinition: GameDefinition<
  SheepsheadResponse,
  SheepsheadGame,
  Parameters<typeof sheepsheadApi.exec>,
  SheepsheadView
> = {
  key: 'sheepshead',
  useGame: useSheepsheadGame,
  useView: useSheepsheadView,
  tutorial: SH_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 6 },
  cli: () => ({
    parseCommand: parseSheepsheadCommand,
    formatResponse: formatSheepsheadState,
    helpText: SHEEPSHEAD_HELP,
  }),
  phase: ({ state, view }) => {
    const isGameEnd = state.phase === SheepsheadPhase.GAME_END || state.gameEndFlag;
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      name: view.phaseNames[state.phase],
      ended: isGameEnd,
      won: isGameEnd && state.winnerIdx === humanIdx,
    };
  },
  shell: ({ state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    return {
      isHumanTurn: state.currentPlayerIdx === humanIdx,
    };
  },
  reset: ({ game }) => {
    game.reset();
  },
  settings: ({ game, t, tc, frontendHint }) => [
    {
      items: [
        {
          type: 'select',
          id: 'cpuDifficulty',
          label: t('settings.cpuDifficulty'),
          value: game.sheepsheadConfig.cpuDifficulty,
          options: CPU_DIFFICULTY_OPTIONS.map((o) => ({
            value: o.value,
            label: t(`settings.${o.label.toLowerCase()}`),
          })),
          onSelect: (v) => game.handleConfigChange('cpuDifficulty', v),
        },
        {
          type: 'select',
          id: 'targetChips',
          label: t('settings.targetChips'),
          value: game.sheepsheadConfig.targetChips,
          options: TARGET_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetChips', v),
        },
        {
          type: 'select',
          id: 'baseChips',
          label: t('settings.baseChips'),
          value: game.sheepsheadConfig.baseChips,
          options: BASE_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('baseChips', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  cardNav: ({ game, state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const canBury = state.phase === SheepsheadPhase.BURY && state.pickerIdx === humanIdx;
    const canPlay = state.phase === SheepsheadPhase.PLAY && state.currentPlayerIdx === humanIdx;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: () => {
        if (canBury) game.handleBury();
        else if (canPlay) game.handlePlay();
      },
      onClear: game.clearSelection,
      enabled: canBury || canPlay,
    };
  },
  shortcutsTestId: 'sheepshead-kbd-shortcuts',
  resetTutorialId: 'sh-reset-button',
  actionsTutorialId: 'sh-action-buttons',
  actionsClassName: 'flex flex-wrap gap-2 items-center',
  LiveRegion,
  Board,
  FooterHand,
  FooterStatus,
  FooterActions,
};

/** Renders the Sheepshead game page: 5-player partnership trick-taking with pick/bury/call/play phases. */
export const SheepsheadPage = createGamePage(sheepsheadDefinition);
