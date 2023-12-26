###Movie App Backend Server
This is a simple backend server for a movie app written in Go, utilizing a PostgreSQL database. Docker is used for deploying the PostgreSQL database, making it easy to set up and run the entire system, using the Chi router for handling HTTP requests.

##Features
RESTful API for managing movies
PostgreSQL database for data storage
Dockerized PostgreSQL for easy deployment

###API Endpoints
GET /
Home route.

POST /authenticate
Endpoint for user authentication.

GET /allmovies
Get a list of all movies.

GET /refresh
Refresh authentication token.

GET /logout
Logout the user.

GET /movies/{id}
Get details of a specific movie.

GET /allgenres
Get a list of all movie genres.

GET /movies/genres/{id}
Get all movies of a specific genre.

Admin Routes:
GET /admin/movies
Get a movie catalog (admin access required).

GET /admin/movie/{id}
Get details of a specific movie for editing (admin access required).

PUT /admin/movies/0
Insert a new movie (admin access required).

PATCH /admin/movies/{id}
Update details of a specific movie (admin access required).

DELETE /admin/movies/{id}
Delete a specific movie (admin access required).

##Prerequisites
Before you begin, ensure you have the following installed:

Go (version 1.20)
Docker
Docker Compose
