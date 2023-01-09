package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"text/template"

	"github.com/kataras/iris/v12"
	"github.com/oliveagle/jsonpath"
)

func main() {
	templateMap := make(map[string]*template.Template)
	app := iris.New()
	for _, relay := range config.Relays {
		if relay.Enabled {
			log.Println("Apply hook " + relay.Hook)
			app.Post(relay.Hook, func(ctx iris.Context) {
				body, err := ctx.GetBody()
				if err != nil {
					log.Fatal(err)
					return
				}

				var originalData map[string]any
				err = json.Unmarshal(body, &originalData)
				if err != nil {
					log.Fatal(err)
					return
				}
				// Security Checks
				token1 := ctx.URLParam("token")
				token2 := ctx.GetHeader(relay.Security.Token.Header)
				log.Println("Token1:" + token1 + " Token2:" + token2)
				if token1 != relay.Security.Token.Value && token2 != relay.Security.Token.Value {
					ctx.StatusCode(401)
					ctx.JSON(iris.Map{"code": 401})
					return
				}
				// Invoke Targets
				for _, target := range relay.Targets {
					if target.Enabled {
						for _, condition := range target.Conditions {
							if strings.EqualFold("eq", condition.Operator) {
								value, err := jsonpath.JsonPathLookup(originalData, condition.Key)
								if err != nil {
									log.Println(err)
									goto next
								}
								if value != condition.Value {
									log.Printf("Condition failed: %s %s %s. Give up target %s\n",
										value, condition.Operator, condition.Value, target.Name)
									goto next
								}
							}
						}
						log.Println("Invoke target " + target.Url)
						tpl := templateMap[target.Name]
						if tpl == nil {
							log.Fatal(err)
							return
						}
						var bodyBuffer bytes.Buffer
						err = tpl.Execute(&bodyBuffer, originalData)
						if err != nil {
							log.Fatal(err)
						}
						resp, err := http.Post(target.Url, "application/json;charset=utf-8", &bodyBuffer)
						if err != nil {
							log.Fatal(err)
						}
						responseBody, err := ioutil.ReadAll(resp.Body)
						if err != nil {
							log.Fatal(err)
						}
						log.Println(string(responseBody))

					}
				next:
				}
				ctx.JSON(iris.Map{"code": 200})
			})
			for _, target := range relay.Targets {
				if target.Enabled {
					log.Println("Register target body template " + target.Name)
					tpl, err := template.New(target.Name).Parse(target.Body)
					if err != nil {
						log.Fatal(err)
						return
					}
					templateMap[target.Name] = tpl
				}
			}
		}
	}

	app.Listen(fmt.Sprintf(":%d", config.ServerPort))
}
