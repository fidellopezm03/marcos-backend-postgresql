package handler

import (
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
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

const cookieToken = "jwt"

func NewAdminHandler(repo repository.AdminRepo) *AdminHandler {

	return &AdminHandler{
		repo: repo,
	}
}
func (h *AdminHandler) RegisterRoutes(r chi.Router) {

	r.Post("/login", h.login)
	r.Get("/{img}", h.serveImg)
	r.Get("/data", h.getData)
	r.Route("/admin", func(r chi.Router) {
		r.Use(jwtauth.Verify(tokenAhuth, jwtauth.TokenFromCookie))
		r.Use(jwtauth.Authenticator)
		r.Get("/verify-session", h.verifySession)
		r.Get("/logout", h.Logout)
		r.Get("/info", h.getInfo)
		r.Get("/content", h.getContent)
		r.Post("/change-password", h.changePassword)
		r.Post("/save-img/{id}", h.saveImg)
		r.Post("/create-img/{id}", h.createImg)
		r.Post("/save-content/{id}", h.saveContent)
		r.Post("/create-content", h.createContent)

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
	username, err := h.repo.Authenticate(req.Usename, req.Password)
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": err.Error()})
		return
	}
	_, tokenString, err := tokenAhuth.Encode(map[string]interface{}{"username": username, "exp": jwtauth.ExpireIn(time.Hour)})
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "Failed to generate token"})
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:     cookieToken,
		Value:    tokenString,
		Path:     "/admin",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteLaxMode,
		Expires:  time.Now().Add(time.Hour),
	})
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "Login successful"})
}
func (h *AdminHandler) Logout(w http.ResponseWriter, r *http.Request) {
	http.SetCookie(w, &http.Cookie{
		Name:  cookieToken,
		Value: "",
		Path:  "/admin",
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
func (h *AdminHandler) verifySession(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(cookieToken)
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Unauthorized"})
		return
	}
	jsonstr := cookie.Value

	jwt, err := tokenAhuth.Decode(jsonstr)
	if err != nil {
		render.Status(r, http.StatusUnauthorized)
		render.JSON(w, r, map[string]string{"error": "Unauthorized"})
		return
	}

	if username, ok := jwt.Get("username"); ok {
		username := username.(string)
		render.Status(r, http.StatusOK)
		render.JSON(w, r, map[string]string{"user_id": username})
		return
	}
	render.Status(r, http.StatusUnauthorized)
	render.JSON(w, r, map[string]string{"error": "Unauthorized"})

}

func (h *AdminHandler) serveImg(w http.ResponseWriter, r *http.Request) {
	imgstr := chi.URLParam(r, "img")

	path, err := h.repo.FindImg(imgstr)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error en el código de la imagen"})
		return
	}

	http.ServeFile(w, r, path)
}

func (h *AdminHandler) saveImg(w http.ResponseWriter, r *http.Request) {
	idImg := chi.URLParam(r, "id")
	if len(idImg) == 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error in url param"})
		return
	}
	id, err := strconv.ParseInt(idImg, 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error in url param"})
		log.Println(err)
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error read photo"})
		return
	}
	defer file.Close()
	filename := header.Filename
	dstPath := filepath.Join("."+repository.UploadDir, filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error creating file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error creating file"})
		return
	}

	pathOldImg, err := h.repo.SaveImg(id, dstPath)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error saving file"})
		log.Println(err)
		return
	}
	if _, err = os.Stat(pathOldImg); os.IsNotExist(err) {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error finding old img"})
		log.Println(err)
		return
	}
	if err = os.Remove(pathOldImg); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error in deleting img"})
		log.Println(err)
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "img save"})

}
func (h *AdminHandler) createImg(w http.ResponseWriter, r *http.Request) {
	idContentstr := chi.URLParam(r, "id")

	if len(idContentstr) == 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error in url param"})
		return
	}
	idContent, err := strconv.ParseInt(idContentstr, 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error in url param"})
		log.Println(err)
		return
	}
	file, header, err := r.FormFile("photo")
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error read photo"})
		return
	}
	defer file.Close()
	filename := header.Filename
	dstPath := filepath.Join("."+repository.UploadDir, filename)
	dst, err := os.Create(dstPath)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error creating file"})
		return
	}
	defer dst.Close()

	if _, err := io.Copy(dst, file); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error creating file"})
		return
	}

	var datos struct {
		Name string `json:"name"`
	}
	datosStr := r.FormValue("datos")
	if err = json.Unmarshal([]byte(datosStr), &datos); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error reading data"})
		log.Println(err)
		return
	}
	id, err := h.repo.CreateImg(idContent, dstPath, datos.Name)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error saving file"})
		log.Println(err)
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]int64{"id": id})

}

func (h *AdminHandler) createContent(w http.ResponseWriter, r *http.Request) {
	var data struct {
		Title    string `json:"title"`
		Contnet  string `json:"content"`
		Location string `json:"location"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error saving content"})
		log.Println(err)
		return
	}
	id, err := h.repo.CreateContent(data.Title, data.Contnet, data.Location)
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error saving content"})
		log.Println(err)
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]int64{"id": id})

}
func (h *AdminHandler) saveContent(w http.ResponseWriter, r *http.Request) {
	idContentstr := chi.URLParam(r, "id")
	if len(idContentstr) == 0 {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error in url param"})
		return
	}
	idContent, err := strconv.ParseInt(idContentstr, 10, 64)
	if err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error in url param"})
		log.Println(err)
		return
	}
	var data struct {
		Title   string `json:"title"`
		Contnet string `json:"content"`
	}
	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		render.Status(r, http.StatusBadRequest)
		render.JSON(w, r, map[string]string{"error": "error saving content"})
		log.Println(err)
		return
	}
	if err := h.repo.SaveContent(idContent, data.Title, data.Contnet); err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error saving content"})
		log.Println(err)
		return
	}
	render.Status(r, http.StatusOK)
	render.JSON(w, r, map[string]string{"message": "created"})

}
func (h *AdminHandler) getData(w http.ResponseWriter, r *http.Request) {
	response := make(map[string]interface{})

	info, err := h.repo.GetAllinfo()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error geting info"})
		log.Println(err)
		return
	}
	content, err := h.repo.GetAllcontent()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error geting content"})
		log.Println(err)
		return
	}
	for key, value := range info {
		response[key] = value
	}
	for key, value := range content {
		response[key] = value
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, response)

}
func (h *AdminHandler) getInfo(w http.ResponseWriter, r *http.Request) {

	info, err := h.repo.GetAllinfo()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error geting info"})
		log.Println(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, info)

}
func (h *AdminHandler) getContent(w http.ResponseWriter, r *http.Request) {

	content, err := h.repo.GetAllcontent()
	if err != nil {
		render.Status(r, http.StatusInternalServerError)
		render.JSON(w, r, map[string]string{"error": "error geting content"})
		log.Println(err)
		return
	}

	render.Status(r, http.StatusOK)
	render.JSON(w, r, content)

}
