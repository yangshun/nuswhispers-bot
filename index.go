package main

import (
  "fmt"
  "net/http"
  "io/ioutil"
  "encoding/json"
  "time"
  "strconv"
  "./config"
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

func getUpdates(body []byte) (*GetUpdates, error) {
  var s = new(GetUpdates)
  err := json.Unmarshal(body, &s)
  if err != nil {
    fmt.Printf("%s", err)
  }
  return s, err
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

      getUpdatesData, err := getUpdates([]byte(body))
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
}
