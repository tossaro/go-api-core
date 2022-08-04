# Go API Core [![Go Report Card](https://goreportcard.com/badge/github.com/tossaro/go-api-core)](https://goreportcard.com/report/github.com/tossaro/go-api-core)

Based on common API stack, here is a list of enhanced packages to simplify your go (or 'golang') API development.

## Contents
 - [Getting started](#getting-started)
 - [Enhanced Packages](#enhanced-packages)
 - [Example](https://github.com/tossaro/go-api-core/tree/main/example)

## Getting started

You have an option, you can use core as framework modular, or your own config with manual use of package, or event use directly enhanced package.

1. Download core by using:
```sh
    $ go get -u github.com/tossaro/go-api-core
```

2. Add file `.env` on your `main.go` folder, see [.env](https://github.com/tossaro/go-api-core/blob/main/example/manual/.env)

3. Import the following package:
```go
import core "github.com/tossaro/go-api-core"
```

### Modular Framework [example](https://github.com/tossaro/go-api-core/tree/main/example/modular)

On your `main.go` initialize every package before, and inject to ModuleParams:
```go
//...
captcha := true
privateKeyPath := "./key_private.pem"
publicKeyPath := "./key_public.pem"
core.NewHttp(core.Options{
    EnvPath:        "./.env",
    AuthType:       gin.AuthTypeJwt,
    // if AuthTypeJwt is AuthTypeJwt
    PrivateKeyPath: &privateKeyPath,
    PublicKeyPath:  &publicKeyPath,
    // if AuthType type is AuthTypeGrpc
    // AuthUrl         localhost:50051
    I18n:           bI18n,
    Captcha:        &captcha,
    Modules:        []func(...interface{}){module1.NewHttpV1},
    ModuleParams:   append(make([]interface{}, 0), param1, param2, ...),
})
```

### Core config with manual use of package [example](https://github.com/tossaro/go-api-core/tree/main/example/manual)

1. Initial the config in `main.go` code:
```go
func main() {
    cfg, err := core.NewConfig()
    if err != nil {
        log.Fatal("Config error: %s", err)
    }
    //...
}
```

2. Add every package that you need for your API as example `gin`:
```go
//...
captcha := true
g := gin.New(&gin.Options{
    I18n:     bI18n,
    Mode:     cfg.HTTP.Mode,
    Version:  cfg.App.Version,
    BaseUrl:  cfg.App.Name,
    Log:      log,
    AuthType: gin.AuthTypeJwt,
    // if AuthType type is AuthTypeJwt
    Jwt: jwt,
    // if AuthType type is AuthTypeGrpc
    // AuthService:  &cfg.Services[0].Url,
    Captcha: &captcha,
})
//...
```

## Enhanced Packages
- [Gin](https://github.com/tossaro/go-api-core/blob/main/gin/gin.go)
- [HTTP Server](https://github.com/tossaro/go-api-core/blob/main/httpserver/server.go)
- [JWT RSA](https://github.com/tossaro/go-api-core/blob/main/jwt/jwt.go)
- [Logger](https://github.com/tossaro/go-api-core/blob/main/logger/logger.go)
- [Postgres](https://github.com/tossaro/go-api-core/blob/main/postgres/postgres.go)
- [Redis Cacher](https://github.com/tossaro/go-api-core/blob/main/redis/redis.go)
- [Captcha](https://github.com/tossaro/go-api-core/blob/main/captcha/http.go)

## The MIT License (MIT)

Copyright © `2022` `Hamzah Tossaro`

Permission is hereby granted, free of charge, to any person
obtaining a copy of this software and associated documentation
files (the “Software”), to deal in the Software without
restriction, including without limitation the rights to use,
copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the
Software is furnished to do so, subject to the following
conditions:

The above copyright notice and this permission notice shall be
included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED “AS IS”, WITHOUT WARRANTY OF ANY KIND,
EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES
OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT
HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY,
WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING
FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR
OTHER DEALINGS IN THE SOFTWARE.