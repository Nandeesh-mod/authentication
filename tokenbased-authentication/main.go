package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-jwt/jwt/v5"
)

type User struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AppError struct {
	Message   string `json:"message"`
	ErrorCode int    `json:"error_code"`
}

type Student struct {
	Id       int      `json:"id"`
	Email    string   `json:"email"`
	Name     string   `json:"name"`
	Place    string   `json:"place"`
	Age      int      `json:"age"`
	Semister int      `json:"int"`
	Courses  []string `json:"courses"`
}

// server need a jwt secret
var jwtSecret = []byte("superprotectedkey")

type CustomClaims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

var users = map[string]string{
	"admin": "adminpass",
	"user1": "userpass",
}

type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

var students = []Student{
	{
		Id:       1,
		Email:    "nandeeshm2020@gmail.com",
		Name:     "Nandeesh M",
		Place:    "Shivamogga",
		Age:      21,
		Semister: 4,
		Courses:  []string{"Computer Networks", "Operating Systems", "Database Systems", "DSA"},
	},
	{
		Id:       2,
		Email:    "rahul.p@example.com",
		Name:     "Rahul Patil",
		Place:    "Bengaluru",
		Age:      22,
		Semister: 5,
		Courses:  []string{"Software Engineering", "Machine Learning", "Compiler Design", "Web Technologies"},
	},
	{
		Id:       3,
		Email:    "ananya.k@example.com",
		Name:     "Ananya Kulkarni",
		Place:    "Mysuru",
		Age:      20,
		Semister: 3,
		Courses:  []string{"Discrete Mathematics", "Data Structures", "Python Programming", "Computer Architecture"},
	},
	{
		Id:       4,
		Email:    "vikram.r@example.com",
		Name:     "Vikram Reddy",
		Place:    "Hubballi",
		Age:      23,
		Semister: 6,
		Courses:  []string{"Cloud Computing", "Big Data", "Cybersecurity", "Agile Methodologies"},
	},
	{
		Id:       5,
		Email:    "sneha.s@example.com",
		Name:     "Sneha Shetty",
		Place:    "Mangaluru",
		Age:      19,
		Semister: 2,
		Courses:  []string{"Introduction to Programming", "Mathematics-I", "English Communication", "Environmental Studies"},
	},
}

func GetUsers(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	encoder := json.NewEncoder(w)
	err := encoder.Encode(students)
	if err != nil {
		fmt.Println("Error")
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
}

func WriteError(w http.ResponseWriter, errmsg string, code int) {
	error := AppError{
		Message:   errmsg,
		ErrorCode: code,
	}
	encoder := json.NewEncoder(w)
	encoder.Encode(error)
}

func BasicAuthFailed(w http.ResponseWriter) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("WWW-Authentication", "Failed")
	w.WriteHeader(http.StatusUnauthorized)
	WriteError(w, "Unauthorized", http.StatusUnauthorized)
}

func AuthMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if authHeader == "" {
			http.Error(w, "Authorization header required", http.StatusUnauthorized)
			return
		}

		// Check if the header starts with "Bearer "
		if !strings.HasPrefix(authHeader, "Bearer ") {
			http.Error(w, "Authorization header must be in Bearer format", http.StatusUnauthorized)
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		// Parse and validate the token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validate the signing method
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return jwtSecret, nil // Return the secret key for verification
		})

		if err != nil {
			if errors.Is(err, jwt.ErrTokenMalformed) {
				http.Error(w, "Malformed token", http.StatusUnauthorized)
				return
			} else if errors.Is(err, jwt.ErrTokenExpired) {
				http.Error(w, "Token is expired", http.StatusUnauthorized)
				return
			} else if errors.Is(err, jwt.ErrTokenNotValidYet) {
				http.Error(w, "Token not yet active", http.StatusUnauthorized)
				return
			} else {
				http.Error(w, fmt.Sprintf("Couldn't handle token: %v", err), http.StatusUnauthorized)
				return
			}
		}

		// Check if the token is valid
		if !token.Valid {
			http.Error(w, "Invalid token", http.StatusUnauthorized)
			return
		}

		// Extract claims and pass them to the request context
		claims, ok := token.Claims.(*CustomClaims)
		if !ok {
			http.Error(w, "Invalid token claims", http.StatusUnauthorized)
			return
		}

		// Store claims in the request context for subsequent handlers to access
		ctx := context.WithValue(r.Context(), "claims", claims)
		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid Requests Body", http.StatusBadRequest)
		return
	}

	expectedPassword, ok := users[req.Username]
	if !ok || expectedPassword != req.Password {
		WriteError(w, "Invalid Credential", http.StatusUnauthorized)
		return
	}

	role := "user"
	if req.Username == "admin" {
		role = "admin"
	}

	claims := CustomClaims{
		Username: req.Username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   req.Username,
		},
	}

	// create a new token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(jwtSecret)
	if err != nil {
		WriteError(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// func AdminHandler(w http.ResponseWriter, r *http.Request) {

// }

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)

	r.Post("/login", LoginHandler)

	r.Group(func(authRouter chi.Router) {
		authRouter.Use(AuthMiddleware)

		authRouter.Get("/users", GetUsers)

	})

	if err := http.ListenAndServe(":8080", r); err != nil {
		log.Fatal("Error: ", err)
	}

}
