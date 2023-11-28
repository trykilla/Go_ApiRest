# Go REST API with Gin and JWT

This project implements a Go-based REST API for user management and document handling using the Gin framework. It includes user authentication, user creation, document retrieval, creation, update, and deletion.

## Project Structure

- **main.go**: Main logic for route configuration and HTTP request handling.
- **cmd/APIRest/docs/**: Directory storing user-associated documents.
- **cmd/APIRest/users.json**: JSON file storing user credentials.
- **cmd/APIRest/run.sh**: Bash script for running the application.

## Dependencies

- [Gin](https://github.com/gin-gonic/gin): Framework for building APIs in Go.
- [bcrypt](https://golang.org/x/crypto/bcrypt): Library for secure password hashing.
- [jwt-go](https://github.com/dgrijalva/jwt-go): JSON Web Tokens for authentication.

## Endpoints

1. **Get API Version**
   - Method: `GET`
   - Path: `/version`
   Returns the API version.

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
     Post a new user with the given username and password. The password will be hashed and stored in the `users.json` file.
     Returns a JWT token for the new user.

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
      Returns a JWT token for the existing user.

4. **Get Documents Associated with a User**
   - Method: `GET`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`
   Returns the raw content of the document with the given ID associated with the given user.

5. **Add New Document**
   - Method: `POST`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`
   - Request Body:
   ```json
     {
       "doc_content": "This is the content of the document."
     }
     ```
   Creates a new document with the given ID associated with the given user.

6. **Update Document**
   - Method: `PUT`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`
   - Request Body:
   ```json
     {
       "doc_content": "This is the updated content of the document."
     }
     ```
   Updates the document with the given ID associated with the given user.

7. **Delete Document**
   - Method: `DELETE`
   - Path: `/:username/:doc_id`
   - Authorization Header: `Authorization: token <token>`
   Deletes the document with the given ID associated with the given user.

## Configuration and Execution

**How to start the API?**

The script must be executed from the root directory of the project, in this case:  `P3/`

```bash
$ ./run.sh
```

The application will run on `myserver.local:5000`. Ensure proper hostname resolution or modify the address and port as needed. YOU MUST CHANGE `/etc/hosts` and change localhost to myserver.local: "127.0.0.1 myserver.local"

## Automated Tests

Testing is yet to be implemented.