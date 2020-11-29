$(document).ready(function () {
    $('.btn_field').on('click', '.reset', function () {
        $.ajax({
            url: "./blackjac/exec",
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                command: 'reset',
            }),
        }).done(function (response, textStatus, jqXHR) {
            // 成功時処理
            // レスポンスデータはパースされた上でresponseに渡される
            console.log('done');
            console.log(response);
            setUi(response);
        }).fail(function (jqXHR, textStatus, errorThrown) {
            // 失敗時処理
            console.log('fall');
        }).always(function (data_or_jqXHR, textStatus, jqXHR_or_errorThrown) {
            // doneまたはfail実行後の共通処理
        });
    });
    $('.btn_field').on('click', '.hit', function () {
        $.ajax({
            url: "./blackjac/exec",
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                command: 'hit',
            }),
        }).done(function (response, textStatus, jqXHR) {
            // 成功時処理
            // レスポンスデータはパースされた上でresponseに渡される
            console.log('done');
            console.log(response);
            setUi(response);
        }).fail(function (jqXHR, textStatus, errorThrown) {
            // 失敗時処理
            console.log('fall');
        }).always(function (data_or_jqXHR, textStatus, jqXHR_or_errorThrown) {
            // doneまたはfail実行後の共通処理
        });
    });
    $('.btn_field').on('click', '.stand', function () {
        $.ajax({
            url: "./blackjac/exec",
            type: 'POST',
            contentType: 'application/json',
            data: JSON.stringify({
                command: 'stand',
            }),
        }).done(function (response, textStatus, jqXHR) {
            // 成功時処理
            // レスポンスデータはパースされた上でresponseに渡される
            console.log('done');
            console.log(response);
            setUi(response);
        }).fail(function (jqXHR, textStatus, errorThrown) {
            // 失敗時処理
            console.log('fall');
        }).always(function (data_or_jqXHR, textStatus, jqXHR_or_errorThrown) {
            // doneまたはfail実行後の共通処理
        });
    });
    $('.reset').click();
});

function setUi(response) {
    $('.trumpcards_field').empty();
    var addEle = '<div>'
    addEle += '<div><h3 class="card-header">ディーラー手札</h3>'
    addEle += '<h3 class="card-header">スコア ' + (response.dealer.score !== 0 ? response.dealer.score : '') + '</h3>'
    response.dealer.cards.forEach((card) => {
        addEle += '<div class="col-xs-3"><image class="img-responsive" src="' + getImagePath(card) + '" /></div>';
    })
    if (response.dealer.score === 0) {
        addEle += '<div class="col-xs-3"><image class="img-responsive" src="./images/z01.png" /></div>';
    }
    addEle += '<div class="clearfix"></div></div>'
    $('.trumpcards_field').append(addEle);

    addEle = '<div>'
    addEle += '<h3 class="card-header">プレイヤー手札</h3>'
    addEle += '<h3 class="card-header">スコア ' + response.player.score + '</h3>'
    response.player.cards.forEach((card) => {
        addEle += '<div class="col-xs-3"><image class="img-responsive" src="' + getImagePath(card) + '" /></div>';
    })
    addEle += '<div class="clearfix"></div></div>'
    $('.trumpcards_field').append(addEle);

    if (response.message !== '') {
        alert(response.message);
    }
}

function getImagePath(card) {
    var resImage = './images/';
    if (card.design == 'JOKER') {
        // *** ジョーカー *** //
        resImage += 'x';
    } else if (card.design == 'SPADE') {
        // *** スペード *** //
        resImage += 's';
    } else if (card.design == 'CLOVER') {
        // *** クローバー *** //
        resImage += 'c';
    } else if (card.design == 'HEART') {
        // *** ハート *** //
        resImage += 'h';
    } else if (card.design == 'DIAMOND') {
        // *** ダイアモンド *** //
        resImage += 'd';
    }
    resImage += String(zeroPadding(card.value, 2)) + '.png';
    return resImage;
}

function zeroPadding(num, length) {
    return ('0000000000' + num).slice(-length);
}