# API Core

Based on common API stack, here is a list of enhanced packages to simplify your go (or 'golang') API development.

## Contents
 - [Getting started](#getting-started)
 - [Packages](#packages)
 - [Example](https://github.com/tossaro/go-api-core/example)

## Getting started

1. Download core by using:
```sh
    $ go get -u github.com/tossaro/go-api-core
```

2. Add file `.env` on your `main.go` folder, see [.env](https://github.com/tossaro/go-api-core/example/.env)

3. Import the following package:
```go
    import "github.com/tossaro/go-api-core"
```

4. Initial the config in `main.go` code:
```go
    func main() {
        cfg, err := core.NewConfig()
        if err != nil {
            log.Fatal("Config error: %s", err)
        }
        //...
    }
```

5. Add every package that you need for your API as example `gin`:
```go
    g := gin.New(&gin.Options{
        Mode:         cfg.HTTP.Mode,
        Version:      cfg.App.Version,
        BaseUrl:      cfg.App.Name,
        Logger:       l,
        Redis:        rdb,
        AccessToken:  cfg.TOKEN.Access,
        RefreshToken: cfg.TOKEN.Refresh,
    })
```

## Packages
- [gin](https://github.com/tossaro/go-api-core/gin)
- [httpserver](https://github.com/tossaro/go-api-core/httpserver)
- [jwt](https://github.com/tossaro/go-api-core/jwt)
- [logger](https://github.com/tossaro/go-api-core/logger)
- [postgres](https://github.com/tossaro/go-api-core/postgres)
- [twilio](https://github.com/tossaro/go-api-core/twilio)
- [captcha](https://github.com/tossaro/go-api-core/captcha)