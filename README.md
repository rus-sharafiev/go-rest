# Go REST Service

Template for creating a REST service using [Go](https://go.dev/)  
The app requires [PostgreSQL](https://www.postgresql.org/download/) (uses the pgx driver) and [Redis](https://redis.io/docs/install/install-redis/)  
This project is inspired by the NestJS framework

Currently has been implemented:
- SPA handler
- Signup
- Signin
- Email verification
- reCAPTCHA validation
- User access control (JWT)
- Refresh tokens ([OAuth 2.0](https://oauth.net/2/refresh-tokens/))
- Exceptions that return JSON
- Localization (En, Ru)
- FormData interceptor (requires [custom FormData converter](https://github.com/rus-sharafiev/fetch-api))

The project is under active development
