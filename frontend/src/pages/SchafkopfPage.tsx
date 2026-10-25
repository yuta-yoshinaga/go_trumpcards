import { useEffect } from 'react';
import type { schafkopfApi } from '../api/gameApi';
import { GameMessageBox } from '../components/GameMessageBox';
import { createGamePage, type GameDefinition, type GamePageContext } from '../components/game/createGamePage';
import { FrontendHintTooltip } from '../components/hint/FrontendHintTooltip';
import { PlayerHandSection } from '../components/PlayerHandSection';
import { TrickDisplay } from '../components/TrickDisplay';
import { useCardDimensions } from '../hooks/useCardDimensions';
import { usePhaseNames } from '../hooks/usePhaseNames';
import { CPU_DIFFICULTY_OPTIONS, TARGET_CHIPS_OPTIONS, useSchafkopfGame } from '../hooks/useSchafkopfGame';
import { badgeWarningColors } from '../styles/badgeStyles';
import { btnPrimary, btnSecondary, btnSuccess } from '../styles/buttonStyles';
import { lgCardAreaConstraint, lgTwoColGrid } from '../styles/gameStyles';
import type { SchafkopfResponse } from '../types/card';
import { SchafkopfPhase } from '../types/phases';
import type { TutorialStep } from '../types/tutorial';
import { cardAlt } from '../utils/cardAlt';
import { parseSchafkopfCommand, SCHAFKOPF_HELP } from '../utils/cli/commands/schafkopfCommands';
import { formatSchafkopfState } from '../utils/cli/formatters/schafkopfFormatter';
import { formatSignedDelta } from '../utils/formatSignedDelta';
import { isRequestedHint } from '../utils/hintRequest';
import { playerName } from '../utils/playerUtils';
import { hintCheckboxItem } from '../utils/settingsItems';

/** Schafkopf tutorial step definitions. */
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

const SCHAFKOPF_PHASE_KEYS: Readonly<Record<number, string>> = {
  [SchafkopfPhase.PICK]: 'pick',
  [SchafkopfPhase.CALL]: 'call',
  [SchafkopfPhase.PLAY]: 'play',
  [SchafkopfPhase.TRICK_END]: 'trickEnd',
  [SchafkopfPhase.ROUND_END]: 'roundEnd',
  [SchafkopfPhase.GAME_END]: 'gameEnd',
};

/**
 * Maps a suit id to its i18n suit key. Hearts are trump under Rufspiel, so a
 * call never names them; a Solo can name any of the four.
 */
const SUIT_KEYS: Readonly<Record<number, string>> = { 1: 'spade', 2: 'club', 3: 'heart', 4: 'diamond' };

/** Suits a Solo may name as trump. */
const SOLO_SUITS = [1, 2, 3, 4] as const;

/** Contract ids, matching `SchafkopfContract` in the Go domain. */
const CONTRACT_RUFSPIEL = 0;
const CONTRACT_WENZ = 1;
const CONTRACT_SOLO = 2;

type SchafkopfGame = ReturnType<typeof useSchafkopfGame>;
type SchafkopfContext = GamePageContext<SchafkopfResponse, SchafkopfGame, SchafkopfView>;

function useSchafkopfView({ game }: { game: SchafkopfGame }) {
  // Fetch a fresh game on mount.
  // biome-ignore lint/correctness/useExhaustiveDependencies: reset is stable per render of the hook; run once on mount.
  useEffect(() => {
    game.reset();
  }, []);

  const { cardWidth, isMobile } = useCardDimensions();
  const phaseNames = usePhaseNames('schafkopf', SCHAFKOPF_PHASE_KEYS);
  return { cardWidth, isMobile, phaseNames };
}

type SchafkopfView = ReturnType<typeof useSchafkopfView>;

