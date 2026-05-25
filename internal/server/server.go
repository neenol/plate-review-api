package server

import (
	"context"
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	chiMiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/cors"

	"github.com/nglambertjr/plate-review-api/internal/config"
	"github.com/nglambertjr/plate-review-api/internal/controllers"
	"github.com/nglambertjr/plate-review-api/internal/middleware"
	clerkPlatform "github.com/nglambertjr/plate-review-api/internal/platform/clerk"
	"github.com/nglambertjr/plate-review-api/internal/repositories"
	"github.com/nglambertjr/plate-review-api/internal/services"
)

// New constructs the full application handler.  It does not start a listener.
func New(ctx context.Context, cfg *config.Config) (http.Handler, error) {
	// --- Database pool ---
	pool, err := pgxpool.New(ctx, cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("server: open db pool: %w", err)
	}
	if err = pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("server: ping db: %w", err)
	}

	// --- Platform clients ---
	clerkClient, err := clerkPlatform.New(cfg.ClerkJWTKey)
	if err != nil {
		return nil, fmt.Errorf("server: init clerk client: %w", err)
	}

	// --- Repositories ---
	userRepo := repositories.NewUserRepository(pool)
	plateRepo := repositories.NewPlateRepository(pool)
	reviewRepo := repositories.NewReviewRepository(pool)
	replyRepo := repositories.NewReplyRepository(pool)
	voteRepo := repositories.NewVoteRepository(pool)

	// --- Services ---
	userSvc := services.NewUserService(userRepo)
	plateSvc := services.NewPlateService(plateRepo, cfg.PlatePepper)
	reviewSvc := services.NewReviewService(reviewRepo, userRepo, plateRepo, cfg.PlatePepper)
	replySvc := services.NewReplyService(replyRepo, reviewRepo, userRepo)
	voteSvc := services.NewVoteService(voteRepo, userRepo, reviewRepo, replyRepo)

	// --- Controllers ---
	userCtrl := controllers.NewUserController(userSvc)
	plateCtrl := controllers.NewPlateController(plateSvc)
	reviewCtrl := controllers.NewReviewController(reviewSvc)
	replyCtrl := controllers.NewReplyController(replySvc)
	voteCtrl := controllers.NewVoteController(voteSvc)

	// --- Router ---
	r := chi.NewRouter()

	c := cors.New(cors.Options{
		AllowedOrigins:   []string{cfg.CORSOrigin},
		AllowedMethods:   []string{"GET", "POST", "PATCH", "DELETE", "OPTIONS"},
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowCredentials: true,
	})

	r.Use(c.Handler)
	r.Use(middleware.RequestID)
	r.Use(middleware.Logger)
	r.Use(chiMiddleware.Recoverer)

	requireAuth := middleware.RequireAuth(clerkClient)

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"status":"ok"}`))
	})

	// Users
	r.With(requireAuth).Post("/users", userCtrl.Create)
	r.With(requireAuth).Get("/users/me", userCtrl.Me)
	r.With(requireAuth).Patch("/users/me", userCtrl.UpdateMe)

	// Plates
	r.Get("/plates/{state}/{plate}", plateCtrl.Get)

	// Reviews
	r.Get("/plates/{state}/{plate}/reviews", reviewCtrl.List)
	r.With(requireAuth).Post("/plates/{state}/{plate}/reviews", reviewCtrl.Create)
	r.With(requireAuth).Delete("/reviews/{id}", reviewCtrl.Delete)

	// Replies on reviews
	r.Get("/reviews/{id}/replies", replyCtrl.ListByReview)
	r.With(requireAuth).Post("/reviews/{id}/replies", replyCtrl.CreateOnReview)

	// Replies on replies
	r.Get("/replies/{id}/replies", replyCtrl.ListByParent)
	r.With(requireAuth).Post("/replies/{id}/replies", replyCtrl.CreateOnReply)
	r.With(requireAuth).Delete("/replies/{id}", replyCtrl.Delete)

	// Votes
	r.With(requireAuth).Post("/reviews/{id}/votes", voteCtrl.CastReviewVote)
	r.With(requireAuth).Delete("/reviews/{id}/votes", voteCtrl.RetractReviewVote)
	r.With(requireAuth).Post("/replies/{id}/votes", voteCtrl.CastReplyVote)
	r.With(requireAuth).Delete("/replies/{id}/votes", voteCtrl.RetractReplyVote)

	return r, nil
}
