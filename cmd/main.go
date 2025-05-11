package main

import (
    "fmt"
    "log"
    "net/http"
    "github.com/gorilla/mux"
    "borrow-service/internal/controllers"
    "borrow-service/internal/db"
)

func main() {
    db.InitDB()
    db.InitKafkaProducer()

    router := mux.NewRouter()

    router.HandleFunc("/borrow", controllers.BorrowBook).Methods("POST")
    router.HandleFunc("/return", controllers.ReturnBook).Methods("PUT")
    router.HandleFunc("/borrowings/{userId}", controllers.GetBorrowingsByUser).Methods("GET")
    router.HandleFunc("/borrowings/overdue", controllers.GetOverdueBorrowings).Methods("GET")
    router.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        fmt.Fprintln(w, "Borrow Service is healthy")
    }).Methods("GET")

    fmt.Println("Borrow Service running on port 8080...")
    log.Fatal(http.ListenAndServe(":8080", router))
}

