package eventdb

import (
	"encoding/base64"
	"encoding/json"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

// Server is an HTTP server that provides an API for adding and querying events.
type Server struct {
	store          *Store
	authorizer     *Authorizer
	logPath        string
	allowedOrigins []string
}

// NewServer creates a new server with the given authorizer, store, allowed origins, and log path.
func NewServer(authorizer *Authorizer, store *Store, allowedOrigins []string, logPath string) *Server {
	return &Server{
		authorizer:     authorizer,
		store:          store,
		logPath:        logPath,
		allowedOrigins: allowedOrigins,
	}
}

// getUserFromAuthorizationHeader extracts the username from an HTTP Authorization header.
func getUserFromAuthorizationHeader(a string) string {
	source := a[len("Bearer"):]
	userpass := make([]byte, 4*len(source)/3)
	_, _ = base64.NewDecoder(base64.StdEncoding, strings.NewReader(source)).Read(userpass)
	return strings.Split(string(userpass), ":")[0]
}

// getLastEvent returns the most recent event with the given ID.
func (s *Server) getLastEvent(c *gin.Context) {
	a := c.GetHeader("Authorization")
	source := getUserFromAuthorizationHeader(a)
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing source parameter"})
		return

	}
	eventid := c.GetHeader("Event")
	if eventid == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing Event header"})
		return
	}
	event, err := s.store.GetLastEvent(source, eventid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, event)
}

func (s *Server) isAuthenticated(c *gin.Context) {
	a := c.GetHeader("Authorization")
	source := getUserFromAuthorizationHeader(a)
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing source parameter"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"error": nil})
}

// addEvent adds a new event to the store.
func (s *Server) addEvent(c *gin.Context) {
	var event map[string]interface{}
	err := c.BindJSON(&event)
	a := c.GetHeader("Authorization")
	source := getUserFromAuthorizationHeader(a)
	if source == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing source parameter"})
		return
	}

	eventId := c.GetHeader("Event")
	if eventId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing Event header"})
		return
	}

	var eventData []byte
	eventData, err = json.Marshal(event)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	d := c.Copy()
	go func() {
		err = s.store.AddEvent(eventId, source, eventData)
		if err != nil {
			d.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		d.Status(http.StatusCreated)
	}()
}

// Serve starts the server on the given address.
func (s *Server) Serve(addr string) error {
	gin.DisableConsoleColor()
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     s.allowedOrigins,
		AllowMethods:     []string{"GET", "POST"},
		AllowHeaders:     []string{"Authorization", "Event", "Content-Type", "Origin"},
		ExposeHeaders:    []string{"Content-Length", "Authorization", "Event"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	a := r.Group("/", gin.BasicAuth(s.authorizer.authentication))

	// Logging to a file.
	f, err := os.OpenFile(s.logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		return err
	}
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)
	r.Use(gin.LoggerWithWriter(gin.DefaultWriter))
	a.Use(gin.LoggerWithWriter(gin.DefaultWriter))

	a.POST("/event", s.addEvent)
	a.GET("/event", s.getLastEvent)
	a.GET("/auth", s.isAuthenticated)
	r.GET("/ping", func(c *gin.Context) {
		err = s.store.AddEvent("ping", "server", []byte(`{"data":"pong", "last_ping":"`+time.Now().Format(time.RFC3339)+`"}`))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ev, err2 := s.store.GetLastEvent("server", "ping")
		if err2 != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err2.Error()})
			return
		}
		c.JSON(http.StatusOK, ev)
	})
	return r.Run(addr)
}
