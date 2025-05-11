# **Borrow Service**
The **Borrow Service** manages book borrow and return operations. It ensures users can **borrow** books, **return** them, and **track overdue books** while interacting with the **Book Service** to check availability.

## **Features**
1. **Allows users to borrow and return books**  
2. **Tracks borrowed books in MySQL**  
3. **Checks book availability via API calls to Book Service** 
4. **Asynchronously communicates to notification service via producing kafka events on apt topics**
4. **Computes `due_date` dynamically in code (14 days from borrow date)**  
5. **Provides overdue tracking**  

---

## **Database Schema**
The Borrow Service uses **MySQL** to store borrow transactions independently of Book Service.

```sql
CREATE TABLE borrowings (
    id INT AUTO_INCREMENT PRIMARY KEY,
    user_id VARCHAR(36) NOT NULL,
    book_id VARCHAR(36) NOT NULL,
    borrow_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```
1. **`due_date` is computed dynamically in code (not stored in DB)**  
2. **No foreign key constraints—Borrow Service interacts with Book Service via API calls**  

---

## **API Endpoints & Detailed Flow**
### **1. Borrow a Book**
**POST** `/borrow`  
**Checks book availability, stores the borrow record, and updates Book Service**

#### **Request**
```json
{
  "userId": "123",
  "bookId": "book_456"
}
```
#### **Flow**
1️⃣ **Check if book is available** by calling Book Service:
   ```http
   GET /books/book_456
   ```
2️⃣ If **available**, store borrow record:
   ```sql
   INSERT INTO borrowings (user_id, book_id, borrow_date) 
   VALUES ('123', 'book_456', NOW());
   ```
3️⃣ **Update book availability** in Book Service:
   ```http
   PUT /books/book_456
   ```
   ```json
   { "available": false }
   ```

#### **Response**
```json
{
  "message": "Book borrowed successfully"
}
```

---

### **2. Return a Book**
**PUT** `/return`  
**Deletes borrow record and updates Book Service to make book available, also produces an event in kafka for topic book.available_reserved**

#### **Request**
```json
{
  "userId": "123",
  "bookId": "book_456"
}
```
#### **Flow**
1️⃣ **Delete borrow record** from MySQL:
   ```sql
   DELETE FROM borrowings WHERE user_id = '123' AND book_id = 'book_456';
   ```
2️⃣ **Update book availability** in Book Service:
   ```http
   PUT /books/book_456
   ```
   ```json
   { "available": true }
   ```

#### **Response**
```json
{
  "message": "Book returned successfully"
}
```

---

### **3. Get Borrowed Books by User**
**GET** `/borrowings/{userId}`  
**Retrieves all books currently borrowed by a user**

#### **Request**
```http
GET /borrowings/123
```
#### **Flow**
1️⃣ Fetch borrow records from MySQL:
   ```sql
   SELECT book_id, borrow_date FROM borrowings WHERE user_id = '123';
   ```
2️⃣ Compute **due dates dynamically**:
   ```go
   dueDate := borrowDate.Add(14 * 24 * time.Hour)
   ```
#### **Response**
```json
[
  {
    "bookId": "book_456",
    "borrowDate": "2025-05-01T14:00:00Z",
    "dueDate": "2025-05-15T14:00:00Z"
  }
]
```

---

### **4. Get Overdue Borrowings**
**GET** `/borrowings/overdue`  
**Finds borrow records that are overdue (past 14 days) and produces an event in kafka for topic book.overdue**

#### **Request**
```http
GET /borrowings/overdue
```
#### **Flow**
1️⃣ Find overdue books in MySQL:
   ```sql
   SELECT user_id, book_id, borrow_date 
   FROM borrowings 
   WHERE borrow_date < DATE_SUB(NOW(), INTERVAL 14 DAY);
   ```
2️⃣ Compute **overdue status dynamically** in code:
   ```go
   if borrowDate.Add(14 * 24 * time.Hour).Before(time.Now()) {
       // Book is overdue
   }
   ```
#### **Response**
```json
[
  {
    "userId": "123",
    "bookId": "book_456",
    "borrowDate": "2025-04-10T14:00:00Z",
    "dueDate": "2025-04-24T14:00:00Z"
  }
]
```

---

## **Borrow Service Architecture**
**Backend:** Golang  
   - Handles book borrowing and return logic.  
   - Calls Book Service to check availability and update book status.  

**Database:** MySQL  
   - Stores borrowing records (`user_id`, `book_id`, `borrow_date`).  
   - Ensures transactional integrity for borrow/return operations.  

**Inter-Service Communication:** Kubernetes DNS  
   - Uses `book-service.default.svc.cluster.local` for book status updates.  
   - Enables seamless communication between services within the cluster.  

**Produces events to Kafka**
   - Uses the exposed kafka endpoints to connect to kafka
   - Sends events to topics book.available_reserved and book.overdue

**Deployment:** Kubernetes (Dockerized)  
   - Runs as a **scalable deployment** with **ClusterIP** service exposure.  
   - MySQL database is deployed separately with **Persistent Volumes (PV)** for data retention.  

**Data Persistence:** MySQL with PV & PVC  
   - Borrow records are stored **persistently** using Kubernetes **Persistent Volume Claim (PVC)**.  
   - Ensures **data is retained** even if MySQL pods restart.  
---

## **Setup & Deployment**
### **1. Build & Push Docker Image**
```sh
docker build -t your-dockerhub-username/borrow-service .
docker push your-dockerhub-username/borrow-service:latest
```

### **2. Deployment in Kubernetes**
The **Borrow Service** is deployed within a Kubernetes cluster alongside MySQL, ensuring smooth inter-service communication.

#### **2.1 Borrow Service Deployment**
1. Creates a **pod** for Borrow Service using the Docker image.  
2. Runs on **port 5000**, accessible within Kubernetes.  
3. Can be **scaled dynamically** to handle high traffic loads.  

#### **2.2 Borrow Service Kubernetes Service**
1. Exposes Borrow Service **internally** via `ClusterIP`.  
2. Accessible via Kubernetes DNS at:  
   ```
   borrow-service.default.svc.cluster.local
   ```
3. Used by **Book Service** to send borrow and return requests.

#### **2.3 MySQL Deployment**
1. Deploys MySQL **inside Kubernetes** to store borrow records.  
2. Initializes the **borrow_service** database using environment variables.  
3. Runs on **port 3306** to handle database transactions.

#### **2.4 MySQL Kubernetes Service**
1. Exposes MySQL **internally** (`ClusterIP`) for database communication.  
2. Borrow Service connects to MySQL via:  
   ```
   mysql-service.default.svc.cluster.local:3306
   ```
3. Manages **database networking inside the cluster** without external exposure.

#### **2.5 Persistent Storage (PV & PVC)**
1. **Persistent Volume (PV)** provides static storage for MySQL.  
2. **Persistent Volume Claim (PVC)** ensures MySQL retains data across pod restarts.  
3. Prevents **data loss when MySQL pods restart** by preserving borrow records.

### **3. Apply Kubernetes Deployment**
```sh
kubectl apply -f borrow-service-deployment.yaml
```

### **4. Verify Deployment**
```sh
kubectl get pods
kubectl get services
```