package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	cache "github.com/go-pkgz/expirable-cache/v2"
	"github.com/kataras/iris/v12"
	"github.com/rkonfj/hooklay/internal"
)

var (
	config         *internal.Config          = internal.NewConfig()
	authenticator  *internal.Authenticator   = internal.NewAuthenticator(config.Security.Token.Value)
	cond           *internal.Conditions      = internal.NewConditions()
	tplMan         *internal.TemplateManager = internal.NewTemplateManager(config.Templates)
	idempotentKeys cache.Cache[string, any]  = cache.NewCache[string, any]().WithMaxKeys(4096).WithTTL(time.Second * 30)
)

func main() {
	app := iris.New()
	app.Post(config.Hook, handle)
	log.Println("Hook ", config.Hook)
	app.Listen(fmt.Sprintf(":%d", config.ServerPort))
}

func handle(ctx iris.Context) {
	body, err := ctx.GetBody()
	if err != nil {
		log.Println("[error] get request body:", err)
		ctx.StatusCode(400)
		ctx.JSON(iris.Map{"code": 400})
		return
	}

	var originalData map[string]any
	err = json.Unmarshal(body, &originalData)
	if err != nil {
		log.Println("[error] parse body:", err)
		return
	}

	// Security Checks
	token1 := ctx.URLParam("token")
	token2 := ctx.GetHeader(config.Security.Token.Header)
	err = authenticator.Authenticate(token1)
	if err != nil {
		err = authenticator.Authenticate(token2)
		if err != nil {
			ctx.StatusCode(401)
			ctx.JSON(iris.Map{"code": 401})
			return
		}
	}

	// Invoke Targets
	for _, target := range config.Targets {
		if target.Enabled {
			err := cond.Meet(originalData, target.Conditions)
			if err != nil {
				continue
			}

			idempotentKey := target.Name + tplMan.Render(target.IdempotentTemplate, originalData).String()
			if _, exists := idempotentKeys.Get(idempotentKey); exists {
				log.Println("Invoke target", target.Name, "(idempotent:", idempotentKey, ")")
				continue
			}

			log.Println("Invoke target", target.Name)
			bodyBuffer := tplMan.Render(target.BodyTemplate, originalData)
			resp, err := http.Post(target.Url, "application/json;charset=utf-8", bodyBuffer)
			if err != nil {
				log.Println("[error] execute:", err)
				continue
			}
			responseBody, err := io.ReadAll(resp.Body)
			if err != nil {
				log.Println("[error] read:", err)
			}
			log.Println("Invoke response", string(responseBody))

			idempotentKeys.Set(idempotentKey, nil, 0)
			log.Println("IdempotentKeys Stats", idempotentKeys.Stat())
		}
	}
	ctx.JSON(iris.Map{"code": 200})
}
