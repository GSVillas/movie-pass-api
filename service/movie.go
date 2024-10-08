package service

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/GSVillas/movie-pass-api/client"
	"github.com/GSVillas/movie-pass-api/domain"
	"github.com/GSVillas/movie-pass-api/utils"
	"github.com/google/uuid"
	"github.com/samber/do"
)

type movieService struct {
	i                 *do.Injector
	movieRepository   domain.MovieRepository
	cloudFlareService client.CloudFlareService
}

func NewMovieService(i *do.Injector) (domain.MovieService, error) {
	movieRepository, err := do.Invoke[domain.MovieRepository](i)
	if err != nil {
		return nil, err
	}

	cloudFlareService, err := do.Invoke[client.CloudFlareService](i)
	if err != nil {
		return nil, err
	}

	return &movieService{
		i:                 i,
		movieRepository:   movieRepository,
		cloudFlareService: cloudFlareService,
	}, nil
}

func (m *movieService) GetAllIndicativeRatings(ctx context.Context) ([]*domain.IndicativeRatingResponse, error) {
	indicativeRatings, err := m.movieRepository.GetAllIndicativeRating(ctx)
	if err != nil {
		return nil, fmt.Errorf("error to get all indicative ratings %w", err)
	}

	if indicativeRatings == nil {
		return nil, domain.ErrIndicativeRatingsNotFound
	}

	var indicativeRatingsResponse []*domain.IndicativeRatingResponse
	for _, indicativeRattings := range indicativeRatings {
		indicativeRatingsResponse = append(indicativeRatingsResponse, indicativeRattings.ToIndicativeRatingResponse())
	}

	return indicativeRatingsResponse, nil
}

func (m *movieService) Create(ctx context.Context, payload domain.MoviePayload) (*domain.MovieResponse, error) {
	log := slog.With(
		slog.String("service", "movie"),
		slog.String("func", "create"),
	)

	session, ok := ctx.Value(domain.SessionKey).(*domain.Session)
	if !ok || session == nil {
		return nil, domain.ErrUserNotFoundInContext
	}

	indicativeRating, err := m.movieRepository.GetIndicativeRatingByID(ctx, payload.IndicativeRatingID)
	if err != nil {
		return nil, fmt.Errorf("error to get indicative rating by id %w", err)
	}

	if indicativeRating == nil {
		return nil, domain.ErrIndicativeRatingNotFound
	}

	movie := payload.ToMovie(session.UserID)

	if err := m.movieRepository.Create(ctx, *movie); err != nil {
		return nil, fmt.Errorf("error to create movie %w", err)
	}

	for _, image := range payload.Images {
		imageBytes, err := utils.ConvertImageToBytes(image)
		if err != nil {
			log.Error("error to convert image to bytes", slog.String("error", err.Error()))
			continue
		}

		task := domain.MovieImageUploadTask{
			MovieID: movie.ID,
			Image:   imageBytes,
			UserID:  session.UserID,
		}

		if err := m.movieRepository.AddUploadTaskToQueue(ctx, task); err != nil {
			log.Error(err.Error())
		}
	}

	indicativeRatingResponse := indicativeRating.ToIndicativeRatingResponse()
	movieResponse := movie.ToMovieResponse()
	movieResponse.IndicativeRating = indicativeRatingResponse

	return movieResponse, nil
}

func (m *movieService) GetAllByUserID(ctx context.Context, pagination *domain.Pagination) (*domain.Pagination, error) {
	session, ok := ctx.Value(domain.SessionKey).(*domain.Session)
	if !ok || session == nil {
		return nil, domain.ErrUserNotFoundInContext
	}

	moviesPagination, err := m.movieRepository.GetALlByUserID(ctx, session.UserID, pagination)
	if err != nil {
		return nil, fmt.Errorf("error to get all movies by user id. error: %w", err)
	}

	if moviesPagination == nil {
		return nil, domain.ErrMoviesNotFoundByUserID
	}

	var moviesResponse []*domain.MovieResponse
	for _, movie := range moviesPagination.Rows.([]*domain.Movie) {
		moviesResponse = append(moviesResponse, movie.ToMovieResponse())
	}

	moviesPagination.Rows = moviesResponse

	return moviesPagination, nil
}

