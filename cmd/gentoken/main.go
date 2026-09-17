package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"tensechassignment/internal/auth"
)

func main() {
	tenant := flag.String("tenant", "tenant1", "tenant id")
	user := flag.String("user", "user1", "user id")
	roles := flag.String("roles", "admin", "comma separated roles")
	secret := flag.String("secret", envOr("JWT_SECRET", "supersecretkey"), "jwt secret")
	issuer := flag.String("issuer", envOr("JWT_ISSUER", "user-management-service"), "jwt issuer")
	audience := flag.String("audience", envOr("JWT_AUDIENCE", "user-management-clients"), "jwt audience")

	flag.Parse()

	roleList := strings.Split(*roles, ",")

	token, err := auth.GenerateToken(
		*secret,
		*issuer,
		*audience,
		*user,
		*tenant,
		roleList,
		24*time.Hour,
	)
	if err != nil {
		fmt.Println("failed to generate token:", err)
		os.Exit(1)
	}

	fmt.Println(token)
}

func envOr(key, fallback string) string {
	value := os.Getenv(key)
	if value != "" {
		return value
	}

	return fallback
}