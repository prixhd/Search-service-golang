package main

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type XMLUser struct {
	ID        int    `xml:"id"`
	FirstName string `xml:"first_name"`
	LastName  string `xml:"last_name"`
	Age       int    `xml:"age"`
	About     string `xml:"about"`
	Gender    string `xml:"gender"`
}

type XMLDataset struct {
	Rows []XMLUser `xml:"row"`
}

var datasetPath = "dataset.xml"

func loadUsers() ([]User, error) {
	file, err := os.Open(datasetPath)
	var dataset XMLDataset

	if err != nil {
		return nil, err
	}
	defer file.Close()

	decoder := xml.NewDecoder(file)

	err = decoder.Decode(&dataset)

	if err != nil {
		return nil, err
	}
	userSlice := make([]User, 0, len(dataset.Rows))

	for _, user := range dataset.Rows {
		userSlice = append(userSlice, User{
			ID:     user.ID,
			Name:   user.FirstName + " " + user.LastName,
			Age:    user.Age,
			About:  user.About,
			Gender: user.Gender,
		})
	}

	return userSlice, nil
}

func SearchServer(w http.ResponseWriter, r *http.Request) {

	token := r.Header.Get("AccessToken")

	if token != "secret" {
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	users, err := loadUsers()

	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	query := r.URL.Query().Get("query")
	orderField := r.URL.Query().Get("order_field")

	orderBy, err := strconv.Atoi(r.URL.Query().Get("order_by"))
	if err != nil {
		http.Error(w, "Invalid order_by", http.StatusBadRequest)
		return
	}

	limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, "Invalid limit", http.StatusBadRequest)
		return
	}

	offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
	if err != nil {
		http.Error(w, "Invalid offset", http.StatusBadRequest)
		return
	}

	filteredUsers := make([]User, 0, len(users))

	for _, user := range users {
		if strings.Contains(user.Name, query) || strings.Contains(user.About, query) {
			filteredUsers = append(filteredUsers, user)
		}
	}

	switch orderField {
	case "Id", "Age", "Name", "":
		if orderBy != OrderByAsIs {
			sort.Slice(filteredUsers, func(i, j int) bool {
				switch orderField {
				case "Age":
					if orderBy == OrderByAsc {
						return filteredUsers[i].Age < filteredUsers[j].Age
					} else if orderBy == OrderByDesc {
						return filteredUsers[i].Age > filteredUsers[j].Age
					}
					return false

				case "Id":
					if orderBy == OrderByAsc {
						return filteredUsers[i].ID < filteredUsers[j].ID
					} else if orderBy == OrderByDesc {
						return filteredUsers[i].ID > filteredUsers[j].ID
					}
					return false

				case "Name", "":
					if orderBy == OrderByAsc {
						return filteredUsers[i].Name < filteredUsers[j].Name
					} else if orderBy == OrderByDesc {
						return filteredUsers[i].Name > filteredUsers[j].Name
					}
					return false

				}
				return false
			})
		}
	default:
		response := SearchErrorResponse{
			Error: ErrorBadOrderField,
		}

		data, err := json.Marshal(response)
		if err != nil {
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)

		_, err = w.Write(data)
		if err != nil {
			return
		}
		return
	}

	var result []User

	if offset < 0 || limit < 0 {
		http.Error(w, "Invalid pagination parameters", http.StatusBadRequest)
		return
	}

	if offset >= len(filteredUsers) {
		result = []User{}
	} else {
		end := offset + limit

		if end > len(filteredUsers) {
			end = len(filteredUsers)
		}

		result = filteredUsers[offset:end]
	}

	data, err := json.Marshal(result)
	if err != nil {
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	_, err = w.Write(data)

	if err != nil {
		return
	}

}
