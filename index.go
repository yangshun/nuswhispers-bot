package main

import (
  "fmt"
  "net/http"
  "net/url"
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
  Id string `json:"confession_id"`
}

type ConfessionData struct {
  Confession Confession `json:"confession"`
}

type GetConfession struct {
  Success bool `json:"success"`
  Data ConfessionData `json:"data"`
}

type ConfessionsData struct {
  ConfessionList []Confession `json:"confessions"`
}

type GetRecentConfessions struct {
  Data ConfessionsData `json:"data"`
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
    fmt.Println("Getting updates for offset: " + strconv.Itoa(lastOffset + 1))
    getUpdatesUrl := config.TelegramBotUrl + "/getUpdates?offset=" + strconv.Itoa(lastOffset + 1)
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

  chatId := update.Message.Chat.Id

  switch command {
  case "/start":
    sendMessage(chatId, "- Get a confession: `/id <confession id>`\n" +
                        "- Recent confessions (max 5): `/recent <count>`\n")
  case "/id":
    if len(messageSegments) != 2 {
      sendMessage(chatId, "Error understanding the command. Please use the structure: `/id <confession id>`")
    } else {
      confessionId := messageSegments[1]
      if len(confessionId) > 0 {
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
          if getConfessionData.Success {
            sendMessage(chatId, getConfessionData.Data.Confession.Content)
          } else {
            sendMessage(chatId, "Oops, confession not found!")
          }
        }
      }
    }
  case "/recent":
    number := "5"
    if len(messageSegments) > 1 {
      number = messageSegments[1]
    }
    response, err := http.Get(config.NUSWhispersAPI + "/confessions/recent?count=" + number)
    if err != nil {
      fmt.Printf("%s", err)
    } else {
      defer response.Body.Close()
      body, err := ioutil.ReadAll(response.Body)
      if err != nil {
        fmt.Printf("%s", err)
      }

      var getRecentConfessionsData = new(GetRecentConfessions)
      error := json.Unmarshal([]byte(body), &getRecentConfessionsData)
      if error != nil {
        fmt.Printf("%s", error)
      }

      message := ""
      for index, confession := range getRecentConfessionsData.Data.ConfessionList {
        if index != 0 {
          message = "\n\n-----\n\n" + message
        }
        const contentLimit = 800
        content := confession.Content
        if len(confession.Content) > contentLimit {
          content = content[:contentLimit] + "..."
        }
        message = "*#" + confession.Id + "*: http://www.nuswhispers.com/confession/" +
                    confession.Id + "\n\n" + content + message

      }
      sendMessage(chatId, message)
    }
  default:
    return
  }
}

func sendMessage(chatId int, message string) {

  urlData := make(url.Values)
  urlData.Set("chat_id", strconv.Itoa(chatId))
  urlData.Set("text", message)
  urlData.Set("parse_mode", "Markdown")
  urlData.Set("disable_web_page_preview", "true")

  sendMessageUrl := config.TelegramBotUrl + "/sendMessage"
  response, err := http.PostForm(sendMessageUrl, urlData)
  fmt.Println("Status: ", response.Status)
  if response.Status != "200 OK" {
    urlData.Set("text", "Sorry the request failed with: " + response.Status)
    http.PostForm(sendMessageUrl, urlData)
  }

  if err != nil {
    fmt.Println(err)
  }
}