function Board({ game, state, t, view }: SchafkopfContext) {
  const { cardWidth, isMobile } = view;
  const loading = game.loading;

  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;

  const isPickPhase = state.phase === SchafkopfPhase.PICK;
  const isPlayPhase = state.phase === SchafkopfPhase.PLAY;
  const isTrickEnd = state.phase === SchafkopfPhase.TRICK_END;
  const isRoundEnd = state.phase === SchafkopfPhase.ROUND_END;
  const isGameEnd = state.phase === SchafkopfPhase.GAME_END || state.gameEndFlag;

  // Phase-specific "it's the human's actionable turn" flags.
  const canPick = isPickPhase && isHumanTurn;

  // Partner is only shown once revealed (or at round/game end).
  const showPartner = state.partnerRevealed || isRoundEnd || isGameEnd;
  // The contract is what counts as trump, so it has to be on screen: without
  // it there is no way to tell why an Ober is not trump in a Wenz.
  // 上回れない契約はボタンに出さない。押せるのに必ず拒否される操作面を作らない。
  const beatable = state.beatableContracts ?? [];
  const contractLabel =
    state.contract === 1
      ? t('contractWenz')
      : state.contract === 2
        ? t('contractSolo', { suit: t(`suit.${SUIT_KEYS[state.soloSuit] ?? 'spade'}`) })
        : t('contractRufspiel');
  const pickerLabel = state.pickerIdx >= 0 ? playerName(state.pickerIdx, state.pickerIdx === humanIdx) : '-';

  // **ピックの順番は親の左隣から回る。** 誰が親かが出ていないと、自分が
  // 何番目に「やる／降りる」を決めるのか盤から読めない。サーバは最初から
  // `dealerIdx` を送っているのに、どちらの一覧も読んでいなかった (#6617)。
  // モバイルとデスクトップで同じものを出すので、閉包はここに1つだけ置く。
  const dealerBadge = (id: number) =>
    id === state.dealerIdx ? (
      <span
        className={`ml-1 rounded px-1.5 py-0.5 text-xs ${badgeWarningColors}`}
        data-testid={`schafkopf-dealer-${id.toString()}`}
      >
        <span aria-hidden="true">{t('dealerBadge')}</span>
        <span className="sr-only">{t('dealerAria')}</span>
      </span>
    ) : null;
  const partnerLabel =
    showPartner && state.partnerIdx >= 0
      ? playerName(state.partnerIdx, state.partnerIdx === humanIdx)
      : t('partnerHidden');

  return (
    <>
      <div className="text-ds-text-primary text-center mb-2">
        <span className="mr-4">{t('round', { n: state.roundNumber })}</span>
        <span className="mr-4">{t('trick', { n: state.trickNumber })}</span>
        <span>{t('contractLabel', { contract: contractLabel })}</span>
      </div>

      {canPick && (
        <div className="grid grid-cols-2 sm:grid-cols-4 gap-1 items-center" data-tutorial="sh-action-buttons">
          {/* Only the contracts that outrank the standing bid: a button
            the server is bound to refuse is worse than no button. */}
          {beatable.includes(CONTRACT_RUFSPIEL) && (
            <button type="button" className={btnPrimary} onClick={game.handlePick} disabled={loading}>
              {t('pickButton')}
            </button>
          )}
          {beatable.includes(CONTRACT_WENZ) && (
            <button type="button" className={btnPrimary} onClick={() => game.handleDeclare(1)} disabled={loading}>
              {t('wenzButton')}
            </button>
          )}
          {/* Solo names its own trump, so each suit is its own button
            rather than a declare-then-choose second step. */}
          {beatable.includes(CONTRACT_SOLO) &&
            SOLO_SUITS.map((suit) => {
              const suitName = t(`suit.${SUIT_KEYS[suit]}`);
              return (
                <button
                  key={`solo-${suit}`}
                  type="button"
                  className={btnPrimary}
                  onClick={() => game.handleDeclare(2, suit)}
                  disabled={loading}
                  aria-label={t('soloButtonAriaLabel', { suit: suitName })}
                >
                  {t('soloButton', { suit: suitName })}
                </button>
              );
            })}
        </div>
      )}

      {(isPlayPhase || isTrickEnd) && (
        <div className="mb-2 text-center text-sm text-ds-text-muted" data-testid="schafkopf-round-points">
          {t(state.roundPickerPointsTeamTotal ? 'roundProgress.pickerPoints' : 'roundProgress.pickerOnlyPoints', {
            points: state.roundPickerPoints,
            target: state.pickerTargetPoints,
          })}
        </div>
      )}

      <div className={lgTwoColGrid}>
        {/* Left: play area */}
        <div>
          <TrickDisplay
            currentTrick={state.currentTrick}
            players={state.players}
            cardWidth={cardWidth}
            label={t('currentTrick')}
            winnerIdx={isTrickEnd && state.winnerIdx >= 0 ? state.winnerIdx : undefined}
            winnerLabel={
              isTrickEnd && state.winnerIdx >= 0
                ? t('trickWinner', { name: playerName(state.winnerIdx, state.players[state.winnerIdx].isHuman) })
                : undefined
            }
            cardAriaLabelFor={(player, card) =>
              t('trickCardByPlayer', { name: playerName(player.id, player.isHuman), card: cardAlt(card) })
            }
            dataTutorial="sh-trick-display"
          />
          <div className="sr-only" aria-live="polite" aria-atomic="true">
            {isTrickEnd && state.winnerIdx >= 0
              ? t('trickWinner', {
                  name: playerName(state.winnerIdx, state.players[state.winnerIdx].isHuman),
                })
              : isPlayPhase
                ? t('leadSeat', {
                    name: playerName(state.leadPlayerIdx, state.players[state.leadPlayerIdx].isHuman),
                  })
                : ''}
          </div>
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
                    {t('cards', { count: p.cardCount })} | {t('tricks', { count: p.trickCount })}
                    {dealerBadge(p.id)}
                  </div>
                ))}
              </div>
            </details>
          ) : (
            <div className="mb-2 p-2 rounded bg-black/30">
              {state.players.map((p) => (
                <div key={p.id} className="text-ds-text-muted text-sm py-0.5">
                  {playerName(p.id, p.isHuman)}: {t('chips', { count: p.chips })} | {t('cards', { count: p.cardCount })}{' '}
                  | {t('tricks', { count: p.trickCount })}
                  {dealerBadge(p.id)}
                </div>
              ))}
            </div>
          )}

          {/* Round result */}
          {(isRoundEnd || isGameEnd) && (
            <div className="my-3 p-2 rounded bg-black/30 text-ds-text-muted text-sm">
              <div className="mb-1 text-ds-text-primary">{t('roundResult.title')}</div>
              <div>{t('roundResult.pickerPoints', { points: state.roundPickerPoints })}</div>
              <div>{t('roundResult.multiplier', { multiplier: state.roundMultiplier })}</div>
              <div>{state.roundPickerWon ? t('roundResult.pickerWon') : t('roundResult.pickerLost')}</div>
              <div className="mt-1">{t('roundResult.chipChanges')}</div>
              {state.players.map((p) => {
                const delta = p.chipDelta;
                const formatted = formatSignedDelta(delta);
                return (
                  <div key={p.id}>
                    {playerName(p.id, p.isHuman)}: {formatted}
                  </div>
                );
              })}
            </div>
          )}
        </div>
      </div>

      {/* Message */}
      <GameMessageBox message={state.message} messageCode={state.messageCode} messageParams={state.messageParams} />
    </>
  );
}

