package main

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"NetUtil/jwtutil"
)

func main() {
	// 初始化 JWT 管理器
	jwtManager := jwtutil.NewJWTManager(
		"mySecretKey",   // 密钥
		time.Minute*10,  // token 10 分钟过期
		"session_token", // cookie 名称
	)

	// 登录：生成 token 放入 cookie
	http.HandleFunc("/login", func(w http.ResponseWriter, r *http.Request) {
		claims := map[string]interface{}{
			"user_id": 123,
			"role":    "admin",
		}

		token, err := jwtManager.GenerateToken(claims)
		if err != nil {
			http.Error(w, "failed to generate token", http.StatusInternalServerError)
			return
		}

		jwtManager.SetTokenCookie(w, token)
		fmt.Fprintf(w, "Login OK. Token is set in cookie.")
	})

	// 读取用户信息：从 cookie → token → claims
	http.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		tokenStr, err := jwtManager.ReadTokenFromCookie(r)
		if err != nil {
			http.Error(w, "no token found", http.StatusUnauthorized)
			return
		}

		claims, err := jwtManager.ParseToken(tokenStr)
		if err != nil {
			http.Error(w, "invalid token", http.StatusUnauthorized)
			return
		}

		fmt.Fprintf(w, "User info: %v\n", claims)
	})

	log.Println("Server running at http://localhost:8080 ...")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
