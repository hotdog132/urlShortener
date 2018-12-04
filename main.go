package main

import (
	"encoding/json"
	"io/ioutil"
	"net/url"
	"os"
	"path/filepath"

	// "regexp"
	"bytes"
	"errors"
	"fmt"
	"net/http"

	// "os"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	// "github.com/gin-contrib/logger"
	// "github.com/rs/zerolog"
	// "github.com/rs/zerolog/log"
)


var (
	globalS             *mgo.Session
	recaptchaPublicKey  string
	recaptchaPrivateKey string
	hostName string
	dialURL string
	dbName string
	dbCollection string

	// rxURL   = regexp.MustCompile(`^/regexp\d*`)
)

// ShortenURL shortenURL
type ShortenURL struct {
	ID          bson.ObjectId `bson:"_id,omitempty"`
	OriginalURL string        `bson:"url" json:"url"`
	EncodeURL   string        `bson:"encodeurl" json:"encodeurl"`
}

type RecaptchaResp struct {
	Success bool    `json:"success"`
	Score   float32 `json:"score"`
}

func ping(c *gin.Context) {
	c.JSON(200, gin.H{
		"message": "pong",
	})
}

// func initLog() {
// 	zerolog.SetGlobalLevel(zerolog.InfoLevel)
// 	if gin.IsDebugging() {
// 		zerolog.SetGlobalLevel(zerolog.DebugLevel)
// 	}

// 	log.Logger = log.Output(
// 		zerolog.ConsoleWriter{
// 			Out:     os.Stderr,
// 			NoColor: false,
// 		},
// 	)
// }

func home(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"recaptchaPublicKey": recaptchaPublicKey,
	})
}

func redirectURL(c *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			c.JSON(200, gin.H{
				"error": r,
			})
		}
	}()
	encodeString := c.Param("encode")
	if len(encodeString) == 0 {
		c.JSON(200, gin.H{
			"error": "get encode string nil",
		})
		return
	}
	var shortenURL *ShortenURL
	collection := globalS.DB(dbName).C(dbCollection)
	err := collection.FindId(bson.ObjectIdHex(encodeString)).One(&shortenURL)
	if err != nil {
		c.JSON(200, gin.H{
			"message": "error during find objectid from db",
			"error":   err,
			"encode":  encodeString,
		})
		return
	}
	if shortenURL == nil {
		c.JSON(200, gin.H{
			"message": "Cannot find url",
		})
		return
	}
	// c.JSON(200, gin.H{
	// 	"url": shortenURL.OriginalURL,
	// })
	fmt.Println("originalURL", shortenURL.OriginalURL)
	c.Redirect(http.StatusMovedPermanently, shortenURL.OriginalURL)
	return
}

// isValidURL tests a string to determine if it is a url or not.
func isValidURL(toTest string) bool {
	_, err := url.ParseRequestURI(toTest)
	if err != nil {
		return false
	} else {
		return true
	}
}

func verifyToken(token string) (bool, error) {
	jsonStr := fmt.Sprintf(`secret=%s&response=%s`, recaptchaPrivateKey, token)
	fmt.Println("jsonStr", jsonStr)
	// fmt.Println("jsonStr:", jsonStr)
	jsonByte := []byte(jsonStr)
	url := "https://www.google.com/recaptcha/api/siteverify"
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonByte))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded; charset=utf-8")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()
	fmt.Println("response Status:", resp.Status)
	if resp.StatusCode != http.StatusOK {
		return false, errors.New("wrong response status")
	}
	// fmt.Println("response Headers:", resp.Header)
	body, _ := ioutil.ReadAll(resp.Body)
	fmt.Println(string(body))
	// if err != nil {
	// 	return false, err
	// }
	recaptchaResp := &RecaptchaResp{}
	// err = json.NewDecoder(resp.Body).Decode(recaptchaResp)
	// if err != nil {
	// 	fmt.Println(err)
	// 	return false, err
	// }
	err = json.Unmarshal(body, recaptchaResp)
	if err != nil {
		return false, err
	}
	// fmt.Println(recaptchaResp.Success, recaptchaResp.Score)
	// fmt.Println("response Body success:", string(body))
	if recaptchaResp.Success && recaptchaResp.Score >= 0.5 {
		return true, nil
	}
	return false, nil
}

func urlRequest(c *gin.Context) {
	token := c.PostForm("token")
	if len(token) == 0 {
		c.JSON(200, gin.H{
			"message": "token error",
		})
		return
	}

	isVerified, err := verifyToken(token)
	if err != nil || isVerified == false {
		c.JSON(200, gin.H{
			"message": "verified token error",
			"error":   err,
		})
		return
	}

	var shortenURL ShortenURL
	err = c.Bind(&shortenURL)
	if err != nil || len(shortenURL.OriginalURL) == 0 {
		c.JSON(200, gin.H{
			"message": "parse json error",
			"error":   err,
		})
		return
	}

	if !isValidURL(shortenURL.OriginalURL) {
		c.JSON(200, gin.H{
			"message": "wrong url format",
		})
		return
	}

	collection := globalS.DB(dbName).C(dbCollection)
	err = collection.Find(bson.M{"url": shortenURL.OriginalURL}).One(&shortenURL)
	if err == nil {
		c.JSON(200, gin.H{
			"encodeurl": shortenURL.ID,
		})
		return
	}
	objectID := bson.NewObjectId()
	err = collection.Insert(bson.M{"_id": objectID, "url": shortenURL.OriginalURL})
	if err != nil {
		c.JSON(200, gin.H{
			"message": "insert db error",
			"error":   err,
		})
		return
	}
	c.JSON(http.StatusOK, gin.H{"encodeurl": hostName + "/v/" + fmt.Sprint(objectID.Hex())})
}

func ginEngine() *gin.Engine {
	r := gin.Default()
	// r.Use(logger.SetLogger())
	// // Custom logger
	// subLog := zerolog.New(os.Stdout).With().
	// 	Str("foo", "bar").
	// 	Logger()

	// r.Use(logger.SetLogger(logger.Config{
	// 	Logger:         &subLog,
	// 	UTC:            true,
	// 	SkipPath:       []string{"/skip"},
	// 	SkipPathRegexp: rxURL,
	// }))
	r.LoadHTMLFiles("tpl/index.html")
	r.GET("/", home)
	r.GET("/v/:encode", redirectURL)
	// r.GET("/ping", ping)
	r.POST("/shortenUrl", urlRequest)
	return r
}

func getMongoDBInstance() *mgo.Session {
	session, err := mgo.Dial(dialURL)
	if err != nil {
		panic(err)
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)
	return session
}

func initArgs() {
	if len(os.Args) != 3 {
		fmt.Printf("usage: %s <reCaptcha public key> <reCaptcha private key>\n", filepath.Base(os.Args[0]))
		os.Exit(1)
	}
	recaptchaPublicKey = os.Args[1]
	recaptchaPrivateKey = os.Args[2]
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil { // Handle errors reading the config file
		panic(fmt.Errorf("Fatal error config file: %s", err))
	}

	hostName = viper.GetString("hostname")
	dialURL = viper.GetString("mongodb.dialurl")
	dbName = viper.GetString("mongodb.dbname")
	dbCollection = viper.GetString("mongodb.collection")
}

func main() {
	initArgs()
	loadConfig()
	// initLog()
	globalS = getMongoDBInstance()
	defer func() {
		if globalS != nil {
			globalS.Close()
		}
	}()
	collection := globalS.DB(dbName).C(dbCollection)
	collection.RemoveAll(nil)
	ginEngine().Run() // listen and serve on 0.0.0.0:8080
}
