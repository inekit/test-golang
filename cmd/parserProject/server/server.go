package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"regexp"

	"github.com/gin-gonic/gin"
	_ "github.com/lib/pq"
)

// Отправляем json с данными из БД
func getByStruct(c *gin.Context) {

	// Получаем данные в виде структуры
	parseResult, err := ConnectDB().Struct()

	// Переводим структуру в json файл
	jsonRes, err := json.Marshal(parseResult)

	if err != nil {
		SendErrorJSON(c, err, true)
		return
	}

	c.Data(http.StatusOK, "application/json", jsonRes)
}

// Обработка запросов на парсинг XML по ссылке
func parseByLink(mode string) gin.HandlerFunc {
	fn := func(c *gin.Context) {
		{
			// Проверяем, что введен url
			url := c.PostForm("link")
			if !ValidateURL(url) {
				SendErrorJSON(c, errors.New("invalid link"), false)
				return
			}
			// Запрашиваем xml-файл
			resp, err := http.Get(url)
			if resp.StatusCode != 200 || err != nil {
				//sendErrorJSON(c, errors.New("status:"+string(resp.StatusCode)), true)
				return
			}

			// Читаем в переменную
			body, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				SendErrorJSON(c, err, true)
				return
			}

			if mode == "struct" {
				parseResult, err := internals.ParseToStruct(body)

				// Переводим структуру в json файл
				jsonRes, err := json.Marshal(parseResult)
				if err != nil {
					SendErrorJSON(c, err, true)
					return
				}

				c.Data(http.StatusOK, "application/json", jsonRes)
			}
			if mode == "db" {
				// Переводим структуру в БД
				if internals.ParseToDB(body) != nil {
					SendErrorJSON(c, errors.New("cant send to db"), true)
				} else {
					c.Data(http.StatusOK, "application/json", []byte("{'status':'sended to db'}"))
				}
			}

		}
	}

	return gin.HandlerFunc(fn)
}

func ValidateURL(url string) bool {
	validID := regexp.MustCompile(`^(https?:\/\/)([a-zA-Z0-9.\/\&:])*$`)
	if !validID.MatchString(url) {
		return false
	}
	return true
}

func SendErrorJSON(c *gin.Context, err error, cLog bool) {
	c.AbortWithStatusJSON(200, map[string]interface{}{"error": err})
	if cLog {
		fmt.Println(err)
	}
}
