package main

import (
	"fmt"
	"net/http"
	"sync"
)

var (
	users = map[string]string{
		"user1": "password123",
	}
	balances = map[string]float64{
		"user1": 100,
	}
	products = map[string]float64{
		"apple":  1.0,
		"banana": 0.5,
	}
	cart = make(map[string]map[string]int)
	mu   = &sync.Mutex{}
)

func main() {
	http.HandleFunc("/login", loginHandler)
	http.HandleFunc("/", indexHandler)
	http.HandleFunc("/add_to_cart/", addToCartHandler)
	http.HandleFunc("/checkout", checkoutHandler)
	http.ListenAndServe(":8080", nil)
}

func loginHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodPost {
		username := r.FormValue("username")
		password := r.FormValue("password")

		if storedPass, ok := users[username]; ok && storedPass == password {
			http.SetCookie(w, &http.Cookie{
				Name:  "session",
				Value: username,
			})
			http.Redirect(w, r, "/", http.StatusSeeOther)
			return
		}
		http.Error(w, "Invalid credentials", http.StatusUnauthorized)
		return
	}
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w,
		`<!DOCTYPE html>
		<html lang="en">
		<head>
			<meta charset="UTF-8">
			<meta name="viewport" content="width=device-width, initial-scale=1.0">
			<title>Login</title>
		</head>
		<body>
			<br>Login</h2>
			<form action="/login" method="post">
				<div>
					<label>Username:</label>
					<input type="text" name="username">
				</div>
				<div>
					<label>Password:</label>
					<input type="password" name="password">
				</div>
				<div>
					<input type="submit" value="Login">
				</div>
			</form>
		</body>
		</html>
	`)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		userCart = make(map[string]int)
		cart[username] = userCart
	}

	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><body>")
	fmt.Fprintf(w, "Hello, %s! Available products:<br>", username)
	fmt.Fprintf(w, "<ul>")
	for product, price := range products {
		fmt.Fprintf(w, "<li>%s: $%.2f <a href=\"/add_to_cart/%s\">Add to cart</a></li>", product, price, product)
	}
	fmt.Fprintf(w, "</ul>")
	fmt.Fprint(w, "Your cart:<br>")
	fmt.Fprintf(w, "<ul>")
	for product, quantity := range userCart {
		fmt.Fprintf(w, "<li>%s: %d</li>", product, quantity)
	}
	fmt.Fprintf(w, "</ul>")
	fmt.Fprint(w, `<a href="/checkout">Checkout</a>`)
	fmt.Fprintf(w, "</body></html>")
}

func addToCartHandler(w http.ResponseWriter, r *http.Request) {
	product := r.URL.Path[len("/add_to_cart/"):]
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		userCart = make(map[string]int)
		cart[username] = userCart
	}

	userCart[product]++
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func checkoutHandler(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie("session")
	if err != nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	mu.Lock()
	defer mu.Unlock()

	username := cookie.Value
	userCart, ok := cart[username]
	if !ok {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	total := 0.0
	for product, quantity := range userCart {
		price, ok := products[product]
		if !ok {
			continue
		}
		total += price * float64(quantity)
	}
	balances[username] -= total
	w.Header().Set("Content-Type", "text/html")
	fmt.Fprintf(w, "<html><body>")
	fmt.Fprintf(w, "</ul>")
	fmt.Fprintf(w, "Checkout successful!: $%.2f<br>", balances[username])
	fmt.Fprintf(w, "<ul>")
	fmt.Fprint(w, `<a href="/">Go back to home</a>`)
	fmt.Fprintf(w, "</body></html>")
}
