package main

import (
	"log"
	"net/http"

	"agnos-backend-assignment/internal/client"
	"agnos-backend-assignment/internal/config"
	"agnos-backend-assignment/internal/database"
	"agnos-backend-assignment/internal/handler"
	"agnos-backend-assignment/internal/middleware"
	"agnos-backend-assignment/internal/migration"
	"agnos-backend-assignment/internal/repository"
	"agnos-backend-assignment/internal/service"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	cfg := config.Load()

	db, err := database.Connect(cfg)
	if err != nil {
		log.Fatal(err)
	}

	log.Println(
		"database connected successfully",
	)

	if err := migration.Run(
		db,
	); err != nil {
		log.Fatal(err)
	}

	log.Println(
		"database migration completed",
	)

	if err := migration.Seed(
		db,
		cfg,
	); err != nil {
		log.Fatal(err)
	}

	log.Println(
		"database seed completed",
	)

	/*
		Repositories
	*/

	hospitalRepo :=
		repository.NewHospitalRepository(
			db,
		)

	staffRepo :=
		repository.NewStaffRepository(
			db,
		)

	patientRepo :=
		repository.NewPatientRepository(
			db,
		)

	/*
		External clients
	*/

	hospitalClient := client.NewHospitalClient(
		5 * time.Second,
	)

	/*
		Services
	*/

	jwtService :=
		service.NewJWTService(
			cfg.JWTSecret,
			24*time.Hour,
		)

	staffService :=
		service.NewStaffService(
			staffRepo,
			hospitalRepo,
		)

	authService :=
		service.NewAuthService(
			staffRepo,
			hospitalRepo,
			jwtService,
		)

	patientService :=
		service.NewPatientService(
			patientRepo,
			hospitalRepo,
			hospitalClient,
		)

	/*
		Handlers
	*/

	staffHandler :=
		handler.NewStaffHandler(
			staffService,
		)

	authHandler :=
		handler.NewAuthHandler(
			authService,
		)

	patientHandler :=
		handler.NewPatientHandler(
			patientService,
		)

	/*
		Router
	*/

	router := gin.Default()

	router.GET(
		"/health",
		func(c *gin.Context) {
			c.JSON(
				http.StatusOK,
				gin.H{
					"status":  "ok",
					"message": "Agnos backend is running",
				},
			)
		},
	)

	/*
		Public routes
	*/

	router.POST(
		"/staff/create",
		staffHandler.Create,
	)

	router.POST(
		"/staff/login",
		authHandler.Login,
	)

	/*
		Protected routes
	*/

	protected := router.Group("/")

	protected.Use(
		middleware.Auth(
			jwtService,
		),
	)

	protected.GET(
		"/staff/me",
		handler.Profile,
	)

	protected.GET(
		"/patient/search",
		patientHandler.Search,
	)

	/*
		Start server
	*/

	log.Printf(
		"server is running on port %s",
		cfg.AppPort,
	)

	if err := router.Run(
		":" + cfg.AppPort,
	); err != nil {
		log.Fatal(err)
	}
}
