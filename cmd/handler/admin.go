package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/fidellopezm03/marcos-backend-postgresql/cmd/repository"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/jwtauth"
	"github.com/go-chi/render"
)

type AdminHandler struct {
	repo repository.AdminRepo
}

var tokenAhuth = jwtauth.New("HS256", []byte("secret"), nil)

func NewAdminHandler(repo repository.AdminRepo) *AdminHandler {
	return &AdminHandler{
		repo: repo,
	}
}
func (h *AdminHandler) RegisterRoutes(r chi.Router) {
	r.Post("/login", h.login)

	r.Route("/admin", func(r chi.Router) {
		r.Use(jwtauth.Verify(tokenAhuth, jwtauth.TokenFromCookie))
		r.Use(jwtauth.Authenticator)
		r.Get("/logout", h.Logout)
		r.Post("/change-password", h.changePassword)
	})

}

func (h *AdminHandler) login(w http.ResponseWriter, r *http.Request) {
	type LoginRequest struct {
		Usename  string `json:"username"`
		Password string `json:"password"`
	}
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}
	id, err := h.repo.Authenticate(req.Usename, req.Password)
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}
	_, tokenString, err := tokenAhuth.Encode(map[string]interface{}{"user_id": id, "exp": jwtauth.ExpireIn(60 * 60 * time.Second)})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to generate token"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     "jwt",
		Value:    tokenString,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(60 * 60 * time.Second),
	})
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "Login successful"})
}
func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  "jwt",
		Value: "",
	})
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "Logout successful"})
}

func (h *AdminHandler) changePassword(w http.ResponseWriter, r *http.Request) {

	type ChangePasswordRequest struct {
		ID          int64  `json:"id"`
		OldPassword string `json:"old_password"`
		NewPassword string `json:"new_password"`
	}
	var req ChangePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "Invalid request body"})
		return
	}
	if err := h.repo.ChangePassword(req.ID, req.OldPassword, req.NewPassword); err != nil {
		if err.Error() == "user not found" || err.Error() == "old password is incorrect" {
			render.Status(r, http.StatusBadRequest)
			render.JSON(w, r, map[string]string{"error": err.Error()})
			return
		}
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to change password"})

		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "Password changed successfully"})
}
