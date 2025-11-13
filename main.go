package main

import (
	"database/sql"
	"fmt"
	"log"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	_ "github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

// -------- MODELS --------

type Event struct {
	ID           string `json:"id"`
	EventName    string `json:"eventName"`
	CustomerName string `json:"customerName"`
	Phone        string `json:"phone"`
	Address      string `json:"address"`
	DataType     string `json:"dataType"`
	CreatedBy    string `json:"createdBy"`
	Paid         int64  `json:"paid"`
	Balance      int64  `json:"balance"`
	TotalCost    int64  `json:"totalCost"`
	Status       string `json:"status"`
	Venue        string `json:"venue"`
	DateTime     string `json:"dateTime"`
	CreatedAt    string `json:"createdAt"`
}

var jwtSecret = []byte("SUPERSECRET_KEY123")

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type User struct {
	ID           int
	Name         string
	Email        string
	PasswordHash string
	Role         string
}

// -------- DB GLOBAL --------

// -------- JWT HELPERS --------

func generateToken(user User) (string, error) {
	claims := jwt.MapClaims{
		"user_id": user.ID,
		"email":   user.Email,
		"role":    user.Role,
		"exp":     time.Now().Add(24 * time.Hour).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

// -------- HANDLERS --------

// POST /api/login
func loginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": "Invalid JSON"})
		return
	}

	row := DB.QueryRow("SELECT id, name, email, password_hash, role FROM users WHERE email=$1", req.Email)

	var user User
	err := row.Scan(&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.Role)

	if err == sql.ErrNoRows {
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}
	if err != nil {
		log.Println("query error:", err)
		c.JSON(500, gin.H{"error": "Server error"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		c.JSON(401, gin.H{"error": "Invalid email or password"})
		return
	}

	token, _ := generateToken(user)

	// 👇 match frontend: it expects access_token + user
	c.JSON(200, gin.H{
		"message":      "Login successful",
		"access_token": token,
		"user": gin.H{
			"id":    user.ID,
			"name":  user.Name,
			"email": user.Email,
			"role":  user.Role,
		},
	})
}

// GET /api/profile (requires auth)
func profileHandler(c *gin.Context) {
	c.JSON(200, gin.H{"user": c.MustGet("user")})
}

// POST /api/events (save event to Postgres)
func createEventHandler(c *gin.Context) {
	var input Event
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	// If ID is empty, generate one
	if input.ID == "" {
		input.ID = fmt.Sprintf("ev_%d", time.Now().UnixNano())
	}

	// Balance
	input.Balance = input.TotalCost - input.Paid

	// If CreatedAt is empty, set now
	if input.CreatedAt == "" {
		input.CreatedAt = time.Now().Format(time.RFC3339)
	}

	_, err := DB.Exec(`
        INSERT INTO events (
            id, event_name, customer_name, phone, address,
            data_type, created_by, paid, balance, total_cost,
            status, venue, date_time, created_at
        )
        VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11,$12,$13,$14)
    `,
		input.ID,
		input.EventName,
		input.CustomerName,
		input.Phone,
		input.Address,
		input.DataType,
		input.CreatedBy,
		input.Paid,
		input.Balance,
		input.TotalCost,
		input.Status,
		input.Venue,
		input.DateTime,
		input.CreatedAt,
	)
	if err != nil {
		log.Println("Insert error:", err)
		c.JSON(500, gin.H{"error": "failed to save event"})
		return
	}

	c.JSON(201, gin.H{"message": "event created", "event": input})
}

// -------- MIDDLEWARE --------

func authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenString := c.GetHeader("Authorization")

		if tokenString == "" {
			c.JSON(401, gin.H{"error": "Missing token"})
			c.Abort()
			return
		}

		claims := jwt.MapClaims{}
		token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return jwtSecret, nil
		})

		if err != nil || !token.Valid {
			c.JSON(401, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		c.Set("user", claims)
		c.Next()
	}
}

func adminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		claims := c.MustGet("user").(jwt.MapClaims)
		role, _ := claims["role"].(string)

		if role != "Admin" {
			c.JSON(403, gin.H{"error": "Admin access only"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// -------- MAIN --------

func main() {
	ConnectDB()
	defer DB.Close()

	r := gin.Default()

	// CORS
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:5500",
			"http://127.0.0.1:5500",
			"http://localhost:2001",
			"http://127.0.0.1:2001",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	api := r.Group("/api")
	{
		api.POST("/login", loginHandler)
		api.GET("/profile", authMiddleware(), profileHandler)

		// Events
		api.POST("/events", authMiddleware(), createEventHandler)
		api.PUT("/events/:id", authMiddleware(), updateEventHandler)
		api.DELETE("/events/:id", authMiddleware(), deleteEventHandler)

		admin := api.Group("/admin", authMiddleware(), adminOnly())
		{
			admin.GET("/test", func(c *gin.Context) {
				c.JSON(200, gin.H{"message": "Admin route working"})
			})
		}
	}

	log.Println("🔥 Server running on :8081")
	r.Run(":8081")
}

// PUT /api/events/:id  -> update existing event
func updateEventHandler(c *gin.Context) {
	id := c.Param("id")

	var input Event
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(400, gin.H{"error": "invalid body"})
		return
	}

	// Make sure we have an ID and keep balance in sync
	if input.ID == "" {
		input.ID = id
	}
	input.Balance = input.TotalCost - input.Paid

	// If CreatedAt is empty, keep the original value (if any)
	if input.CreatedAt == "" {
		var createdAt string
		err := DB.QueryRow("SELECT created_at FROM events WHERE id = $1", id).Scan(&createdAt)
		if err != nil && err != sql.ErrNoRows {
			log.Println("select created_at error:", err)
			c.JSON(500, gin.H{"error": "failed to load existing event"})
			return
		}
		if err == nil {
			input.CreatedAt = createdAt
		} else {
			// if not found, just use now
			input.CreatedAt = time.Now().Format(time.RFC3339)
		}
	}

	res, err := DB.Exec(`
        UPDATE events
        SET event_name=$1,
            customer_name=$2,
            phone=$3,
            address=$4,
            data_type=$5,
            created_by=$6,
            paid=$7,
            balance=$8,
            total_cost=$9,
            status=$10,
            venue=$11,
            date_time=$12,
            created_at=$13
        WHERE id=$14
    `,
		input.EventName,
		input.CustomerName,
		input.Phone,
		input.Address,
		input.DataType,
		input.CreatedBy,
		input.Paid,
		input.Balance,
		input.TotalCost,
		input.Status,
		input.Venue,
		input.DateTime,
		input.CreatedAt,
		input.ID,
	)
	if err != nil {
		log.Println("update error:", err)
		c.JSON(500, gin.H{"error": "failed to update event"})
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(404, gin.H{"error": "event not found"})
		return
	}

	c.JSON(200, gin.H{"message": "event updated", "event": input})
}

// DELETE /api/events/:id -> delete from Postgres
func deleteEventHandler(c *gin.Context) {
	id := c.Param("id")

	res, err := DB.Exec(`DELETE FROM events WHERE id = $1`, id)
	if err != nil {
		log.Println("delete error:", err)
		c.JSON(500, gin.H{"error": "failed to delete event"})
		return
	}

	rows, _ := res.RowsAffected()
	if rows == 0 {
		c.JSON(404, gin.H{"error": "event not found"})
		return
	}

	c.JSON(200, gin.H{"message": "event deleted", "id": id})
}
