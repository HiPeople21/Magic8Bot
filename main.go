package main

import (
  "os"
  "fmt"
  "strings"
  "net/http"
  "log"
  "flag"
  "math/rand"
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

var (
	GuildID        = flag.String("guild", "", "Test guild ID. If not passed - bot registers commands globally")
    
	RemoveCommands = flag.Bool("rmcmd", true, "Remove all commands after shutdowning or not")
)

var BotID string

var Keys = []string{
    "Affirmative",
    "Non-committal",
    "Negative",
}

var Responses = map[string][]string{
    "Affirmative": {
        "It is certain.",
        "It is decidedly so.",
        "Without a doubt.",
        "Yes definitely.",
        "You may rely on it.",
        "As I see it, yes.",
        "Most likely.",
        "Outlook good.",
        "Yes.",
        "Signs point to yes.",
    },
    "Non-committal": {
        "Reply hazy, try again.",
        "Ask again later.",
        "Better not tell you now.",
        "Cannot predict now.",
        "Concentrate and ask again.",
    },
    "Negative": {
        "Don't count on it.",
        "My reply is no.",
        "My sources say no.",
        "Outlook not so good.",
        "Very doubtful.",
    },
}

var commands = []*discordgo.ApplicationCommand{
  {
    Name: "magic8",
    Description: "Command for getting a magic 8 ball response",
    Options: []*discordgo.ApplicationCommandOption{
      {
        Type: discordgo.ApplicationCommandOptionString,
        Name: "question",
        Description: "question",
        Required: true,
      },
    },
  },
}

var commandHandlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate){
  "magic8": func(s *discordgo.Session, i *discordgo.InteractionCreate) {
      channel := make(chan ResponseMagic)
      go GetResponse(i.ApplicationCommandData().Options[0].StringValue(), channel)
      response := <-channel
      embed := &discordgo.MessageEmbed{
        Title: response.Response.Answer,
        Fields: []*discordgo.MessageEmbedField{
          &discordgo.MessageEmbedField{
            Name: "Question",
            Value: response.Response.Question,
          },
          &discordgo.MessageEmbedField{
            Name: "Answer Type",
            Value: response.Response.Type,
          },
        },
      }
			s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
				Type: discordgo.InteractionResponseChannelMessageWithSource,
				Data: &discordgo.InteractionResponseData{
					Embeds: []*discordgo.MessageEmbed{
            embed,
          },
				},
			})
		},
}

var discord *discordgo.Session

func init() {
  var err error
  token := os.Getenv("token")
	discord, err = discordgo.New("Bot " + token)
	if err != nil {
		log.Fatalf("Invalid bot parameters: %v", err)
	}
  discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := commandHandlers[i.ApplicationCommandData().Name]; ok {
			h(s, i)
		}
	})
}

func main() {

  user, err := discord.User("@me")

  if err != nil {
    fmt.Println(err.Error())
  }

  BotID = user.ID

  discord.AddHandler(func(s *discordgo.Session, r *discordgo.Ready) {
		log.Println("Bot is up!")
	})
  discord.AddHandler(MessageHandler)
  
  err = discord.Open()

  for _, v := range commands {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, "", v)
		if err != nil {
			log.Panicf("Cannot create '%v' command: %v", v.Name, err)
		}
	}

  if err != nil {
    fmt.Println(err.Error())
    return
  }
    
  go KeepAlive()
    
  defer discord.Close()
    
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
        Title: "Question cannot be nothing.",
        Description:"How to use command: `!magic8 (question)`",
        Fields: []*discordgo.MessageEmbedField{
        &discordgo.MessageEmbedField{
          Name: "Asker",
          Value: message.Author.Mention(),
        },
        },
      }
    
      session.ChannelMessageSendEmbed(message.ChannelID, embed)
      return
    }
    channel := make(chan ResponseMagic)
    go GetResponse(question, channel)
    response := <-channel
    embed := &discordgo.MessageEmbed{
      Title: response.Response.Answer,
      Fields: []*discordgo.MessageEmbedField{
        &discordgo.MessageEmbedField{
          Name: "Question",
          Value: response.Response.Question,
        },
        &discordgo.MessageEmbedField{
          Name: "Answer Type",
          Value: response.Response.Type,
        },
        &discordgo.MessageEmbedField{
          Name: "Asker",
          Value: message.Author.Mention(),
        },
      },
    }

    session.ChannelMessageSendEmbed(message.ChannelID, embed)
  }
}


func GetResponse(question string, channel chan ResponseMagic) {
  
  type_ := Keys[rand.Intn(len(Keys))]

  answer := Responses[type_][rand.Intn(len(Responses[type_]))]
  responseObject := ResponseMagic{
      Response: Response{
          Question: question,
          Answer: answer,
          Type: type_,
      },
  }
  channel<-responseObject
}

func KeepAlive(){
  http.HandleFunc("/", IndexHandler)
  http.ListenAndServe(":8000", nil)
}

func IndexHandler(w http.ResponseWriter, r *http.Request){
  fmt.Fprintf(w, "<a href='https://www.youtube.com/watch?v=dQw4w9WgXcQ'>Click here</a>")
}