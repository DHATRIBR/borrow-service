package tests

import (
    "bytes"
    "encoding/json"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"

    "borrow-service/internal/controllers"
    "borrow-service/internal/db"

    "github.com/DATA-DOG/go-sqlmock"
    "github.com/stretchr/testify/mock"
)

// Mock Kafka Event Publisher
type MockKafka struct {
    mock.Mock
}

func (m *MockKafka) PublishEvent(topic, message string) {
    m.Called(topic, message)
}


func TestBorrowBook(t *testing.T) {
    mockDB, mockSQL, _ := sqlmock.New()
    db.DB = mockDB

    mockKafka := &MockKafka{}
    mockKafka.On("PublishEvent", "book.borrowed", `{"userId":"user1","bookId":"book1"}`)

    mockSQL.ExpectQuery("SELECT available FROM books WHERE id=?").
        WithArgs("book1").
        WillReturnRows(sqlmock.NewRows([]string{"available"}).AddRow(true))

    mockSQL.ExpectExec("INSERT INTO borrowings").
        WithArgs("user1", "book1").
        WillReturnResult(sqlmock.NewResult(1, 1))

    requestBody := map[string]string{"userId": "user1", "bookId": "book1"}
    body, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", "/borrow", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    http.HandlerFunc(controllers.BorrowBook).ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Got status %v, want %v", rr.Code, http.StatusOK)
    }

    expected := `{"message":"Book borrowed successfully"}`
    actual := strings.TrimSpace(rr.Body.String())
    if actual != expected {
        t.Errorf("Got body %v, want %v", actual, expected)
    }

    mockKafka.AssertCalled(t, "PublishEvent", "book.borrowed", `{"userId":"user1","bookId":"book1"}`)
}

func TestReturnBook(t *testing.T) {
    mockDB, mockSQL, _ := sqlmock.New()
    db.DB = mockDB

    mockKafka := &MockKafka{}
    mockKafka.On("PublishEvent", "book.available_reserved", `{"userId":"user1","bookId":"book1"}`)

    mockSQL.ExpectQuery("SELECT COUNT FROM borrowings WHERE user_id=? AND book_id=?").
        WithArgs("user1", "book1").
        WillReturnRows(sqlmock.NewRows([]string{"count"}).AddRow(1))

    mockSQL.ExpectExec("DELETE FROM borrowings").
        WithArgs("user1", "book1").
        WillReturnResult(sqlmock.NewResult(1, 1))

    requestBody := map[string]string{"userId": "user1", "bookId": "book1"}
    body, _ := json.Marshal(requestBody)
    req, _ := http.NewRequest("POST", "/return", bytes.NewBuffer(body))
    req.Header.Set("Content-Type", "application/json")

    rr := httptest.NewRecorder()
    http.HandlerFunc(controllers.ReturnBook).ServeHTTP(rr, req)

    if rr.Code != http.StatusOK {
        t.Errorf("Got status %v, want %v", rr.Code, http.StatusOK)
    }

    expected := `{"message":"Book returned successfully"}`
    actual := strings.TrimSpace(rr.Body.String())
    if actual != expected {
        t.Errorf("Got body %v, want %v", actual, expected)
    }

    mockKafka.AssertCalled(t, "PublishEvent", "book.available_reserved", `{"userId":"user1","bookId":"book1"}`)
}
