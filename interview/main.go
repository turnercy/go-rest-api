package main

import (
	"database/sql"
	//"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	_ "github.com/go-sql-driver/mysql"
)

type PersonInfo struct {
	Name        string `json:"name"`
	PhoneNumber string `json:"phone_number"`
	City        string `json:"city"`
	State       string `json:"state"`
	Street1     string `json:"street1"`
	Street2     string `json:"street2"`
	ZipCode     string `json:"zip_code"`
}

func main() {
	// Connect to MySQL database
	db, err := sql.Open("mysql", "root:password@tcp(127.0.0.1:3306)/testdb") 
	if err != nil {
		panic(err)
	}
	defer db.Close()


	router := gin.Default()

	// Task 1: make a REST endpoint (GET)
	router.GET("/person/:person_id/info", func(c *gin.Context) {
		personID := c.Param("person_id")

		query := `
			SELECT p.name, ph.number, a.city, a.state, a.street1, a.street2, a.zip_code
			FROM person p
			JOIN phone ph ON p.id = ph.person_id
			JOIN address_join aj ON p.id = aj.person_id
			JOIN address a ON aj.address_id = a.id
			WHERE p.id = ? LIMIT 1;
		`

		var personInfo PersonInfo
		err := db.QueryRow(query, personID).Scan(
			&personInfo.Name,
			&personInfo.PhoneNumber,
			&personInfo.City,
			&personInfo.State,
			&personInfo.Street1,
			&personInfo.Street2,
			&personInfo.ZipCode,
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch person info"})
			return
		}

		c.JSON(http.StatusOK, personInfo)
	})

	// Task 2: make a REST endpoint (POST)
	router.POST("/person/create", func(c *gin.Context) {
		var personInfo PersonInfo
		if err := c.BindJSON(&personInfo); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid input"})
			return
		}

		// Insert Person
		result, err := db.Exec("INSERT INTO person (name) VALUES (?)", personInfo.Name)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting person"})
			return
		}
		personID, _ := result.LastInsertId()

		// Insert phone
		result, err = db.Exec("INSERT INTO phone (number, person_id) VALUES (?, ?)", personInfo.PhoneNumber, personID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting phone"})
			return
		}

		// Insert address
		result, err = db.Exec("INSERT INTO address (city, state, street1, street2, zip_code) VALUES (?, ?, ?, ?, ?)",
			personInfo.City, personInfo.State, personInfo.Street1, personInfo.Street2, personInfo.ZipCode)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting address"})
			return
		}
		addressID, _ := result.LastInsertId()

		// Insert address join
		_, err = db.Exec("INSERT INTO address_join (person_id, address_id) VALUES (?, ?)", personID, addressID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Error inserting address join"})
			return
		}

		c.JSON(http.StatusOK, gin.H{"message": "Person created successfully"})
	})
	
	router.Run(":8080")
}
