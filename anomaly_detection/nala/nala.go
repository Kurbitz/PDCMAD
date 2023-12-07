package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type album struct {
	ID     string  `json:"id"`
	Title  string  `json:"title"`
	Artist string  `json:"artist"`
	Price  float64 `json:"price"`
}

type metric struct {
	ID   string  `json:"id"`
	Host string  `json:"host"`
	Data float64 `json:"data"`
}

func (m *metric) getID() (string, error) {
	foo := &m.ID
	*foo = "5"
	return m.ID, nil
}

// albums slice to seed record album data.
var albums = []album{
	{ID: "1", Title: "Blue Train", Artist: "John Coltrane", Price: 56.99},
	{ID: "2", Title: "Jeru", Artist: "Gerry Mulligan", Price: 17.99},
	{ID: "3", Title: "Sarah Vaughan and Clifford Brown", Artist: "Sarah Vaughan", Price: 39.99},
}

var metrics = []metric{
	{ID: "1", Host: "server9000", Data: 123.231341},
	{ID: "2", Host: "server9001", Data: 0},
	{ID: "3", Host: "server9002", Data: 124.231341},
}

func getMetrics(ctx *gin.Context) {
	ctx.IndentedJSON(http.StatusOK, metrics)
}

// getAlbumByID locates the album whose ID value matches the id
// parameter sent by the client, then returns that album as a response.
func getAlbumByID(c *gin.Context) {
	id := c.Param("id")

	// Loop over the list of albums, looking for
	// an album whose ID value matches the parameter.
	for _, a := range albums {
		if a.ID == id {
			c.IndentedJSON(http.StatusOK, a)
			return
		}
	}
	c.IndentedJSON(http.StatusNotFound, gin.H{"message": "album not found"})
}

// getAlbums responds with the list of all albums as JSON.
func getAlbums(c *gin.Context) {
	c.IndentedJSON(http.StatusOK, albums)
}

// postAlbums adds an album from JSON received in the request body.
func postAlbums(c *gin.Context) {
	var newAlbum album

	// Call BindJSON to bind the received JSON to
	// newAlbum.
	if err := c.BindJSON(&newAlbum); err != nil {
		return
	}
	{
		err := c.BindJSON(&newAlbum)
		if err != nil {
			return
		}
	}
	// Add the new album to the slice.
	albums = append(albums, newAlbum)
	c.IndentedJSON(http.StatusCreated, newAlbum)
}

func main() {
	router := gin.Default()
	foo, _ := metrics[0].getID()
	fmt.Println(foo)
	router.GET("/albums", getAlbums)
	router.GET("/metrics", getMetrics)
	router.Run("localhost:8080")
}
