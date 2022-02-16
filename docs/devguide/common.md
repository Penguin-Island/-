# 共通

開発環境ではデータベースなどは Docker を使って準備します。

データベースを立ち上げる時と終了させる時のコマンドはそれぞれ以下です。

データベースを立ち上げる時:

```shell
$ docker compose up -d
```

データベースを終了させる時:

```shell
$ docker compose down
```

現時点では `docker compose down` すると保存した内容が全て失われます。
注意してください。
