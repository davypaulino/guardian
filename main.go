package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"

	"github.com/markbates/goth"
	"github.com/markbates/goth/gothic"
)

var db *sql.DB

var UserAccount = map[string]func(goth.User) User{
    "github": NewGithubUser,
}

func callbackHandler(w http.ResponseWriter, r *http.Request) {
	user, err := gothic.CompleteUserAuth(w, r)
	fmt.Printf("User Info: %+v\n", user)
	if err != nil {
		http.Error(w, "Authentication failed: "+err.Error(), http.StatusUnauthorized)
		return
	}

	if user.AccessToken == "" {
		http.Error(w, "No access token received", http.StatusInternalServerError)
		return
	}

	newUser := UserAccount[user.Provider](user)
	_, err = db.Exec(`
		INSERT INTO users (id, provider, provider_user_id, nickname,
			email, avatar_url, provider_access_token,
			provider_refresh_token, updated_at, status, "role") 
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		ON CONFLICT (provider_user_id) DO UPDATE SET 
			provider_access_token = $7,
			provider_refresh_token = $8,
			updated_at = NOW();`,
		newUser.ID, newUser.Provider, newUser.ProviderUserID,
		newUser.NickName, newUser.Email, newUser.ImgURL,
		newUser.ProviderAccessToken, newUser.ProviderRefreshToken,
		nil, newUser.Status, newUser.Role)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		log.Println("DB Error:", err)
		return
	}

	err = db.QueryRow(`
        SELECT nickname, email, avatar_url,
			access_token, refresh_token, status,
			role
        FROM users
        WHERE provider_user_id = $1`,
		newUser.ProviderUserID).Scan(&newUser.NickName, &newUser.Email, &newUser.ImgURL,
        	&newUser.AccessToken, &newUser.RefreshToken, &newUser.Status,
			&newUser.Role)

	token, refresh, _ := GenerateTokens(newUser)
	newUser.AccessToken = &token
	newUser.RefreshToken = &refresh

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		log.Println("DB Error:", err)
		return
	}

	_, err = db.Exec(`
		UPDATE users SET 
			access_token = $1,
			refresh_token = $2,
			updated_at = NOW()
		WHERE id = $3`,
		newUser.AccessToken, newUser.RefreshToken, newUser.ID)

	if err != nil {
		http.Error(w, "Database error: "+err.Error(), http.StatusInternalServerError)
		log.Println("DB Error:", err)
		return
	}
	
	redirectURL := fmt.Sprintf("http://localhost:3000/home?token=%s", *newUser.AccessToken)
	if (newUser.Status == Pending) {
		redirectURL = fmt.Sprintf("http://localhost:3000/register?token=%s", *newUser.AccessToken)
	}
	http.Redirect(w, r, redirectURL, http.StatusFound)
}

func getUserInfo(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	
	// Extract token from request
	authHeader := r.Header.Get("Authorization")
	fmt.Printf("%+v", r.Header)
	fmt.Printf("Token: %s", authHeader)
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}
	token := authHeader[len("Bearer "):]

	// Fetch user from the database using access token
	var user struct {
		ID        string `json:"id"`
		Name      string `json:"nickname"`
		Email     string `json:"email"`
		AvatarURL string `json:"avatar_url"`
	}

	fmt.Printf("Token: %s", token)

	err := db.QueryRow(`
        SELECT id, nickname, email, avatar_url 
        FROM users WHERE access_token = $1`, token).Scan(&user.ID, &user.Name, &user.Email, &user.AvatarURL)
	
	if err == sql.ErrNoRows {
		http.Error(w, "User not found or invalid token", http.StatusUnauthorized)
		return
	} else if err != nil {
		http.Error(w, "Database error", http.StatusInternalServerError)
		log.Println("Database query error:", err)
		return
	}

	// Return user data in JSON format
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(user)
}

func logoutHandler(res http.ResponseWriter, req *http.Request) {
	gothic.Logout(res, req)
	res.Header().Set("Location", "/")
	res.WriteHeader(http.StatusTemporaryRedirect)
}

func providerAuthHandler(res http.ResponseWriter, req *http.Request) {
	res.Header().Set("Access-Control-Allow-Origin", "*")
	res.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS") // Specify allowed HTTP methods
	res.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
	// try to get the user without re-authenticating
	if gothUser, err := gothic.CompleteUserAuth(res, req); err == nil {
		t, _ := template.New("foo").Parse(userTemplate)
		t.Execute(res, gothUser)
	} else {
		gothic.BeginAuthHandler(res, req)
	}
}

var indexTemplate = `{{range $key,$value:=.Providers}}
    <p><a href="/auth/{{$value}}">Log in with {{index $.ProvidersMap $value}}</a></p>
{{end}}`

var userTemplate = `
<p><a href="/logout/{{.Provider}}">logout</a></p>
<p>Name: {{.Name}} [{{.LastName}}, {{.FirstName}}]</p>
<p>Email: {{.Email}}</p>
<p>NickName: {{.NickName}}</p>
<p>Location: {{.Location}}</p>
<p>AvatarURL: {{.AvatarURL}} <img src="{{.AvatarURL}}"></p>
<p>Description: {{.Description}}</p>
<p>UserID: {{.UserID}}</p>
<p>AccessToken: {{.AccessToken}}</p>
<p>ExpiresAt: {{.ExpiresAt}}</p>
<p>RefreshToken: {{.RefreshToken}}</p>
`

func main() {
	Init()

	http.HandleFunc("/auth/{provider}/callback", callbackHandler)
	http.HandleFunc("/logout/{provider}", logoutHandler)
	http.HandleFunc("/auth/{provider}", providerAuthHandler)

	http.HandleFunc("/users/1", getUserInfo)

	log.Println("listening on localhost:3001")
	log.Fatal(http.ListenAndServe(":3001", nil))
}
