#

## build and run

```bash
go build -o main -v .
./main -path config/ -rpc http://127.0.0.1:9091/transmission/rpc -host :9093
```

immidiately run once

```bash
curl http://127.0.0.1:9093/start_job
```

## config

```json
{
    "rss": [
        {
            "name": "rss1",
            "url": "https://example.com/RSS1",
            "download_dir": "/download/rss1",
            "regexp": [
                "\\(CR"
            ],
            "exclude_regexp":  [
                "\\(Baha"
            ]
        },
        {
            "name": "rss2",
            "url": "https://example.com/RSS2",
            "download_dir": "/download/rss2",
            "regexp": [
                "\\(CR,RSS2",
                "RSS2"
            ],
            "exclude_regexp":  [
                "\\(Baha"
            ]
        }
    ]
}
```