package auth

import "net/http"

func authHandler() *http.ServeMux {
	authRouter := http.NewServeMux()

	authRouter.Handle("/api/auth/signup", SignUp)
	authRouter.Handle("/api/auth/signup/verify", VerifySignup)
	authRouter.Handle("/api/auth/login", LogIn)
	authRouter.Handle("/api/auth/logout", LogOut)
	authRouter.Handle("/api/auth/user", GetAuthUser)
	authRouter.Handle("/api/auth/refresh", Refresh)
	authRouter.Handle("/api/auth/password/reset", ResetPassword)
	authRouter.Handle("/api/auth/password/update", UpdatePassword)

	return authRouter
}

var Controller = authHandler()
