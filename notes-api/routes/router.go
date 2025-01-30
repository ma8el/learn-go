package routes

import (
	"notes-api/handlers"

	"github.com/go-chi/chi/v5"
)

func SetupRouter() *chi.Mux {
	r := chi.NewRouter()
	r.Post("/notes", handlers.CreateNote)
	r.Get("/notes", handlers.GetNotes)
	r.Get("/notes/{id}", handlers.GetNote)
	r.Put("/notes/{id}", handlers.UpdateNote)
	r.Delete("/notes/{id}", handlers.DeleteNote)
	return r
}
