package handlers

import (
	"encoding/json"
	"log"
	"net/http"

	"notes-api/db"
	"notes-api/models"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

func CreateNote(w http.ResponseWriter, r *http.Request) {
	var note models.Note

	json.NewDecoder(r.Body).Decode(&note)
	note.ID = uuid.New().String()
	log.Println("Creating note with ID:", note.ID)
	log.Println("Creating note with title:", note.Title)
	log.Println(note)

	db.DB.Create(&note)
	json.NewEncoder(w).Encode(note)
}

func GetNotes(w http.ResponseWriter, r *http.Request) {
	var notes []models.Note
	db.DB.Find(&notes)
	json.NewEncoder(w).Encode(notes)
}

func GetNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var note models.Note
	if err := db.DB.First(&note, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	json.NewEncoder(w).Encode(note)
}

func UpdateNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var note models.Note
	if err := db.DB.First(&note, id).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			http.Error(w, "Note not found", http.StatusNotFound)
		} else {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}
	json.NewDecoder(r.Body).Decode(&note)
	db.DB.Save(&note)
	json.NewEncoder(w).Encode(note)
}

func DeleteNote(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	var note models.Note
	if err := db.DB.Delete(&note, id).Error; err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	json.NewEncoder(w).Encode("Note deleted successfully")
}
