package handler

import (
	"github.com/GSVillas/movie-pass-api/domain"
	"github.com/GSVillas/movie-pass-api/middleware"
	"github.com/labstack/echo/v4"
	"github.com/samber/do"
)

func SetupRoutes(e *echo.Echo, i *do.Injector) {
	setupUserRoutes(e, i)
	setupCinemaRoutes(e, i)
	setupMovieRoutes(e, i)
}

func setupUserRoutes(e *echo.Echo, i *do.Injector) {
	userHandler, err := do.Invoke[domain.UserHandler](i)
	if err != nil {
		panic(err)
	}

	group := e.Group("/v1/users")
	group.POST("", userHandler.Create)
	group.POST("/sign-in", userHandler.SignIn)
}

func setupCinemaRoutes(e *echo.Echo, i *do.Injector) {
	cinemaHandler, err := do.Invoke[domain.CinemaHandler](i)
	if err != nil {
		panic(err)
	}

	group := e.Group("/v1/cinemas", middleware.EnsureAuthenticated(i))
	group.POST("", cinemaHandler.Create)
	group.GET("", cinemaHandler.GetAll)
	group.GET("/:cinemaId", cinemaHandler.GetByID)
	group.DELETE("/:cinemaId", cinemaHandler.Delete)
}

func setupMovieRoutes(e *echo.Echo, i *do.Injector) {
	movieHandler, err := do.Invoke[domain.MovieHandler](i)
	if err != nil {
		panic(err)
	}

	adminGroup := e.Group("/v1/admin/movies")
	adminGroup.POST("", movieHandler.Create, middleware.EnsureAuthenticated(i))
	adminGroup.GET("/indicative-rating", movieHandler.GetAllIndicativeRating)
	adminGroup.GET("", movieHandler.GetAllByUserID, middleware.EnsureAuthenticated(i), middleware.EnsureAuthenticated(i))
}
