package main

import (
  "os"
  "fmt"
  "strings"
  "io/ioutil"
  "encoding/json"
  "net/http"
  "log"

  "github.com/bwmarrin/discordgo"
)

type ResponseMagic struct {
  Response Response `json:"magic"`
}

type Response struct {
  Question string `json:"question"`
  Answer string `json:"answer"`
  Type string `json:"type"`
}

var BotID string

func main() {
  token := os.Getenv("token")
  discord, err := discordgo.New("Bot " + token)
  
  if err != nil {
    fmt.Println(err.Error())
    return
  }

  user, err := discord.User("@me")

  if err != nil {
    fmt.Println(err.Error())
  }

  BotID = user.ID

  discord.AddHandler(MessageHandler)

  err = discord.Open()

  if err != nil {
    fmt.Println(err.Error())
    return
  }

  fmt.Println("Bot is running!")

  <- make(chan struct{})
  return
}


func MessageHandler(session *discordgo.Session, message *discordgo.MessageCreate) {

  if message.Author.ID == BotID {
    return
  }
  if strings.HasPrefix(message.Content, "!magic8"){
    question := strings.TrimSpace(strings.TrimPrefix(message.Content, "!magic8"))
    if question == ""{
      embed := &discordgo.MessageEmbed{
        Title: message.Author.Username +", question cannot be nothing.",
        Description:"How to use command: '!magic8 (question)'",
        }
      session.ChannelMessageSendEmbed(message.ChannelID, embed)
      return
    }

    response := GetResponse(question)
    embed := &discordgo.MessageEmbed{
      Title: response.Response.Answer,
      Fields: []*discordgo.MessageEmbedField{
        &discordgo.MessageEmbedField{
          Name: "Question",
          Value: response.Response.Question,
        },
        &discordgo.MessageEmbedField{
          Name: "Type",
          Value: response.Response.Type,
        },
      }
    }

    session.ChannelMessageSendEmbed(message.ChannelID, embed)
  }
}


func GetResponse(question string) ResponseMagic {
  res, err := http.Get("https://8ball.delegator.com/magic/JSON/" + question)
  if err != nil {
    return ResponseMagic{
      Response: Response{
        Question: question,
        Answer: "Error. Please try again.",
        Type: "N/A",
      },
    }
  }

  responseData, err := ioutil.ReadAll(res.Body)
  if err != nil {
    return ResponseMagic{
      Response: Response{
        Question: question,
        Answer: "Error. Please try again.",
        Type: "N/A",
      },
    }
  }

  var responseObject ResponseMagic
  jsonErr := json.Unmarshal(responseData, &responseObject)
  if jsonErr != nil {
		log.Fatal(jsonErr)
	}
  return responseObject
}