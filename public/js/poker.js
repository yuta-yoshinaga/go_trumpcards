$(document).ready(function () {
    // 選択中のカードインデックス
    var selectedIndices = [];
    // 現在のフェーズ (0=init, 1=deal, 2=end)
    var currentPhase = 0;

    // PokerPhase定数
    var PHASE_INIT = 0;
    var PHASE_DEAL = 1;
    var PHASE_END  = 2;

    // --- ボタンイベント ---

    $('#btn-reset').on('click', function () {
        pokerRequest({ command: 'reset' });
    });

    $('#btn-exchange').on('click', function () {
        pokerRequest({ command: 'exchange', indices: selectedIndices });
    });

    $('#btn-stand').on('click', function () {
        pokerRequest({ command: 'stand' });
    });

    // カード選択（プレイヤーエリアのみ）
    $('#player-cards').on('click', '.card-wrap', function () {
        if (currentPhase !== PHASE_DEAL) return;
        var idx = parseInt($(this).data('idx'));
        $(this).toggleClass('selected');
        var pos = selectedIndices.indexOf(idx);
        if (pos === -1) {
            selectedIndices.push(idx);
        } else {
            selectedIndices.splice(pos, 1);
        }
    });

    // --- API通信 ---

    function pokerRequest(payload) {
        $.ajax({
            url: './poker/exec',
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify(payload),
        }).done(function (response) {
            console.log(response);
            updateUi(response);
        }).fail(function () {
            console.log('request failed');
        });
    }

    // --- UI更新 ---

    function updateUi(response) {
        currentPhase = response.phase;
        selectedIndices = [];

        renderPlayerCards(response.player);
        renderDealerCards(response.dealer, currentPhase);
        renderResult(response);
        updateButtons(currentPhase);
    }

    function renderPlayerCards(player) {
        var $area = $('#player-cards');
        $area.empty();
        if (!player || !player.cards) return;

        player.cards.forEach(function (card, idx) {
            var $wrap = $('<div class="card-wrap"></div>').attr('data-idx', idx);
            var $img = $('<img>').attr('src', getImagePath(card));
            var $label = $('<div class="select-label">交換</div>');
            $wrap.append($img).append($label);
            $area.append($wrap);
        });

        // ハンド名表示（ゲーム終了時のみ）
        if (currentPhase === PHASE_END && player.handName) {
            $('#player-hand-name').text(player.handName).show();
        } else {
            $('#player-hand-name').hide();
        }

        // フェーズヒント
        if (currentPhase === PHASE_DEAL) {
            $('#phase-hint').text('交換したいカードをクリックして選択し、「交換」または「スタンド」を押してください。');
        } else {
            $('#phase-hint').text('');
        }
    }

    function renderDealerCards(dealer, phase) {
        var $area = $('#dealer-cards');
        $area.empty();

        if (phase === PHASE_END && dealer && dealer.cards && dealer.cards.length > 0) {
            dealer.cards.forEach(function (card) {
                var $wrap = $('<div class="dealer-card-wrap"></div>');
                var $img = $('<img>').attr('src', getImagePath(card));
                $wrap.append($img);
                $area.append($wrap);
            });
            if (dealer.handName) {
                $('#dealer-hand-name').text(dealer.handName).show();
            }
        } else {
            // 伏せカード表示
            for (var i = 0; i < 5; i++) {
                var $wrap = $('<div class="dealer-card-wrap"></div>');
                var $img = $('<img>').attr('src', './images/z01.png');
                $wrap.append($img);
                $area.append($wrap);
            }
            $('#dealer-hand-name').hide();
        }
    }

    function renderResult(response) {
        var $box = $('#result-box');
        if (response.message && response.message !== '') {
            $box.text(response.message);
        } else {
            $box.text('');
        }
    }

    function updateButtons(phase) {
        if (phase === PHASE_DEAL) {
            $('#btn-exchange').prop('disabled', false);
            $('#btn-stand').prop('disabled', false);
            $('#btn-reset').prop('disabled', false);
        } else {
            $('#btn-exchange').prop('disabled', true);
            $('#btn-stand').prop('disabled', true);
            $('#btn-reset').prop('disabled', false);
        }
    }

    // --- 画像パス生成 ---

    function getImagePath(card) {
        var prefix = '';
        if (card.design === 'SPADE') {
            prefix = 's';
        } else if (card.design === 'CLOVER') {
            prefix = 'c';
        } else if (card.design === 'HEART') {
            prefix = 'h';
        } else if (card.design === 'DIAMOND') {
            prefix = 'd';
        } else {
            prefix = 'x';
        }
        return './images/' + prefix + zeroPadding(card.value, 2) + '.png';
    }

    function zeroPadding(num, length) {
        return ('0000000000' + num).slice(-length);
    }

    // --- 初期化 ---
    $('#btn-reset').click();
});
