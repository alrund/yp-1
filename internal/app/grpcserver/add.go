package grpcserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"

	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
)

type JSONRequest struct {
	URL string `json:"url"`
}

type JSONResponse struct {
	Result string `json:"result"`
}

// Add adds a URL string to shorten.
func (s *Server) Add(ctx context.Context, in *pb.AddRequest) (*pb.AddResponse, error) {
	var response pb.AddResponse
	response.Code = http.StatusCreated

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	token, err := s.us.Add(userID, in.Url)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			return &response, nil
		}
		response.Code = http.StatusConflict
	}

	response.ShortUrl = s.us.GetBaseURL() + token.Value

	return &response, nil
}

// AddJSON adds a URL string to shorten as a JSON object.
func (s *Server) AddJSON(ctx context.Context, in *pb.AddJSONRequest) (*pb.AddJSONResponse, error) {
	var response pb.AddJSONResponse
	response.Code = http.StatusCreated

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	jsonRequest := JSONRequest{}
	err := json.Unmarshal([]byte(in.UrlJson), &jsonRequest)
	if err != nil {
		response.Code = http.StatusBadRequest
		response.Message = err.Error()
		return &response, err
	}

	token, err := s.us.Add(userID, jsonRequest.URL)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			return &response, err
		}
		response.Code = http.StatusConflict
	}

	jsonResponse := JSONResponse{Result: s.us.GetBaseURL() + token.Value}
	result, err := json.Marshal(jsonResponse)
	if err != nil {
		response.Code = http.StatusInternalServerError
		response.Message = err.Error()
		return &response, err
	}

	response.ShortUrlJson = string(result)

	return &response, nil
}
