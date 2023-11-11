# Go REST API

This project implements a Go-based REST API for managing users and associated documents. It utilizes the Gin framework for handling routes and HTTP requests. The application supports user authentication, user creation, fetching API version information, retrieving documents associated with users, and creating new documents.

## Project Structure

- **main.go**: Contains the main logic of the application, including route configuration and HTTP request handling.
- **docs/**: Directory storing documents associated with users.

## Dependencies

- [Gin](https://github.com/gin-gonic/gin): Framework for building APIs in Go.
- [bcrypt](https://golang.org/x/crypto/bcrypt): Library for secure password hashing.
- [jwt-go](https://github.com/dgrijalva/jwt-go): JSON Web Tokens for authentication.

## Endpoints

1. **Get API Version**
   - Method: `GET`
   - Path: `/version`

2. **Sign Up New User**
   - Method: `POST`
   - Path: `/signup`
   - Request Body:
     ```json
     {
       "username": "newUser",
       "password": "newPassword"
     }
     ```

3. **Log In**
   - Method: `POST`
   - Path: `/login`
   - Request Body:
     ```json
     {
       "username": "existingUser",
       "password": "existingPassword"
     }
     ```
   
4. **Get Documents Associated with a User**
   - Method: `GET`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`

5. **Add New Document**
   - Method: `POST`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`
   - Request Body: Raw content of the document.

6. **Update Document**
   - Method: `PUT`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`
   - Request Body: Raw content of the updated document.

7. **Delete Document**
   - Method: `DELETE`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`

## Configuration and Execution

1. **Install Dependencies:**
   ```bash
   go get -u github.com/gin-gonic/gin
   go get -u golang.org/x/crypto/bcrypt
   go get -u github.com/dgrijalva/jwt-go
   ```

2. **Run the Application:**
   ```bash
   go run main.go
   ```

The application will run on `myserver.local:5000`. Ensure that you have proper hostname resolution or modify the address and port as needed.

## Automated Tests

Yet to be implemented.