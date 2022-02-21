# ゲームの通信

## ゲームの開始

エンドポイントに接続するとプレイヤーが待機していることになります。
両方のプレイヤーが待機状態になるとゲームが開始されます。

## プロトコル

サーバは以下のような JSON オブジェクトを返します。

```js
{
    "type": "イベントの種類",
    "data": {
        // イベント固有のデータが格納されます
    }
}
```

### イベント

#### サーバ→クライアント

##### `onStart`

ゲームの開始を通知します。

- データは空です。

ペイロード例:
```js
{
    "type": "onStart",
    "data": {}
}
```

##### `onTick`

ゲーム全体の残り秒数が変化するたび (つまり 1 秒ごと) に発生します。

- `remainSec`: 残り秒数が格納されます。
- `turnRemainSec`: ターンの残り秒数です。
- `finished`: ゲームが終了したかどうかが格納されます。
- `waitingRetry`: リトライ待ちかどうかが格納されます。

ペイロード例:
```js
{
    "type": "onTick",
    "data": {
        "remainSec": 5,
        "turnRemainSec": 2,
        "finished": false,
        "waitingRetry": false
    }
}
```

##### `onFailure`

どちらかが失敗し、リトライができない場合に発生します。

- データは空です。

ペイロード例:
```js
{
    "type": "onFailure",
    "data": {}
}
```

##### `onChangeTurn`

答えが入力され、ターンが変わったときに発生します。

- `prevAnswer`: 直前に答えられた単語
- `yourTurn`: 自分の番かどうか

ペイロード例:
```js
{
    "type": "onChangeTurn",
    "data": {
        "prevAnswer": "はな",
        "yourTurn", true
    }
}
```

##### `onInput`

ユーザのが文字を入力した際に発生します。自分の入力でも発生します。

- `value`: 入力した値

ペイロード例:
```js
{
    "type": "onInput",
    "data": {
        "value": "ほげ"
    }
}
```

##### `onError`

何らかのエラーが発生した際に送られます。

- `reason`: エラーの内容を説明する文字列

ペイロード例:
```js
{
    "type": "onError",
    "data": {
        "reason": "参加可能な時間ではありません"
    }
}
```

#### クライアント→サーバ

##### `sendAnswer`

答えを送信します。

- `word`: 送信する単語

ペイロード例:
```js
{
    "type": "sendAnswer",
    "data": {
        "word": "はな"
    }
}
```

##### `confirmRetry`

リトライします。

- データは空です。

ペイロード例
```js
{
    "type": "confirmRetry",
    "data": {}
}
```

##### `onInput`

ユーザのが文字を入力した際に発生します。

- `value`: 入力した値

ペイロード例:
```js
{
    "type": "onInput",
    "data": {
        "value": "ほげ"
    }
}
```
