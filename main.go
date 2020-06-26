package main

import (
	"github.com/danil-lashin/twitter-rewards/config"
	"github.com/danil-lashin/twitter-rewards/db"
	"github.com/danil-lashin/twitter-rewards/minter"
	"github.com/danil-lashin/twitter-rewards/push"
	"net/http"
	"time"

	"log"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/gin-gonic/gin"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	gtw "github.com/markbates/goth/providers/twitter"
)

func main() {
	cfg := config.ParseConfig("config.json")
	storage := db.NewDB(cfg)
	minterService := minter.NewService(cfg)
	minterPush := push.NewMinterPush(cfg)

	goth.UseProviders(
		gtw.New(cfg.TwitterKey, cfg.TwitterSecret, cfg.TwitterCallback),
	)
	gothic.GetProviderName = func(req *http.Request) (string, error) {
		return "twitter", nil
	}

	router := gin.Default()
	router.LoadHTMLGlob("templates/*")
	router.GET("/callback", func(c *gin.Context) {
		user, err := gothic.CompleteUserAuth(c.Writer, c.Request)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		createdAt, err := time.Parse(time.RubyDate, user.RawData["created_at"].(string))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if createdAt.After(time.Now().AddDate(-1, 0, 0)) {
			c.JSON(http.StatusBadRequest, gin.H{
				"account": false,
				"rules":   false,
				"error":   "Sorry, but your account is not qualified for this reward",
			})

			return
		}

		oauthCfg := oauth1.NewConfig(cfg.TwitterKey, cfg.TwitterKey)
		token := oauth1.NewToken(user.AccessToken, user.AccessTokenSecret)
		httpClient := oauthCfg.Client(oauth1.NoContext, token)

		client := twitter.NewClient(httpClient)

		tweet, _, err := client.Statuses.Show(cfg.TweetID, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})

			return
		}

		if !tweet.Favorited || !tweet.Retweeted {
			c.JSON(http.StatusBadRequest, gin.H{
				"account": true,
				"rules":   false,
			})

			return
		}

		if storage.IsUserExists(user.UserID) {
			c.JSON(http.StatusBadRequest, gin.H{
				"account": false,
				"rules":   false,
				"error":   "You have already received your reward",
			})

			return
		}

		if err := storage.AddUser(user.UserID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"account": false,
				"rules":   false,
				"error":   "Internal db error",
			})

			return
		}

		address, id := minterPush.Create(cfg.Sender, "@"+user.NickName, cfg.Coin)

		minterService.AddJob(address, func() {
			for {
				balance := minterPush.Activate(id)
				if balance != 0 {
					break
				}

				time.Sleep(1 * time.Second)
			}
		})

		c.JSON(http.StatusOK, gin.H{
			"account": true,
			"rules":   true,
			"address": address,
			"id":      id,
		})
	})

	router.GET("/login", func(c *gin.Context) {
		gothic.BeginAuthHandler(c.Writer, c.Request)
	})

	router.GET("/", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.html", gin.H{
			"tweet_id": cfg.TweetID,
		})
	})

	log.Printf("listening on %s", cfg.ListenAt)
	log.Fatal(http.ListenAndServe(cfg.ListenAt, router))
}
