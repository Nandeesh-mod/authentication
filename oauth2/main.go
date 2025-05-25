package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
	"github.com/markbates/goth/providers/google"
)

func main() {
	r := gin.Default()

	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error ", err)
	}

	client_ID := os.Getenv("CLIENT_ID")
	client_Secret := os.Getenv("CLIENT_SECRET")
	clientCallbackURL := os.Getenv("CLIENT_CALLBACK_URL")

	if client_ID == "" || client_Secret == "" || clientCallbackURL == "" {
		log.Fatal("Environment variables (CLIENT_ID, CLIENT_SECRET, CLIENT_CALLBACK_URL) are required")
	}

	goth.UseProviders(google.New(client_ID, client_Secret, clientCallbackURL))

	r.GET("/", home)
	r.GET("/auth/:provider", signInWithProvider)
	r.GET("/auth/:provider/callback", callbackHandler)
	r.GET("/success", Success)

	r.Run(":5005")

}

func home(g *gin.Context) {
	temp, err := template.ParseFiles("template/index.html")
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{
			"error":       "Failed to parse",
			"status_code": http.StatusBadRequest,
		})
		return
	}
	err = temp.Execute(g.Writer, gin.H{})
	if err != nil {
		g.JSON(http.StatusBadRequest, gin.H{
			"error":       "Failed to parse",
			"status_code": http.StatusBadRequest,
		})
		return
	}
}

func signInWithProvider(g *gin.Context) {
	provider := g.Param("provider")
	q := g.Request.URL.Query()
	q.Add("provider", provider)
	g.Request.URL.RawQuery = q.Encode()

	gothic.BeginAuthHandler(g.Writer, g.Request)
}

func callbackHandler(g *gin.Context) {
	provider := g.Param("provider")
	q := g.Request.URL.Query()
	q.Add("provider", provider)
	g.Request.URL.RawQuery = q.Encode()

	user, err := gothic.CompleteUserAuth(g.Writer, g.Request)
	if err != nil {
		g.AbortWithError(http.StatusInternalServerError, err)
		return
	}
	fmt.Println("***********************************")
	fmt.Println(user)
	fmt.Println("***********************************")
	g.Redirect(http.StatusTemporaryRedirect, "/success")

}

func Success(g *gin.Context) {

	g.Data(http.StatusOK, "text/html; charset=utf-8", []byte(fmt.Sprintf(`
      <div style="
          background-color: #fff;
          padding: 40px;
          border-radius: 8px;
          box-shadow: 0 2px 4px rgba(0, 0, 0, 0.1);
          text-align: center;
      ">
          <h1 style="
              color: #333;
              margin-bottom: 20px;
          ">You have Successfull signed in!</h1>
          
          </div>
      </div>
  `)))
}
