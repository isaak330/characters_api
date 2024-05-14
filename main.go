package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

type Character struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Photo       string `json:"photo"`
}

func main() {
	//connect to database
	db, err := sql.Open("postgres", os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	//create the table if it doesn't exist
	_, err = db.Exec("CREATE TABLE IF NOT EXISTS characters (id SERIAL PRIMARY KEY, name TEXT, description TEXT,photo TEXT)")

	if err != nil {
		log.Fatal(err)
	}

	//create router
	router := mux.NewRouter()
	router.HandleFunc("/characters", getCharacters(db)).Methods("GET")
	router.HandleFunc("/characters/{id}", getCharacter(db)).Methods("GET")
	router.HandleFunc("/characters", createCharacter(db)).Methods("POST")
	router.HandleFunc("/characters/{id}", updateCharacter(db)).Methods("PUT")
	router.HandleFunc("/characters/{id}", deleteCharacter(db)).Methods("DELETE")

	//start server
	log.Fatal(http.ListenAndServe(":8000", jsonContentTypeMiddleware(router)))
}

func jsonContentTypeMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		next.ServeHTTP(w, r)
	})
}

// get all users
func getCharacters(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		rows, err := db.Query("SELECT * FROM characters")
		if err != nil {
			log.Fatal(err)
		}
		defer rows.Close()

		characters := []Character{}
		for rows.Next() {
			var c Character
			if err := rows.Scan(&c.ID, &c.Name, &c.Description, &c.Photo); err != nil {
				log.Fatal(err)
			}
			characters = append(characters, c)
		}
		if err := rows.Err(); err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(characters)
	}
}

// get user by id
func getCharacter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var c Character
		err := db.QueryRow("SELECT * FROM characters WHERE id = $1", id).Scan(&c.ID, &c.Name, &c.Description, &c.Photo)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		json.NewEncoder(w).Encode(c)
	}
}

// create user
func createCharacter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Character
		json.NewDecoder(r.Body).Decode(&c)

		err := db.QueryRow("INSERT INTO characters (name, description, photo) VALUES ($1, $2, $3) RETURNING id", c.Name, c.Description, c.Photo).Scan(&c.ID)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(c)
	}
}

// update user
func updateCharacter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var c Character
		json.NewDecoder(r.Body).Decode(&c)

		vars := mux.Vars(r)
		id := vars["id"]

		_, err := db.Exec("UPDATE characters SET name = $1, description = $2, photo = $3 WHERE id = $4", c.Name, c.Description, c.Photo, id)
		if err != nil {
			log.Fatal(err)
		}

		json.NewEncoder(w).Encode(c)
	}
}

// delete user
func deleteCharacter(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		id := vars["id"]

		var c Character
		err := db.QueryRow("SELECT * FROM characters WHERE id = $1", id).Scan(&c.ID, &c.Name, &c.Description, &c.Photo)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			return
		} else {
			_, err := db.Exec("DELETE FROM characters WHERE id = $1", id)
			if err != nil {
				//todo : fix error handling
				w.WriteHeader(http.StatusNotFound)
				return
			}

			json.NewEncoder(w).Encode("User deleted")
		}
	}
}
