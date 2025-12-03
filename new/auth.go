package main

import (
	"context"
	"net/http"
	"strings"

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

func VerifySessionCookieAndCheckRevoked(sessionCookie string) (*auth.Token, error) {
    return firebaseAuth.VerifySessionCookieAndCheckRevoked(context.Background(), sessionCookie)
}

func FirebaseWebsocketAuth() gin.HandlerFunc {
    return func(c *gin.Context) {

        idToken := c.Query("token")
        if idToken == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "トークンがありません"})
            return
        }

        // 2) Firebase で検証
        token, err := VerifyIDToken(idToken)
        if err != nil {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"message": "無効なトークンです"})
            return
        }


        uid := token.UID
        email, _ := token.Claims["email"].(string)
        name, _ := token.Claims["name"].(string)

        c.Set("user", UserInfo{
            ID:   uid,
            Email: email,
            Name:  name,
        })

        c.Next()
    }
}

func FirebaseHttpAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1) Cookie（session）を優先して検証
		if sessionCookie, err := c.Cookie("session"); err == nil && sessionCookie != "" {
			token, err := VerifySessionCookieAndCheckRevoked(sessionCookie)
			if err == nil {
				email, _ := token.Claims["email"].(string)
				name, _ := token.Claims["name"].(string)
				c.Set("user", UserInfo{
					ID:    token.UID,
					Email: email,
					Name:  name,
				})
				c.Next()
				return
			}
			// Cookieが壊れていても、ここではフォールバックを許可する（すぐに401にしない）
			// ログを出したい場合はここで warn を記録
		}

		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
            err := &ErrorResponse{
                "トークンがありません",
                http.StatusUnauthorized,
                "no-token",
            }
            err.response(c)
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
            err := &ErrorResponse{
                "無効のトークンです",
                http.StatusUnauthorized,
                "invalid-token",
            }
            err.response(c)
			return
		}

		idToken := parts[1]

		token, err := VerifyIDToken(idToken)
		if err != nil {
            err := &ErrorResponse{
                "無効または期限切れのトークンです",
                http.StatusUnauthorized,
                "invalid-token",
            }
            err.response(c)
			return
		}

		email, _ := token.Claims["email"].(string)
		name, _ := token.Claims["name"].(string)

        c.Set("user", UserInfo{
            ID:   token.UID,
            Email: email,
            Name:  name,
        })

		c.Next()
	}
}