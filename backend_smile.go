package nigonigo

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type SmileSession struct {
	VideoData *VideoData
}

func (s *SmileSession) FileExtension() string {
	return s.VideoData.SmileFileExtension()
}

func CreateSmileSessionByVideoData(client *http.Client, data *VideoData) (*SmileSession, error) {
	return &SmileSession{VideoData: data}, nil
}

func (s *SmileSession) Download(ctx context.Context, client *http.Client, w io.Writer, optionalStreamID string) error {
	if !s.VideoData.IsSmile() {
		return errors.New("SMILE is not available")
	}

	if !strings.HasPrefix(s.VideoData.Video.Smile.URL, "http") {
		return fmt.Errorf("unsported protocol : %v", s.VideoData.Video.Smile.URL)
	}
	return NewHTTPDownloader(client).Download(ctx, s.VideoData.Video.Smile.URL, w)
}
