package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "time"
  "strconv"
  "./config"
  "strings"
)

type Chat struct {
  Id int `json:"id"`
}

type Message struct {
  Id int `json:"message_id"`
  Text string `json:"text"`
  Chat Chat `json:"chat"`
}

type Update struct {
  Id int `json:"update_id"`
  Message Message `json:"message"`
}

type GetUpdates struct {
  Ok bool `json:"ok"`
  UpdateList []Update `json:"result"`
}

type Confession struct {
  Content string `json:"content"`
}

type ConfessionData struct {
  Confession Confession `json:"confession"`
}

type GetConfession struct {
  Success bool `json:"success"`
  Data ConfessionData `json:"data"`
}

func getUpdates(body []byte) (*GetUpdates) {
  var s = new(GetUpdates)
  err := json.Unmarshal(body, &s)
  if err != nil {
    fmt.Printf("%s", err)
  }
  return s
}

func main() {
  ticker := time.NewTicker(time.Millisecond * 1000)

  lastOffset := 0

  for {
    fmt.Println("Getting updates...")
    getUpdatesUrl := config.TelegramBotUrl + "/getUpdates?offset=" + strconv.Itoa(lastOffset + 1)
    fmt.Println(getUpdatesUrl)
    response, err := http.Get(getUpdatesUrl)
    if err != nil {
      fmt.Printf("%s", err)
    } else {
      defer response.Body.Close()
      body, err := ioutil.ReadAll(response.Body)
      if err != nil {
        fmt.Printf("%s", err)
      }

      getUpdatesData := getUpdates([]byte(body))
      updateList := getUpdatesData.UpdateList
      if len(updateList) > 0 {
        lastOffset = updateList[len(updateList) - 1].Id
      }
      for _, update := range updateList {
        processUpdate(update)
      }
    }
    <-ticker.C
  }
}

func processUpdate(update Update) {
  fmt.Println("Update id:", update.Id)
  text := update.Message.Text
  if len(text) == 0 {
    return
  }

  messageSegments := strings.Split(text, " ")
  command := messageSegments[0]

  reply := ""
  switch command {
  case "/start":
    reply = "- Get a confession: `/id <confession id>`%0A" +
            "- Subscribe to confessions: `/subscribe <frequency in hours>`%0A"
  case "/id":
    if len(messageSegments) != 2 {
      reply = "Error understanding the command. Please use the structure: `/id <confession id>`"
    } else {
      confessionId := messageSegments[1]
      if len(confessionId) > 1 {
        response, err := http.Get(config.NUSWhispersAPI + "/confessions/" + confessionId)
        if err != nil {
          fmt.Printf("%s", err)
        } else {
          defer response.Body.Close()
          body, err := ioutil.ReadAll(response.Body)
          if err != nil {
            fmt.Printf("%s", err)
          }

          var getConfessionData = new(GetConfession)
          error := json.Unmarshal([]byte(body), &getConfessionData)
          if error != nil {
            fmt.Printf("%s", error)
          }
          reply = getConfessionData.Data.Confession.Content
          fmt.Println(reply)
        }
      }
    }
  default:
    return
  }
  sendMessageUrl := config.TelegramBotUrl +
                      "/sendMessage?" +
                      "chat_id=" + strconv.Itoa(update.Message.Chat.Id) + "&" +
                      "text=" + reply + "&" +
                      "parse_mode=" + "Markdown"
  http.Get(sendMessageUrl)
}
