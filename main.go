package main

import (
	"encoding/json"
	"fmt"
	"os"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	"github.com/line/line-bot-sdk-go/linebot"
)

const (
	OpenDataURL string = "http://opendata2.epa.gov.tw/AQI.json"
)

type AQI struct {
	SiteName    string `json:"SiteName"`
	County      string `json:"County"`
	AQI         string `json:"AQI"`
	Pollutant   string `json:"Pollutant"`
	Status      string `json:"Status"`
	SO2         string `json:"SO2"`
	CO          string `json:"CO"`
	CO_8hr      string `json:"CO_8hr"`
	O3          string `json:"O3"`
	O3_8hr      string `json:"O3_8hr"`
	PM10        string `json:"PM10"`
	PM25        string `json:"PM2.5"`
	NO2         string `json:"NO2"`
	NOx         string `json:"NOx"`
	NO          string `json:"NO"`
	WindSpeed   string `json:"WindSpeed"`
	WindDirec   string `json:"WindDirec"`
	PublishTime string `json:"PublishTime"`
	PM25_AVG    string `json:"PM2.5_AVG"`
	PM10_AVG    string `json:"PM10_AVG"`
	Latitude    string `json:"Latitude"`
	Longitude   string `json:"Longitude"`
}

var bot *linebot.Client
var AQIData []AQI
var AQIStatus string

func getDtypeReply(dtype string, site_name string) string {
	res, err := http.Get(OpenDataURL)
	if err != nil {
		log.Println("Get json data error")
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Println("Read json data error")
	}

	err = json.Unmarshal([]byte(body), &AQIData)
	if err != nil {
		log.Println("Parse json data error")
	}

	len := len(AQIData)
	for i := 0; i < len; i++ {
		if AQIData[i].SiteName == site_name {
			if dtype == "AQI" {
				return AQIData[i].SiteName + "的" + " AQI(空氣品質指標) 數值為 " + AQIData[i].AQI
		    }else if dtype == "PM2.5" {
			    return AQIData[i].SiteName + "的" + " PM2.5 數值為 " + AQIData[i].PM25
			}
		}
	}

	return "無此地區的空氣品質資訊"
}

func parseMessage(message string) (bool, string, string) {
	words := strings.Fields(message)
	words_len := len(words)

	if words_len == 1 {
		return true, "AQI", words[0]
	}else if words_len == 2 && words[1] == "PM2.5" {
		return true, "PM2.5", words[0]
    }
	return false, "", ""
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	events, err := bot.ParseRequest(r)

	if err != nil {
		if err == linebot.ErrInvalidSignature {
			w.WriteHeader(400)
		} else {
			w.WriteHeader(500)
		}
		return
	}

	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				status, dtype, site_name := parseMessage(message.Text)
				fmt.Println("status, dtype, site_name: ", status, dtype, site_name)
				log.Println("site_name: " + site_name)
				
				var reply string
				if !status {
					reply = "輸入格式錯誤，AQI格式需為: 地區; PM2.5格式需為: 地區 PM2.5"
					log.Println(reply)
				} else {
					reply = getDtypeReply(dtype, site_name)
					log.Println(reply)
				}

				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(message.ID+": "+reply), linebot.NewTextMessage(AQIStatus)).Do(); err != nil {
					log.Print(err)
				}
			case *linebot.StickerMessage:
				if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage("測試測試"), linebot.NewStickerMessage("1", "1")).Do(); err != nil {
					log.Print(err)
				}
			}
		}
	}
}

func main() {
	var err error
	bot, err = linebot.New(os.Getenv("ChannelSecret"), os.Getenv("ChannelAccessToken"))
	log.Println("Bot:", bot, " err:", err)

	status, _ := ioutil.ReadFile("aqi-status")
	AQIStatus = string(status)

	http.HandleFunc("/callback", callbackHandler)
	port := "8000"
	addr := fmt.Sprintf(":%s", port)
	http.ListenAndServe(addr, nil)
}
