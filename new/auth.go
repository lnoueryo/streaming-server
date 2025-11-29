package main

import (
	"context"
	"net/http"

	"firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"google.golang.org/api/option"
)

type UserInfo struct {
    ID    string
    Email string
    Name  string
}

var firebaseApp *firebase.App
var firebaseAuth *auth.Client

func InitFirebase() {
    opt := option.WithCredentialsFile(".credentials/firebase-admin.json")
    config := &firebase.Config{ProjectID: "streaming-cb408"}

    app, err := firebase.NewApp(context.Background(), config, opt)
    if err != nil {
        log.Fatalf("Failed to init Firebase: %v", err)
    }
    firebaseApp = app

    authClient, err := app.Auth(context.Background())
    if err != nil {
        log.Fatalf("Failed to init Firebase Auth: %v", err)
    }
    firebaseAuth = authClient
}

func VerifyIDToken(idToken string) (*auth.Token, error) {
    return firebaseAuth.VerifyIDToken(context.Background(), idToken)
}

func FirebaseAuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {

        idToken := c.Query("token")
        if idToken == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "token required"})
            return
        }

        // 2) Firebase で検証
        token, err := VerifyIDToken(idToken)
        if err != nil {
            log.Error("invalid token: ", err)
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid token"})
            return
        }


        uid := token.UID
        email, _ := token.Claims["email"].(string)
        name, _ := token.Claims["name"].(string)

        log.Info("Authenticated UID:", uid)

        c.Set("user", UserInfo{
            ID:   uid,
            Email: email,
            Name:  name,
        })

        c.Next()
    }
}