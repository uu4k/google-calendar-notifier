# Google Calendar Notifier

## 前準備

1. 以下のURLを参考にGoogle Calendar APIを有効にしてclient_secret.jsonファイルをダウンロードし、実行ファイルと同じディレクトリに配置する
    - [Go Quickstart  |  Calendar API  |  Google Developers](https://developers.google.com/calendar/quickstart/go#step_1_turn_on_the_api_name)
2. 実行ファイルを実行してOauth認証のURLを取得して取得したトークンIDを貼り付ける

```sh
$ ./google-calendar-notifier -i 1
Go to the following link in your browser then type the authorization code:
https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=xxx
# 上のURLからトークン取得して貼り付け->enter->"Saving credential file to: token.json"って出ればOK
```
