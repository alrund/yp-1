package grpcserver

import (
	"context"
	"errors"
	"net/http"

	"github.com/alrund/yp-1/internal/app/storage"
	pb "github.com/alrund/yp-1/internal/proto"
)

// AddBatch adds multiple URLs at once for shortening.
func (s *Server) AddBatch(ctx context.Context, in *pb.AddBatchRequest) (*pb.AddBatchResponse, error) {
	var response pb.AddBatchResponse
	response.Code = http.StatusCreated

	contextUserID := ctx.Value(UserIDContextKey)
	userID, ok := contextUserID.(string)
	if !ok {
		response.Code = http.StatusInternalServerError
		response.Message = http.StatusText(http.StatusInternalServerError)
		return &response, nil
	}

	URLs, URL2Row := getURL2Row(in.Urls)
	tokens, err := s.us.AddBatch(userID, URLs)
	if err != nil {
		if !errors.Is(err, storage.ErrURLAlreadyExists) {
			response.Code = http.StatusInternalServerError
			response.Message = err.Error()
			return &response, nil
		}
		response.Code = http.StatusConflict
	}

	for URL, token := range tokens {
		row, ok := URL2Row[URL]
		if !ok {
			response.Code = http.StatusInternalServerError
			response.Message = "URL not found in URL2Row map"
			return &response, nil
		}
		response.ShortUrls = append(response.ShortUrls, &pb.AddBatchResponse_Url{
			CorrelationId: row.CorrelationId,
			ShortUrl:      s.us.GetBaseURL() + token.Value,
		})
	}

	return &response, nil
}

func getURL2Row(rows []*pb.AddBatchRequest_Url) ([]string, map[string]*pb.AddBatchRequest_Url) {
	URLs := make([]string, 0)
	URL2Row := map[string]*pb.AddBatchRequest_Url{}

	for _, row := range rows {
		URLs = append(URLs, row.OriginalUrl)
		URL2Row[row.OriginalUrl] = row
	}

	return URLs, URL2Row
}