func (m *movieService) Update(ctx context.Context, ID uuid.UUID, payload domain.MovieUpdatePayload) (*domain.MovieResponse, error) {
	session, ok := ctx.Value(domain.SessionKey).(*domain.Session)
	if !ok || session == nil {
		return nil, domain.ErrUserNotFoundInContext
	}

	movie, err := m.movieRepository.GetByID(ctx, ID, true)
	if err != nil {
		return nil, fmt.Errorf("error to get movie by id: %w", err)
	}

	if movie == nil {
		return nil, domain.ErrMoviesNotFound
	}

	if movie.UserID != session.UserID {
		return nil, domain.ErrMovieNotBelongUser
	}

	if payload.IndicativeRatingID != nil {
		indicativeRating, err := m.movieRepository.GetIndicativeRatingByID(ctx, *payload.IndicativeRatingID)
		if err != nil {
			return nil, fmt.Errorf("error to get indicative rating: %w", err)
		}

		if indicativeRating == nil {
			return nil, domain.ErrIndicativeRatingNotFound
		}
		movie.IndicativeRating = *indicativeRating
	}

	if payload.Title != nil {
		movie.Title = *payload.Title
	}

	if payload.Duration != nil {
		movie.Duration = *payload.Duration
	}

	if err := m.movieRepository.Update(ctx, *movie); err != nil {
		return nil, fmt.Errorf("error to update movie: %w", err)
	}

	return movie.ToMovieResponse(), nil
}

func (m *movieService) Delete(ctx context.Context, ID uuid.UUID) error {
	log := slog.With(
		slog.String("service", "movie"),
		slog.String("func", "delete"),
	)

	session, ok := ctx.Value(domain.SessionKey).(*domain.Session)
	if !ok || session == nil {
		return domain.ErrUserNotFoundInContext
	}

	movie, err := m.movieRepository.GetByID(ctx, ID, true)
	if err != nil {
		return fmt.Errorf("error to get movies by id error:%w", err)
	}

	if movie == nil {
		return domain.ErrMoviesNotFound
	}

	if movie.UserID != session.UserID {
		return domain.ErrMovieNotBelongUser
	}

	for _, image := range movie.Images {
		task := domain.MovieImageDeleteTask{
			CloudFlareID: image.CloudFlareID,
		}

		if err := m.movieRepository.AddDeleteTaskToQueue(ctx, task); err != nil {
			log.Error(err.Error())
		}
	}

	return nil
}

func (m *movieService) ProcessUploadQueue(ctx context.Context, task domain.MovieImageUploadTask) error {
	filename := fmt.Sprintf("movie_%s_image_%d.jpg", task.MovieID.String(), time.Now().Unix())
	response, err := m.cloudFlareService.UploadImage(task.Image, filename)
	if err != nil {
		return fmt.Errorf("error to upload image to Cloudflare %w", err)
	}

	movieImage := domain.MovieImage{
		ID:           uuid.New(),
		MovieID:      task.MovieID,
		ImageURL:     response.URL,
		CloudFlareID: response.ID,
	}

	if err := m.movieRepository.CreateMovieImage(ctx, movieImage); err != nil {
		return fmt.Errorf("error to save movie image to the database error:%w", err)
	}

	return nil
}

func (m *movieService) ProcessDeleteQueue(ctx context.Context, task domain.MovieImageDeleteTask) error {
	if err := m.cloudFlareService.DeleteImage(task.CloudFlareID); err != nil {
		return fmt.Errorf("error to delete image to Cloudflare %w", err)
	}

	if err := m.movieRepository.DeleteMovieImage(ctx, task.CloudFlareID); err != nil {
		return fmt.Errorf("error to delete movie image to the database error:%w", err)
	}

	return nil
}
