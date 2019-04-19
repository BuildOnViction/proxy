## Build
```
make
```

## Run
```
./proxy --help
Usage of ./proxy:
  -cacheExpiration string
        Cache expiration (default "2s")
  -cacheLimit int
        Cache limit (default 100000)
  -config string
        Path to config file (default "./config/default.json")
  -endpoint value
        List of endpoint urls
  -http string
        Address to listen for HTTP requests on (default "0.0.0.0:3000")
  -n int
        The number of workers to start (default 16)  -verbosity int
        Log Verbosity (default 3)
```

E.g run a proxy:
```
./proxy --endpoint https://rpc.tomochain.com --verbosity 5
```

Run with Docker:
```
docker run -p 3000:3000 tomochain/proxy proxy --endpoint https://rpc.tomochain.com
```

## Test
Start proxy:
```
./proxy
```

Run test 100 requests with proxy:
```
time bash test.sh
```

Run test 100 requests without proxy:
```
time bash test.sh
```
