package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/auth0/go-jwt-middleware"
	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/sessions"
)

type Device struct {
	Id   string
	Name string
}

func (p *Device) save() error {
	res1B, _ := json.Marshal(p)
	fmt.Println(string(res1B))
	filename := "devices.json"
	return ioutil.WriteFile(filename, res1B, 0600)
}

func load(title string) (*Device, error) {
	filename := "devices.json"
	device := Device{}
	body, err := ioutil.ReadFile(filename)
	json.Unmarshal(body, &device)
	if err != nil {
		return nil, err
	}
	return &Device{Id: device.Id, Name: device.Name}, nil
}

var store = sessions.NewCookieStore([]byte("something-very-secret"))

func configureLogger() {
	f, err := os.OpenFile("log.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf("error opening file: %v", err)
	}

	defer f.Close()
	log.SetOutput(f)
	log.Println("hello")
}

func sessionCheckingHandler(fn func(http.ResponseWriter, *http.Request)) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, err := store.Get(r, "session-name")
		if err != nil {
			http.Error(w, fmt.Sprintf("The session expired \n"), http.StatusInternalServerError)
			return
		}
		userName := session.Values["user"]
		log.Printf("%v", userName)
		if userName != "adam" {
			http.Error(w, fmt.Sprintf("Invalid login \n"), http.StatusInternalServerError)
			return
		}

		fn(w, r)
	}
}

type User struct {
	Guid           string
	Name           string
	Email          string
	ServiceCookies map[string]string
}

type AuthenticateResponse struct {
	IsSuccessful bool
	User         User
}

func tokenAuth(w http.ResponseWriter, r *http.Request) {
	log.Printf("accessed \n")
}

func main() {
	configureLogger()

	r := mux.NewRouter()
	de := Device{Id: "aa", Name: "bb"}
	de.save()

	r.Handle("/get-token", GetTokenHandler).Methods("GET")
	r.Handle("/status", jwtMiddleware.Handler(ProductsHandler)).Methods("GET")
	r.PathPrefix("/").Handler(jwtMiddleware.Handler(http.FileServer(http.Dir("./client/dist/"))))
	http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r))
}

/* Set up a global string for our secret */
var mySigningKey = []byte("secret")

/* Handlers */
var GetTokenHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	/* Create the token */
	token := jwt.New(jwt.SigningMethodHS256)

	/* Create a map to store our claims*/
	claims := token.Claims.(jwt.MapClaims)

	/* Set token claims */
	claims["iss"] = "golang"
	claims["name"] = "Ado Kukic"
	claims["exp"] = time.Now().Add(time.Hour * 24).Unix()

	/* Sign the token with our secret */
	tokenString, _ := token.SignedString(mySigningKey)

	/* Finally, write the token to the browser window */
	w.Write([]byte(tokenString))
})

var jwtMiddleware = jwtmiddleware.New(jwtmiddleware.Options{
	ValidationKeyGetter: func(token *jwt.Token) (interface{}, error) {
		return mySigningKey, nil
	},
	SigningMethod: jwt.SigningMethodHS256,
})

type Product struct {
	Id          int
	Name        string
	Slug        string
	Description string
}

/* We will create our catalog of VR experiences and store them in a slice. */
var products = []Product{
	Product{Id: 1, Name: "Hover Shooters", Slug: "hover-shooters", Description: "Shoot your way to the top on 14 different hoverboards"},
	Product{Id: 2, Name: "Ocean Explorer", Slug: "ocean-explorer", Description: "Explore the depths of the sea in this one of a kind underwater experience"},
	Product{Id: 3, Name: "Dinosaur Park", Slug: "dinosaur-park", Description: "Go back 65 million years in the past and ride a T-Rex"},
	Product{Id: 4, Name: "Cars VR", Slug: "cars-vr", Description: "Get behind the wheel of the fastest cars in the world."},
	Product{Id: 5, Name: "Robin Hood", Slug: "robin-hood", Description: "Pick up the bow and arrow and master the art of archery"},
	Product{Id: 6, Name: "Real World VR", Slug: "real-world-vr", Description: "Explore the seven wonders of the world in VR"},
}

/* The status handler will be invoked when the user calls the /status route
It will simply return a string with the message "API is up and running" */
var StatusHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("API is up and running"))
})

/* The products handler will be called when the user makes a GET request to the /products endpoint.
This handler will return a list of products available for users to review */
var ProductsHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// Here we are converting the slice of products to json
	payload, _ := json.Marshal(products)

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(payload))
})
