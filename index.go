package main

import (
  "fmt"
  "os"
  "net/http"
  "io/ioutil"
  "encoding/json"
)

type config struct {
  TelegramBotAPI string
  TelegramBotToken string
}

func main() {
  file, e := ioutil.ReadFile("./config.json")
  if e != nil {
    fmt.Printf("File error: %v\n", e)
    os.Exit(1)
  }
  fmt.Printf("%s\n", string(file))

  var appConfig config
  json.Unmarshal(file, &appConfig)

  response, err := http.Get(appConfig.TelegramBotAPI + appConfig.TelegramBotToken + "/getUpdates")
  if err != nil {
    fmt.Printf("%s", err)
  } else {
    defer response.Body.Close()
    contents, err := ioutil.ReadAll(response.Body)
    if err != nil {
      fmt.Printf("%s", err)
    }
    fmt.Printf("%s\n", string(contents))
  }
}