function FooterHand({ game, state, t, view }: SchafkopfContext) {
  const humanPlayer = state.players.find((p) => p.isHuman);
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;
  const canPlay = state.phase === SchafkopfPhase.PLAY && isHumanTurn;

  return humanPlayer ? (
    <PlayerHandSection
      humanPlayer={humanPlayer}
      selectedCardIndices={game.selectedCardIndices}
      toggleCard={game.toggleCard}
      cardWidth={view.cardWidth}
      isMobile={view.isMobile}
      dataTutorialPrefix="sh"
      validIndices={canPlay ? state.playableIndices : undefined}
      // 出せる札にリングを出す。validIndices は出せない札を伏せる
      // だけで、どれが出せるかを目で見つける手がかりにならない。
      legalIndices={canPlay ? state.playableIndices : undefined}
      restrictedTooltip={t('playButton')}
    />
  ) : null;
}

function FooterStatus({ state, t, frontendHint }: SchafkopfContext) {
  return (
    <>
      {/*
        ライブ領域は**常設**。hint がある間だけ現れる内側の div に付けると、
        領域と中身が同じコミットで DOM に入るので変化として扱われず、読み上げ
        られないことがある (#5955)。
      */}
      <div data-testid="schafkopf-hint-live" role="status" aria-live="polite">
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

function primaryActions({ game, state, t, busy }: SchafkopfContext) {
  const humanIdx = state.players.findIndex((p) => p.isHuman);
  const isHumanTurn = state.currentPlayerIdx === humanIdx;

  const isPickPhase = state.phase === SchafkopfPhase.PICK;
  const isCallPhase = state.phase === SchafkopfPhase.CALL;
  const isPlayPhase = state.phase === SchafkopfPhase.PLAY;
  const isTrickEnd = state.phase === SchafkopfPhase.TRICK_END;
  const isRoundEnd = state.phase === SchafkopfPhase.ROUND_END;

  // Phase-specific "it's the human's actionable turn" flags.
  const canPick = isPickPhase && isHumanTurn;
  const canCall = isCallPhase && state.pickerIdx === humanIdx && state.callableSuits.length > 0;
  const canPlay = isPlayPhase && isHumanTurn;

  if (!canPick && !canCall && !canPlay && !isTrickEnd && !isRoundEnd) return null;

  return (
    <div className="flex flex-wrap gap-1 items-center">
      {canPick && (
        <button type="button" className={btnSecondary} onClick={game.handlePass} disabled={busy}>
          {t('passButton')}
        </button>
      )}
      {canCall &&
        state.callableSuits.map((suit) => {
          // callableSuits are the three fail suits (♠♣♦), all present in SUIT_KEYS.
          const suitName = t(`suit.${SUIT_KEYS[suit]}`);
          return (
            <button
              key={`call-${suit}`}
              type="button"
              className={btnPrimary}
              onClick={() => game.handleCall(suit)}
              disabled={busy}
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
          disabled={busy || game.selectedCardIndices.length !== 1}
        >
          {t('playButton')}
        </button>
      )}
      {isTrickEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextTrick} disabled={busy}>
          {t('nextTrick')}
        </button>
      )}
      {isRoundEnd && (
        <button type="button" className={btnSuccess} onClick={game.handleNextRound} disabled={busy}>
          {t('nextRound')}
        </button>
      )}
    </div>
  );
}

const schafkopfDefinition: GameDefinition<
  SchafkopfResponse,
  SchafkopfGame,
  Parameters<typeof schafkopfApi.exec>,
  SchafkopfView
> = {
  key: 'schafkopf',
  useGame: useSchafkopfGame,
  useView: useSchafkopfView,
  tutorial: SH_TUTORIAL_STEPS,
  skeleton: { kind: 'trick-taking', trickArea: true, footerHandSize: 6 },
  cli: () => ({
    parseCommand: parseSchafkopfCommand,
    formatResponse: formatSchafkopfState,
    helpText: SCHAFKOPF_HELP,
  }),
  phase: ({ state, view }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const isGameEnd = state.phase === SchafkopfPhase.GAME_END || state.gameEndFlag;
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
          value: game.schafkopfConfig.cpuDifficulty,
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
          value: game.schafkopfConfig.targetChips,
          options: TARGET_CHIPS_OPTIONS.map((v) => ({ value: v, label: String(v) })),
          onSelect: (v) => game.handleConfigChange('targetChips', v),
        },
        hintCheckboxItem(tc, frontendHint.hintEnabled, frontendHint.setHintEnabled),
      ],
    },
  ],
  settingsPlacement: 'top',
  contentClassName: `pt-3 px-4 lg:px-8 ${lgCardAreaConstraint}`,
  primaryActions,
  // Keyboard hand navigation: number keys toggle a card, Enter confirms,
  // Escape clears. Only PLAY takes a card selection, and handlePlay
  // self-guards on the selected count.
  cardNav: ({ game, state }) => {
    const humanIdx = state.players.findIndex((p) => p.isHuman);
    const canPlay = state.phase === SchafkopfPhase.PLAY && state.currentPlayerIdx === humanIdx;
    return {
      cardCount: state.players.find((p) => p.isHuman)?.cards?.length ?? 0,
      onToggle: game.toggleCard,
      onConfirm: game.handlePlay,
      onClear: game.clearSelection,
      enabled: canPlay,
    };
  },
  shortcutsTestId: 'schafkopf-kbd-shortcuts',
  resetTutorialId: 'sh-reset-button',
  // The hand-wired page rendered a bare reset button; a class-less div lays out the same, unlike the default flex row.
  actionsClassName: '',
  Board,
  FooterHand,
  FooterStatus,
};

/** Renders the Schafkopf game page: 4-player partnership trick-taking with declare/call/play phases. */
export const SchafkopfPage = createGamePage(schafkopfDefinition);
