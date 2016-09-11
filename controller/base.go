package admin 

import (
	"github.com/gin-gonic/gin" 
	"github.com/gin-gonic/contrib/static" // gin middleware for  static content 
	model "github.com/SurgeNews/SurgeServer/model"
	"net/http"
	"fmt"
	"os"
	"io"
	"log"
	"time"
)


var Server *gin.Engine

type SiginResponse struct {
	Success bool
    Token string
}


type LoginCommand struct {
    Username string `json:"username"`
    Password string `json:"password"`
}

func init() {
	Server = gin.Default()

	Server.Use(static.Serve("/static", static.LocalFile("public", false)))

	v1 := Server.Group("api/v1")
	{
		v1.POST("/user/signUp/", signUp)
		v1.POST("/user/signIn/", signIn)
		v1.POST("/audio/upload", uploadAudio)
	}

	Server.GET("/", func(c *gin.Context) {
		c.String(200, model.TypedHello)
	})
}

func signUp(c *gin.Context) {
	var loginCmd LoginCommand
    c.BindJSON(&loginCmd)
    
    token,err := model.AddUser(loginCmd.Username, loginCmd.Password)
    
    if err==nil && token!="" {
    	c.JSON (http.StatusOK, SiginResponse{true, token} )
    	return
    }

    c.JSON(http.StatusInternalServerError, SiginResponse{Success:false})
}

func signIn(c *gin.Context) {
	var loginCmd LoginCommand
    c.BindJSON(&loginCmd)

    fmt.Println(loginCmd.Username, loginCmd.Password)
    token,err := model.AuthoriseUser(loginCmd.Username, loginCmd.Password)

    if err==nil && token!=""{
    	c.JSON (http.StatusOK, SiginResponse{true, token})
    	return
    }    

    c.JSON(http.StatusUnauthorized, SiginResponse{Success:false})
}

func uploadAudio(c *gin.Context) {
	var token  =  c.Request.Header["Xtoken"][0]
	model.ValidateToken(token)
	fmt.Println(c.Request.Header["Xtoken"])
	file, header , err := c.Request.FormFile("filename")
	filename := header.Filename
	fmt.Println(header.Filename)
	out, err := os.Create("tmp/"+filename)
	if err != nil {
	    log.Fatal(err)
	}
	defer out.Close()
	_, err = io.Copy(out, file)
	if err != nil {
	    log.Fatal(err)
	}   

	model.S3Upload("tmp/"+filename)
}

