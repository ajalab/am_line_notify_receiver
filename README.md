# am\_line_notify_receiver

An [Alertmanager](https://prometheus.io/docs/alerting/alertmanager/) webhook receiver for [LINE Notify](https://notify-bot.line.me/).

## Usage

```bash
$ am_line_notify_receiver <addr> <template> [<lineNotifyToken>]
```

You can provide the OAuth2 token for LINE Notify either via the third command-line argument or via environmental variable `LINE_NOTIFY_TOKEN`.

## Message Template

[`Data`](https://prometheus.io/docs/alerting/notifications/#data) is passed to a message template. See [text/template doc](https://golang.org/pkg/text/template/) for the template syntax.